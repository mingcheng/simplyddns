/*!*
 * Copyright (c) 2026 Ming Lyu, aka mingcheng
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: job.go
 * Author: mingcheng <mingcheng@apache.org>
 * File Created: 2026-05-12 12:08:01
 *
 * Modified By: mingcheng <mingcheng@apache.org>
 * Last Modified: 2026-05-12 12:23:12
 */

package simplyddns

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"
)

// SourceConfig describes how a source function obtains the public IP address.
type SourceConfig struct {
	Interval uint `yaml:"interval,omitempty"`
	// Type defaults to auto; explicit names never include automatic sources.
	Type    string `yaml:"type,omitempty"`
	Path    string `yaml:"path"`
	Content string `yaml:"content"`
}

// TargetConfig describes how a target function updates DNS records.
type TargetConfig struct {
	Type    string   `yaml:"type,omitempty"`
	Key     string   `yaml:"key,omitempty"`
	Token   string   `yaml:"token"`
	Proxy   string   `yaml:"proxy"`
	Domains []string `yaml:"domains,omitempty"`
}

// WebHook describes the webhook triggered when the address changes.
type WebHook struct {
	Url      string `yaml:"url" mapstructure:"url"`
	Token    string `yaml:"token" mapstructure:"token"`
	UserName string `yaml:"username" mapstructure:"username"`
	Password string `yaml:"password" mapstructure:"password"`
}

// JobConfig describes a single DDNS job, composed of a source, a target and
// an optional webhook.
type JobConfig struct {
	WebHook WebHook      `yaml:"webhook" mapstructure:"webhook"`
	Source  SourceConfig `yaml:"source,omitempty" mapstructure:"source"`
	Target  TargetConfig `yaml:"target,omitempty" mapstructure:"target"`
}

// Job represents a running DDNS job created from a JobConfig.
type Job struct {
	Config     *JobConfig
	SourceFunc []SourceFunc
	TargetFunc TargetFunc
	stopCtx    context.Context
	cancel     context.CancelFunc
	startOnce  sync.Once
	timeout    time.Duration
	lastIP     net.IP
}

// RunWebhook triggers the configured webhook after a successful address update.
func (j *Job) RunWebhook(ctx context.Context, addr string, domains []string) error {
	if j.Config == nil {
		return fmt.Errorf("job configuration is nil")
	}
	domainList := strings.Join(domains, ",")
	body, err := json.Marshal(struct {
		Address string    `json:"address"`
		Domains string    `json:"domains"`
		Now     time.Time `json:"now"`
	}{addr, domainList, time.Now()})
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, j.Config.WebHook.Url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", DefaultUserAgent)
	request.Header.Set("Address", addr)
	request.Header.Set("Domains", domainList)
	if token := j.Config.WebHook.Token; token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	if username := j.Config.WebHook.UserName; username != "" {
		request.SetBasicAuth(username, j.Config.WebHook.Password)
	}

	client := &http.Client{Timeout: DefaultTimeout}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %s", resp.Status)
	}
	_, err = io.Copy(io.Discard, resp.Body)
	return err
}

// Start runs the job loop until Stop is called or the context is done. Each
// tick fetches the address from the source, optionally validates it against
// the configured domains and applies the target function.
func (j *Job) Start(ctx context.Context) {
	j.startOnce.Do(func() {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		stop := context.AfterFunc(j.stopCtx, cancel)
		defer stop()
		ticker := time.NewTicker(time.Duration(j.Config.Source.Interval) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := j.run(ctx); err != nil && ctx.Err() == nil {
					log.Warnf("DDNS job failed: %v", err)
				}
			case <-j.stopCtx.Done():
				return
			case <-ctx.Done():
				return
			}
		}
	})
}

func (j *Job) run(ctx context.Context) error {
	timeout := j.timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	addr, err := j.Source(ctx, &j.Config.Source)
	if err != nil {
		return err
	}
	if j.lastIP.Equal(*addr) {
		log.Debugf("ignore the cached address %s", addr)
		return nil
	}
	domains := j.Config.Target.Domains
	if len(domains) > 0 {
		if err := ValidateRecordsContext(ctx, domains, addr); err == nil {
			log.Debugf("DNS records already point to %s", addr)
			return nil
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := j.TargetFunc(ctx, addr, &j.Config.Target); err != nil {
		return fmt.Errorf("update target: %w", err)
	}
	j.lastIP = slices.Clone(*addr)
	log.Infof("target function executed successfully")
	if j.Config.WebHook.Url != "" {
		if err := j.RunWebhook(ctx, addr.String(), domains); err != nil {
			return fmt.Errorf("webhook failed after DNS update: %w", err)
		}
	}
	return nil
}

// Stop signals the job loop to exit.
func (j *Job) Stop() {
	log.Debug("stopping job")
	j.cancel()
}

// Source executes only the configured sources. Automatic sources elect an
// address by majority; explicitly listed sources must agree on the address.
func (j *Job) Source(ctx context.Context, config *SourceConfig) (*net.IP, error) {
	if len(j.SourceFunc) == 0 {
		return nil, fmt.Errorf("source functions is empty")
	}

	var (
		errTimes int
		lastAddr net.IP
	)

	for _, fn := range j.SourceFunc {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if fn == nil {
			return nil, fmt.Errorf("source function is nil")
		}
		addr, err := fn(ctx, config)
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if err == nil && (addr == nil || addr.To16() == nil) {
			err = fmt.Errorf("source returned an invalid IP address")
		}
		if err != nil {
			log.Errorf("source function error (%d): %v", errTimes, err)
			if len(j.SourceFunc) == 1 {
				return nil, err
			}
			errTimes++
			continue
		}

		if lastAddr != nil && !addr.Equal(lastAddr) {
			return nil, fmt.Errorf("fetched addresses do not match: %v vs %v", lastAddr, addr)
		}
		lastAddr = slices.Clone(*addr)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// allow some tolerance when multiple source functions are configured
	if len(j.SourceFunc) > 3 && errTimes >= len(j.SourceFunc)/2 {
		return nil, fmt.Errorf("too many source function errors (%d), result is unreliable", errTimes)
	}

	if lastAddr == nil {
		return nil, fmt.Errorf("no source function returned a valid address")
	}

	return &lastAddr, nil
}

// NewJob creates a new Job from the given JobConfig.
func NewJob(config JobConfig) (*Job, error) {
	if strings.TrimSpace(config.Source.Type) == "" {
		config.Source.Type = DefaultSourceType
	}
	if err := ValidateConfig(&config); err != nil {
		return nil, err
	}

	// set default interval if not specified or too small
	if config.Source.Interval == 0 {
		config.Source.Interval = DefaultInterval
	} else if config.Source.Interval < MinInterval {
		config.Source.Interval = MinInterval
	}

	// the source type may contain multiple comma-separated names
	types := strings.Split(config.Source.Type, ",")
	sourceFuncs := make([]SourceFunc, 0, len(types))
	for _, v := range types {
		fn, err := SourceFuncByName(v)
		if err != nil {
			return nil, err
		}
		sourceFuncs = append(sourceFuncs, fn)
	}

	fnTarget, err := TargetFuncByName(config.Target.Type)
	if err != nil {
		return nil, err
	}
	config.Target.Domains = slices.Clone(config.Target.Domains)
	for i, domain := range config.Target.Domains {
		parsed, err := ParseDomain(domain)
		if err != nil {
			return nil, err
		}
		config.Target.Domains[i] = parsed.Hostname()
	}

	stopCtx, cancel := context.WithCancel(context.Background())
	return &Job{
		SourceFunc: sourceFuncs,
		TargetFunc: fnTarget,
		Config:     &config,
		stopCtx:    stopCtx,
		cancel:     cancel,
		timeout:    DefaultTimeout,
	}, nil
}

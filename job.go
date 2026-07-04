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
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// SourceConfig describes how a source function obtains the public IP address.
type SourceConfig struct {
	Interval uint   `yaml:"interval,omitempty"`
	Type     string `yaml:"type,omitempty"`
	Path     string `yaml:"path"`
	Content  string `yaml:"content"`
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
	ticker     *time.Ticker
	done       chan bool
	lastIP     net.IP
}

// RunWebhook triggers the configured webhook after a successful address update.
func (j *Job) RunWebhook(ctx context.Context, addr string, domains []string) error {
	client := resty.New()

	request := client.R().
		SetContext(ctx).
		SetHeader("Address", addr).
		SetHeader("Domains", strings.Join(domains, ",")).
		SetBody(map[string]interface{}{
			"address": addr,
			"domains": strings.Join(domains, ","),
			"now":     time.Now(),
		})

	if token := j.Config.WebHook.Token; token != "" {
		request.SetAuthToken(token)
	}

	if username := j.Config.WebHook.UserName; username != "" {
		request.SetBasicAuth(username, j.Config.WebHook.Password)
	}

	resp, err := request.Post(j.Config.WebHook.Url)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("webhook returned non-OK status: %s", resp.Status())
	}

	return nil
}

// Start runs the job loop until Stop is called or the context is done. Each
// tick fetches the address from the source, optionally validates it against
// the configured domains and applies the target function.
func (j *Job) Start(ctx context.Context) {
	for {
		select {
		case <-j.ticker.C:
			config := j.Config

			// validate the job configure
			if err := ValidateConfig(config); err != nil {
				log.Errorf("validate job configure failed, %v", err)
				continue
			}

			// run source function
			addr, err := j.Source(ctx, &config.Source)
			if err != nil || addr == nil || addr.String() == "" {
				if err != nil {
					log.Error(err)
				}
				continue
			}

			log.Debugf("got address from source %q, value is %s", config.Source.Type, addr.String())

			// ignore the same ip address
			if j.lastIP != nil && j.lastIP.Equal(*addr) {
				log.Debugf("ignore the cached address %s", addr.String())
				continue
			}

			domains := config.Target.Domains
			if len(domains) > 0 {
				// if ValidateRecords returns nil every domain already points
				// to the current address, so there is nothing to do.
				if err := ValidateRecords(domains, addr); err == nil {
					log.Debugf("dns records already point to %s, skip", addr.String())
					continue
				}
			}

			// run the target function
			if err := j.TargetFunc(ctx, addr, &config.Target); err != nil {
				log.Warn(err)
				continue
			}
			log.Infof("target function executed successfully")

			// cache the last ip address
			j.lastIP = *addr

			// trigger the webhook if configured
			if config.WebHook.Url != "" {
				log.Tracef("triggering webhook %s", config.WebHook.Url)
				if err := j.RunWebhook(ctx, addr.String(), domains); err != nil {
					log.Warnf("webhook failed: %v", err)
				} else {
					log.Infof("webhook %s finished", config.WebHook.Url)
				}
			}
		case <-j.done:
			j.ticker.Stop()
			return
		case <-ctx.Done():
			j.ticker.Stop()
			return
		}
	}
}

// Stop signals the job loop to exit.
func (j *Job) Stop() {
	log.Debug("stopping job")
	j.done <- true
}

// Source executes every registered source function for this job and returns
// the resolved address. When multiple source functions are configured, they
// must all agree on the same address; otherwise an error is returned.
func (j *Job) Source(ctx context.Context, config *SourceConfig) (*net.IP, error) {
	if len(j.SourceFunc) == 0 {
		return nil, fmt.Errorf("source functions is empty")
	}

	var (
		errTimes int
		lastAddr net.IP
	)

	for _, fn := range j.SourceFunc {
		addr, err := fn(ctx, config)
		if err != nil {
			log.Errorf("source function error (%d): %v", errTimes, err)
			errTimes++
			continue
		}

		if addr == nil {
			continue
		}

		if lastAddr != nil && !addr.Equal(lastAddr) {
			return nil, fmt.Errorf("fetched addresses do not match: %v vs %v", lastAddr, addr)
		}
		lastAddr = *addr
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
	if config.Source.Type == "" || config.Target.Type == "" {
		return nil, fmt.Errorf("source or target type can not be empty")
	}

	// set default interval if not specified or too small
	if config.Source.Interval <= 0 {
		config.Source.Interval = DefaultInterval
	} else if config.Source.Interval < MinInterval {
		config.Source.Interval = MinInterval
	}

	// the source type may contain multiple comma-separated names
	types := strings.Split(config.Source.Type, ",")
	if len(types) == 0 {
		return nil, fmt.Errorf("source type %q is empty", config.Source.Type)
	}

	sourceFuncs := make([]SourceFunc, 0, len(types))
	for _, v := range types {
		fn, err := SourceFuncByName(strings.ToLower(strings.TrimSpace(v)))
		if err != nil {
			return nil, err
		}
		sourceFuncs = append(sourceFuncs, fn)
	}

	fnTarget, err := TargetFuncByName(config.Target.Type)
	if err != nil {
		return nil, err
	}

	return &Job{
		SourceFunc: sourceFuncs,
		TargetFunc: fnTarget,
		Config:     &config,
		ticker:     time.NewTicker(time.Second * time.Duration(config.Source.Interval)),
		done:       make(chan bool, 1),
	}, nil
}

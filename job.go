/*!*
 * Copyright (c) 2022-2025 Hangzhou Guanwaii Technology Co,.Ltd.
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: job.go
 * Author: mingcheng (mingcheng@apache.org)
 * File Created: 2022-07-22 23:37:43
 *
 * Modified By: mingcheng (mingcheng@apache.org)
 * Last Modified: 2025-02-28 10:46:31
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

// SourceFunc to define the source function
type SourceConfig struct {
	Interval uint   `yaml:"interval,omitempty"`
	Type     string `yaml:"type,omitempty"`
	Path     string `yaml:"path"`
	Content  string `yaml:"content"`
}

// TargetFunc to define the target function
type TargetConfig struct {
	Type    string   `yaml:"type,omitempty"`
	Key     string   `yaml:"key,omitempty"`
	Token   string   `yaml:"token"`
	Proxy   string   `yaml:"proxy"`
	Domains []string `yaml:"domains,omitempty"`
}

// WebHook to define the webhook
type WebHook struct {
	Url      string `yaml:"url" mapstructure:"url"`
	Token    string `yaml:"token" mapstructure:"token"`
	UserName string `yaml:"username" mapstructure:"username"`
	Password string `yaml:"password" mapstructure:"password"`
}

// JobConfig to define the job configure
type JobConfig struct {
	WebHook WebHook      `yaml:"webhook" mapstructure:"webhook"`
	Source  SourceConfig `yaml:"source,omitempty" mapstructure:"source"`
	Target  TargetConfig `yaml:"target,omitempty" mapstructure:"target"`
}

// Job to define the job
type Job struct {
	Config     *JobConfig
	SourceFunc []SourceFunc
	TargetFunc TargetFunc
	ticker     *time.Ticker
	done       chan bool
	lastIP     net.IP
}

// RunWebhook to run the webhook when ip address has updated
func (j *Job) RunWebhook(ctx context.Context, addr string, domains []string) (err error) {
	client := resty.New()

	request := client.R().
		SetContext(ctx).
		SetHeader("Address", addr).
		SetHeader("Domains", strings.Join(domains, ",")).
		SetBody(map[string]interface{}{
			"address": addr,
			"domains": strings.Join(domains, ","),
			"now":     time.Now(),
		}).
		SetError(&err)

	if token := j.Config.WebHook.Token; token != "" {
		request.SetAuthToken(token)
	}

	if username := j.Config.WebHook.UserName; username != "" {
		request.SetBasicAuth(username, j.Config.WebHook.Password)
	}

	var resp *resty.Response
	resp, err = request.Post(j.Config.WebHook.Url)

	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("%v", resp.Status())
	}

	return err
}

// Start to start a job
func (j *Job) Start(ctx context.Context) {
	go func() {
		var (
			err  error
			addr *net.IP
			job  = j
		)

		for {
			select {
			case <-job.ticker.C:
				var config = job.Config

				// check configure
				if err = ValidateConfig(config); err != nil {
					log.Errorf("validate job configure is fail, %v", err)
					continue
				}

				// run source function
				if addr, err = job.Source(ctx, &config.Source); err != nil || addr == nil || addr.String() == "" {
					log.Error(err)
					continue
				}

				// markup the source func result
				log.Debugf("get address from source fun %s, value is %s", config.Source.Type, addr.String())

				// ignore the same ip address
				if j.lastIP != nil && j.lastIP.Equal(*addr) {
					log.Warnf("ignore the cached address %s", addr.String())
					continue
				}

				domains := config.Target.Domains
				if len(domains) > 0 {
					if err = ValidateRecords(domains, addr); err == nil {
						log.Errorf("valdate dns record without error, maybe already setted %s", addr.String())
						continue
					}
				}

				// run the target func
				err = job.TargetFunc(ctx, addr, &job.Config.Target)
				if err != nil {
					log.Warn(err)
					continue
				}
				log.Infof("run target function is successful, please check")

				// cache the last ip address
				job.lastIP = *addr

				// trigger the webhook if configured
				if config.WebHook.Url != "" {
					log.Tracef("the webhook url is %s", config.WebHook.Url)
					if err = job.RunWebhook(ctx, addr.String(), domains); err != nil {
						log.Warnf("run webhook with error %s", err.Error())
					} else {
						log.Infof("run webhook %s is finished", config.WebHook.Url)
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
	}()
}

// Stop to stop a job
func (j *Job) Stop() {
	log.Debug("stopping job")
	j.done <- true
}

// Source to execute multi-source function
func (j Job) Source(ctx context.Context, config *SourceConfig) (*net.IP, error) {
	if len(j.SourceFunc) == 0 {
		return nil, fmt.Errorf("source functions is empty")
	}

	var (
		err      error
		errTimes int
		lastAddr net.IP
	)

	for _, v := range j.SourceFunc {
		var addr *net.IP
		addr, err = v(ctx, config)
		if err != nil {
			log.Error(err, errTimes)
			errTimes = errTimes + 1
		}

		if addr != nil {
			if lastAddr != nil && !addr.Equal(lastAddr) {
				return nil, fmt.Errorf("fetch address is not the same, %v vs %v", lastAddr, addr)
			}

			lastAddr = *addr
		}
	}

	if errTimes > 0 && len(sourceFuncs) > 3 && errTimes >= len(j.SourceFunc)/2 {
		return nil, fmt.Errorf("max error times reached(%d), so the result is not right", errTimes)
	}

	return &lastAddr, nil
}

// NewJob for instance a new ddns job
func NewJob(config JobConfig) (job *Job, err error) {
	// check the configure
	if config.Source.Type == "" || config.Target.Type == "" {
		err = fmt.Errorf("source or target type can not be empty")
		return
	}

	// Set default interval if not specified or too small
	if config.Source.Interval <= 0 {
		config.Source.Interval = DefaultInterval
	} else if config.Source.Interval < MinInterval {
		config.Source.Interval = MinInterval
	}

	// split fn types as array
	types := strings.Split(config.Source.Type, ",")
	if len(types) == 0 {
		err = fmt.Errorf("load source %s is empty", config.Source.Type)
		return
	}

	// notice: the source functions is an array
	var sourceFuncs []SourceFunc

	for _, v := range types {
		var fn SourceFunc
		fn, err = SourceFuncByName(strings.ToLower(strings.TrimSpace(v)))
		if err != nil {
			return
		}

		// add func to source functions
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
		done:       make(chan bool),
	}, nil
}

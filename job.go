/**
 * File: job.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Friday, December 25th 2020, 10:45:54 pm
 * Last Modified: Monday, December 28th 2020, 9:31:27 am
 *
 * http://www.opensource.org/licenses/MIT
 */

package simplyddns

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

type SourceConfig struct {
	Interval uint   `yaml:"interval,omitempty"`
	Type     string `yaml:"type,omitempty"`
	Path     string `yaml:"path"`
	Content  string `yaml:"content"`
}

type TargetConfig struct {
	Type    string   `yaml:"type,omitempty"`
	Key     string   `yaml:"key,omitempty"`
	Token   string   `yaml:"token"`
	Proxy   string   `yaml:"proxy"`
	Domains []string `yaml:"domains,omitempty"`
}

type JobConfig struct {
	WebHook string       `yaml:"webhook" mapstructure:"webhook"`
	Source  SourceConfig `yaml:"source,omitempty" mapstructure:"source"`
	Target  TargetConfig `yaml:"target,omitempty" mapstructure:"target"`
}

type Job struct {
	Config     *JobConfig
	SourceFunc func(context.Context, *SourceConfig) (*net.IP, error)
	TargetFunc func(context.Context, *net.IP, *TargetConfig) error
	ticker     *time.Ticker
	done       chan bool
	lastIP     *net.IP
}

// RunWebhook to run the webhook when ip address has updated
func (j *Job) RunWebhook(ctx context.Context, ip *net.IP, e error, domains []string) error {
	if j.Config.WebHook == "" {
		return fmt.Errorf("webhook address is nil")
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	webhookUrl := j.Config.WebHook

	req, err := http.NewRequest("GET", webhookUrl, nil)
	if err != nil {
		return err
	}

	req.Header.Add("DDNS-New-Address", ip.String())
	req.Header.Add("DDNS-Domains", strings.Join(domains, ","))
	if e != nil {
		req.Header.Add("DDNS-Error", e.Error())
	}
	req.WithContext(ctx)

	if resp, err := client.Do(req); err != nil {
		log.Debug(err)
		return err
	} else {
		log.Infof("run webhook %s is finished, status code is %v", webhookUrl, resp.StatusCode)
	}

	return nil
}

// Start to start a job
func (j *Job) Start(ctx context.Context) {
	var (
		err    error
		addr   *net.IP
		config = j.Config
	)

	go func() {
		for ; true; <-j.ticker.C {
			if addr, err = j.SourceFunc(ctx, &config.Source); err != nil {
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

			// run the target func
			if err = j.TargetFunc(ctx, addr, &j.Config.Target); err != nil {
				log.Warn(err)
			} else {
				// cache the new ip address
				j.lastIP = addr
			}

			// trigger the webhook, whatever target func is fail
			if config.WebHook != "" {
				log.Infof("start run webhook %s, %s %v %v", config.WebHook, addr.String(), err, config.Target.Domains)
				go j.RunWebhook(ctx, addr, err, config.Target.Domains)
			} else {
				log.Warn("webhook config is empty, so ignore")
			}
		}
	}()

	select {
	case <-j.done:
	case <-ctx.Done():
		return
	}
}

// Stop to stop a job
func (j *Job) Stop() {
	log.Debug("stopping job")
	j.done <- true
	j.ticker.Stop()
}

// NewJob for instance a new ddns job
func NewJob(config JobConfig) (*Job, error) {
	// check the configure
	if config.Source.Type == "" || config.Target.Type == "" {
		return nil, fmt.Errorf("source or target type can not be empty")
	}

	if config.Source.Interval <= 0 {
		return nil, fmt.Errorf("source check interval can not below zero or empty")
	}

	fnSource, err := SourceFunc(config.Source.Type)
	if err != nil {
		return nil, err
	}

	fnTarget, err := TargetFunc(config.Target.Type)
	if err != nil {
		return nil, err
	}

	return &Job{
		SourceFunc: fnSource,
		TargetFunc: fnTarget,
		Config:     &config,
		ticker:     time.NewTicker(time.Second * time.Duration(config.Source.Interval)),
		done:       make(chan bool),
	}, nil
}

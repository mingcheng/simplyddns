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

type WebHook struct {
	Method  string `yaml:"method,omitempty" default:"get" mapstructure:"method"`
	Url     string `yaml:"url,omitempty" mapstructure:"url"`
	Timeout uint   `yaml:"timeout" mapstructure:"timeout"`
}

type JobConfig struct {
	WebHook WebHook      `yaml:"webhook" mapstructure:"webhook"`
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
	webHook := j.Config.WebHook
	client := &http.Client{}

	if webHook.Timeout > 0 {
		log.Debugf("set http webhook client timeout %d", webHook.Timeout)
		client.Timeout = time.Duration(webHook.Timeout) * time.Second
	}

	if webHook.Method == "" {
		webHook.Method = "get"
	}

	log.Infof("set webhook request client method %s and url %s", webHook.Method, webHook.Url)
	req, err := http.NewRequest(webHook.Method, webHook.Url, nil)
	if err != nil {
		log.Warn(err.Error())
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
		log.Infof("run webhook %s is finished, status code is %v", webHook.Url, resp.StatusCode)
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

			if len(config.WebHook.Url) > 0 {
				// trigger the webhook
				log.Infof("start run webhook %s, with method %s", config.WebHook.Url, config.WebHook.Method)
				go j.RunWebhook(ctx, addr, err, config.Target.Domains)
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

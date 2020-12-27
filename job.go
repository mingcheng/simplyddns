/**
 * File: job.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Friday, December 25th 2020, 10:45:54 pm
 * Last Modified: Sunday, December 27th 2020, 7:38:30 pm
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

const DefaultSourceInterval = 5

type JobSource struct {
	Interval uint   `yaml:"interval,omitempty"`
	Type     string `yaml:"type,omitempty"`
	Path     string `yaml:"path"`
	Content  string `yaml:"content"`
}

type JobTarget struct {
	Type    string   `yaml:"type,omitempty"`
	Key     string   `yaml:"key,omitempty"`
	Token   string   `yaml:"token"`
	Proxy   string   `yaml:"proxy"`
	Domains []string `yaml:"domains,omitempty"`
}

type JobConfig struct {
	WebHook string    `yaml:"webhook" mapstructure:"webhook"`
	Source  JobSource `yaml:"source,omitempty" mapstructure:"source"`
	Target  JobTarget `yaml:"target,omitempty" mapstructure:"target"`
}

type Job struct {
	Config     *JobConfig
	SourceFunc func(context.Context, *JobSource) (*net.IP, error)
	TargetFunc func(context.Context, *net.IP, *JobTarget) error
	ticker     *time.Ticker
	done       chan bool
	lastIP     *net.IP
}

// RunWebhook to run the webhook when ip address has updated
func (j *Job) RunWebhook(ip *net.IP, e error, domains []string) error {
	if j.Config.WebHook != "" {

		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		req, err := http.NewRequest("GET", j.Config.WebHook, nil)
		if err != nil {
			return err
		}

		req.Header.Add("DDNS-New-Address", ip.String())
		req.Header.Add("DDNS-Domains", strings.Join(domains, ","))
		if e != nil {
			req.Header.Add("DDNS-Error", e.Error())
		}

		if _, err := client.Do(req); err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("webhook address is nil")
}

func (j *Job) Start(ctx context.Context) {
	go func() {
		for ; true; <-j.ticker.C {
			addr, err := j.SourceFunc(ctx, &j.Config.Source)
			log.Debugf("get adress from %v source func is %v", j.Config.Source.Type, addr.String())

			// cache addr
			if err != nil {
				log.Error(err)
				continue
			}

			// ignore the same ip address
			if j.lastIP != nil && j.lastIP.Equal(*addr) {
				log.Infof("ignore the same cached address %s", addr.String())
				continue
			}

			err = j.TargetFunc(ctx, addr, &j.Config.Target)
			log.Debug(err)

			// cache the new ip address
			j.lastIP = addr

			// trigger the webhook
			log.Infof("start run webhook, %s %v %v", addr.String(), err, j.Config.Target.Domains)
			_ = j.RunWebhook(addr, err, j.Config.Target.Domains)
		}
	}()

	select {
	case <-j.done:
	case <-ctx.Done():
		return
	}
}

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

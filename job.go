/**
 * File: job.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Friday, December 25th 2020, 10:45:54 pm
 * Last Modified: Wednesday, July 13th 2022, 12:26:13 pm
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
	Method  string `yaml:"method,omitempty" default:"GET" mapstructure:"method"`
	Url     string `yaml:"url,omitempty" mapstructure:"url"`
	Timeout uint   `yaml:"timeout" mapstructure:"timeout"`
}

// type Notification struct {
// 	MQ         string `yaml:"mq" default:"nsq" mapstructure:"mq"`
// 	Type       string `yaml:"type" default:"bark" mapstructure:"type"`
// 	Addr       string `yaml:"addr,omitempty" mapstructure:"addr"`
// 	Topic      string `yaml:"topic,omitempty" mapstructure:"topic"`
// 	Exchange   string `yaml:"exchange,omitempty" mapstructure:"exchange"`
// 	RoutingKey string `yaml:"routing_key,omitempty" mapstructure:"routing_key"`
// 	Receiver   string `yaml:"receiver" mapstructure:"receiver"`
// 	Subject    string `yaml:"subject,omitempty" mapstructure:"subject"`
// 	Content    string `yaml:"content,omitempty" mapstructure:"content"`
// }

type JobConfig struct {
	WebHook WebHook `yaml:"webhook" mapstructure:"webhook"`
	// Notification Notification `yaml:"notify" mapstructure:"notify"`
	Source SourceConfig `yaml:"source,omitempty" mapstructure:"source"`
	Target TargetConfig `yaml:"target,omitempty" mapstructure:"target"`
}

type Job struct {
	Config     *JobConfig
	SourceFunc []SourceFunc
	TargetFunc TargetFunc
	ticker     *time.Ticker
	done       chan bool
	lastIP     *net.IP
}

// RunWebhook to run the webhook when ip address has updated
func (j *Job) RunWebhook(ctx context.Context, ip *net.IP, _ error, domains []string) error {
	webHook := j.Config.WebHook
	client := &http.Client{}

	if webHook.Timeout > 0 {
		log.Debugf("set http webhook client timeout %d", webHook.Timeout)
		client.Timeout = time.Duration(webHook.Timeout) * time.Second
	}

	if webHook.Method == "" {
		webHook.Method = "GET"
	}

	log.Infof("set webhook request client method %s and url %s", webHook.Method, webHook.Url)
	req, err := http.NewRequest(strings.ToUpper(webHook.Method), webHook.Url, nil)
	if err != nil {
		log.Warn(err.Error())
		return err
	}

	req.Header.Add("SimplyDDNS-Address", ip.String())
	req.Header.Add("SimplyDDNS-Domains", strings.Join(domains, ","))
	req.WithContext(ctx)

	if _, err := client.Do(req); err != nil {
		return err
	}

	return nil
}

// Start to start a job
func (j *Job) Start(ctx context.Context) {
	go func() {
		var (
			err  error
			addr *net.IP
			job  = j
		)

		for ; true; <-job.ticker.C {
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
			if job.lastIP != nil && job.lastIP.Equal(*addr) {
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
			if err = job.TargetFunc(ctx, addr, &job.Config.Target); err != nil {
				log.Warn(err)
				continue
			}

			// cache the last ip address
			job.lastIP = addr

			// trigger the webhook if configured
			if len(config.WebHook.Url) > 0 {
				if err = job.RunWebhook(ctx, addr, err, domains); err != nil {
					log.Warnf("run webhook with error %s", err.Error())
				} else {
					log.Infof("run webhook %s is finished", config.WebHook.Url)
				}
			}
		}
	}()

	select {
	case <-j.done:
	case <-ctx.Done():
		j.ticker.Stop()
		return
	}
}

// Stop to stop a job
func (j *Job) Stop() {
	log.Debug("stopping job")
	j.done <- true
}

// Source to execute multi-source function
func (j Job) Source(ctx context.Context, config *SourceConfig) (*net.IP, error) {
	if j.SourceFunc == nil || len(j.SourceFunc) == 0 {
		return nil, fmt.Errorf("source functions is empty")
	}

	var (
		err      error
		errTimes int
		lastAddr *net.IP
	)

	for _, v := range j.SourceFunc {
		var addr *net.IP
		addr, err = v(ctx, config)
		if err != nil {
			log.Error(err, errTimes)
			errTimes = errTimes + 1
		}

		if addr != nil {
			if lastAddr != nil && !addr.Equal(*lastAddr) {
				return nil, fmt.Errorf("fetch address is not the same, %v vs %v", lastAddr, addr)
			}

			lastAddr = addr
		}
	}

	if errTimes > 0 && len(sourceFuncs) > 3 && errTimes >= len(j.SourceFunc)/2 {
		return nil, fmt.Errorf("max error times reached(%d), so the result is not right", errTimes)
	}

	return lastAddr, nil
}

// NewJob for instance a new ddns job
func NewJob(config JobConfig) (job *Job, err error) {
	// check the configure
	if config.Source.Type == "" || config.Target.Type == "" {
		err = fmt.Errorf("source or target type can not be empty")
		return
	}

	if config.Source.Interval <= 0 {
		err = fmt.Errorf("source check interval can not below zero or empty")
		return
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
		fn, err = SourceFuncByName(strings.ToLower(v))
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

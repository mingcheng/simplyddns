/*!*
 * Copyright (c) 2026 Ming Lyu, aka mingcheng
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: dispatch.go
 * Author: mingcheng <mingcheng@apache.org>
 * File Created: Friday, December 25th 2020, 10:46:17 pm
 *
 * Modified By: mingcheng <mingcheng@apache.org>
 * Last Modified: 2026-10-27 14:49:57
 */

package simplyddns

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Dispatch coordinates a set of DDNS jobs that run concurrently.
type Dispatch struct {
	wg      sync.WaitGroup
	mu      sync.Mutex
	started bool
	stopped bool
	jobs    []*Job
	Timeout time.Duration
	Configs []JobConfig
}

// Start runs every job concurrently and blocks until all of them return.
func (d *Dispatch) Start(ctx context.Context) {
	d.mu.Lock()
	if d.stopped || d.started {
		d.mu.Unlock()
		d.wg.Wait()
		return
	}
	d.started = true
	d.wg.Add(len(d.jobs))
	d.mu.Unlock()
	for _, v := range d.jobs {
		v.timeout = d.Timeout
		go func(job *Job) {
			defer d.wg.Done()
			job.Start(ctx)
		}(v)
	}

	d.wg.Wait()
}

// Stop signals every job to stop and waits for them to return.
func (d *Dispatch) Stop() {
	d.mu.Lock()
	d.stopped = true
	for _, v := range d.jobs {
		v.Stop()
	}
	d.mu.Unlock()
	d.wg.Wait()
}

// NewDispatch creates a new Dispatch from the given job configurations.
func NewDispatch(configs []JobConfig) (*Dispatch, error) {
	if len(configs) == 0 {
		return nil, fmt.Errorf("at least one DDNS job is required")
	}
	dispatch := &Dispatch{
		Configs: configs,
		Timeout: DefaultTimeout,
	}

	for i, v := range configs {
		job, err := NewJob(v)
		if err != nil {
			dispatch.Stop()
			return nil, fmt.Errorf("job %d: %w", i+1, err)
		}
		log.Debugf("add job %d to dispatch queue", i+1)
		dispatch.jobs = append(dispatch.jobs, job)
	}

	return dispatch, nil
}

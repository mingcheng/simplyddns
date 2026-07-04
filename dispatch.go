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
 * Last Modified: 2026-05-12 12:23:33
 */

package simplyddns

import (
	"context"
	"sync"
	"time"
)

// Dispatch coordinates a set of DDNS jobs that run concurrently.
type Dispatch struct {
	wg      sync.WaitGroup
	jobs    []*Job
	Timeout time.Duration
	Configs []JobConfig
}

// Start runs every job concurrently and blocks until all of them return.
func (d *Dispatch) Start(ctx context.Context) {
	for _, v := range d.jobs {
		d.wg.Add(1)
		go func(job *Job) {
			defer d.wg.Done()
			job.Start(ctx)
		}(v)
	}

	d.wg.Wait()
}

// Stop signals every job to stop and waits for them to return.
func (d *Dispatch) Stop() {
	for _, v := range d.jobs {
		v.Stop()
	}
	d.wg.Wait()
}

// NewDispatch creates a new Dispatch from the given job configurations.
func NewDispatch(configs []JobConfig) (*Dispatch, error) {
	log.Debugf("new dispatch instance with configure: %v", configs)
	dispatch := &Dispatch{
		Configs: configs,
	}

	for _, v := range configs {
		job, err := NewJob(v)
		if err != nil {
			return nil, err
		}
		log.Debugf("add job %v to dispatch queue", job)
		dispatch.jobs = append(dispatch.jobs, job)
	}

	return dispatch, nil
}

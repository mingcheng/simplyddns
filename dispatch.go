/*!*
 * Copyright (c) 2025 Hangzhou Guanwaii Technology Co,.Ltd.
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: dispatch.go
 * Author: mingcheng (mingcheng@apache.org)
 * File Created: Friday, December 25th 2020, 10:46:17 pm
 *
 * Modified By: mingcheng (mingcheng@apache.org)
 * Last Modified: 2025-03-12 13:39:40
 */

package simplyddns

import (
	"context"
	"sync"
	"time"
)

type Dispatch struct {
	wg      sync.WaitGroup
	jobs    []*Job
	Timeout time.Duration
	Configs []JobConfig
}

// Start the dispatch
func (d *Dispatch) Start(ctx context.Context) {
	for _, v := range d.jobs {
		d.wg.Add(1)
		go v.Start(ctx)
	}

	d.wg.Wait()
}

// Stop the dispatch
func (d *Dispatch) Stop() {
	for _, v := range d.jobs {
		d.wg.Done()
		go v.Stop()
	}
}

func NewDispatch(configs []JobConfig) (*Dispatch, error) {
	log.Debugf("new displatch instance with configure: %v", configs)
	dispatch := &Dispatch{
		Configs: configs,
	}

	for _, v := range configs {
		if job, err := NewJob(v); err != nil {
			return nil, err
		} else {
			log.Debugf("add job %v to dispatch queue", job)
			dispatch.jobs = append(dispatch.jobs, job)
		}
	}

	return dispatch, nil
}

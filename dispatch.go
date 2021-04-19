/**
 * File: dispatch.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Friday, December 25th 2020, 10:46:17 pm
 * Last Modified: Sunday, December 27th 2020, 8:39:40 pm
 *
 * http://www.opensource.org/licenses/MIT
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

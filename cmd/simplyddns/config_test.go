/*!*
 * Copyright (c) 2025 Hangzhou Guanwaii Technology Co,.Ltd.
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: config_test.go
 * Author: mingcheng@outlook.com
 * File Created: Saturday, December 26th 2020
 *
 * Modified By: mingcheng (mingcheng@apache.org)
 * Last Modified: 2025-03-12 14:36:30
 */

package main

import (
	"context"
	"testing"

	"github.com/mingcheng/simplyddns"
	"github.com/stretchr/testify/assert"

	_ "github.com/mingcheng/simplyddns/target"
)

func TestMultiSourceFunc(t *testing.T) {
	job, err := simplyddns.NewJob(simplyddns.JobConfig{
		Source: simplyddns.SourceConfig{
			Type:     "ipify",
			Interval: 1,
		},
		Target: simplyddns.TargetConfig{
			Type: "sleep",
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, job)

	result, err := job.Source(context.TODO(), &simplyddns.SourceConfig{})
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

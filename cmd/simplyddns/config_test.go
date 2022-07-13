package main

import (
	"context"
	"testing"

	"github.com/mingcheng/simplyddns"
	"github.com/stretchr/testify/assert"
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

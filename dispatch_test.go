package simplyddns

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDispatchConcurrentStartAndStop(t *testing.T) {
	config := testJobConfig(t)
	for range 100 {
		dispatch, err := NewDispatch([]JobConfig{config, config})
		require.NoError(t, err)
		var wg sync.WaitGroup
		wg.Go(func() { dispatch.Start(context.Background()) })
		wg.Go(func() { dispatch.Start(context.Background()) })
		wg.Go(dispatch.Stop)
		wg.Go(dispatch.Stop)
		wg.Wait()
		dispatch.Stop()
		for _, job := range dispatch.jobs {
			assert.ErrorIs(t, job.stopCtx.Err(), context.Canceled)
		}
	}
}

func TestDispatchInvalidConfig(t *testing.T) {
	_, err := NewDispatch(nil)
	require.Error(t, err)
	_, err = NewDispatch([]JobConfig{testJobConfig(t), {}})
	require.ErrorContains(t, err, "job 2")
}

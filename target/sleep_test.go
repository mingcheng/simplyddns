package target

import (
	"context"
	"net"
	"testing"
	"testing/synctest"
	"time"

	ddns "github.com/mingcheng/simplyddns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSleep(t *testing.T) {
	fn, err := ddns.TargetFuncByName("sleep")
	require.NoError(t, err)
	ip := net.ParseIP("192.0.2.1")
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()
		require.NoError(t, fn(context.Background(), &ip, nil))
		assert.Equal(t, 10*time.Second, time.Since(start))
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.ErrorIs(t, fn(ctx, &ip, nil), context.Canceled)
}

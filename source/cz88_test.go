package source

import (
	"context"
	"github.com/mingcheng/simplyddns"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSourceCZ88(t *testing.T) {
	fn, err := simplyddns.SourceFuncByName("cz88")
	assert.NoError(t, err)

	ip, err := fn(context.TODO(), nil)
	assert.NoError(t, err)

	assert.False(t, ip.IsLoopback())
	assert.NotEqual(t, ip.String(), "")

	log.Info(ip)
}

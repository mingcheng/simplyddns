package source

import (
	"context"
	"github.com/mingcheng/simplyddns"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSourceIpPlus360(t *testing.T) {
	fn, err := simplyddns.SourceFuncByName("ipplus360")
	assert.NoError(t, err)

	ip, err := fn(context.TODO(), nil)
	assert.NoError(t, err)

	assert.False(t, ip.IsLoopback())
	assert.NotEqual(t, ip.String(), "")

	log.Info(ip)
}

package source

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mingcheng/simplyddns"
)

func TestSourceMyIP(t *testing.T) {
	fn, err := simplyddns.SourceFuncByName("myip")
	assert.NoError(t, err)

	ip, err := fn(context.TODO(), nil)
	assert.NoError(t, err)

	assert.False(t, ip.IsLoopback())
	assert.NotEqual(t, ip.String(), "")

	log.Info(ip)
}

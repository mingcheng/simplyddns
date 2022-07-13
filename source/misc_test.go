package source

import (
	ddns "github.com/mingcheng/simplyddns"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetAllSupportSourceFunc(t *testing.T) {
	funcs := ddns.GetAllSupportSourceFunc()
	assert.NotEmpty(t, funcs)
}

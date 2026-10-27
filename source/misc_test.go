package source

import (
	ddns "github.com/mingcheng/simplyddns"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetAllSupportSourceFunc(t *testing.T) {
	names := []string{"auto", "auto-ipv4", "auto-ipv6", "lo", "static"}
	for _, tc := range httpSourceTests {
		names = append(names, tc.name)
	}
	assert.ElementsMatch(t, names, ddns.GetAllSupportSourceFunc())
}

/**
 * File: myip_test.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Saturday, December 26th 2020, 10:52:17 pm
 * Last Modified: Saturday, February 13th 2021, 10:39:30 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package source

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mingcheng/simplyddns"
)

func TestSourceIPSB(t *testing.T) {
	fn, err := simplyddns.SourceFunc("ipsb")
	assert.NoError(t, err)

	ip, err := fn(context.TODO(), nil)
	assert.NoError(t, err)
	assert.NotNil(t, ip)

	assert.False(t, ip.IsLoopback())
	assert.NotEqual(t, ip.String(), "")

	log.Info(ip)
}

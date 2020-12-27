/**
 * File: myip_test.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Saturday, December 26th 2020, 10:52:17 pm
 * Last Modified: Sunday, December 27th 2020, 2:08:47 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package source

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/mingcheng/simplyddns"
)

func TestSourceMyIPIP(t *testing.T) {
	fn, err := simplyddns.SourceFunc("myipip")
	assert.NoError(t, err)

	ip, err := fn(context.TODO(), nil)
	assert.NoError(t, err)

	assert.False(t, ip.IsLoopback())
	assert.NotEqual(t, ip.String(), "")

	log.Info(ip)
}

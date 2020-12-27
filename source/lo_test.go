/**
 * File: lo_test.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Sunday, December 27th 2020, 2:23:23 pm
 * Last Modified: Sunday, December 27th 2020, 8:38:43 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package source

import (
	"context"
	"testing"

	"github.com/mingcheng/simplyddns"
	"github.com/stretchr/testify/assert"
)

func TestSourceLookback(t *testing.T) {
	fn, err := simplyddns.SourceFunc("lo")
	assert.NoError(t, err)

	ip, err := fn(context.TODO(), nil)
	assert.NoError(t, err)
	assert.True(t, ip.IsLoopback())
}

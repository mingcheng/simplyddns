/**
 * File: static_test.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Sunday, December 27th 2020, 2:13:46 pm
 * Last Modified: Sunday, December 27th 2020, 8:39:07 pm
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

func TestSourceStatic(t *testing.T) {
	const Address = "192.168.1.1"
	fn, err := simplyddns.SourceFuncByName("static")
	assert.NoError(t, err)
	assert.NotNil(t, fn)

	ip, err := fn(context.TODO(), &simplyddns.SourceConfig{
		Content: Address,
	})

	assert.NoError(t, err)
	assert.Equal(t, ip.String(), Address)
}

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
	"github.com/stretchr/testify/require"
)

func TestSourceStatic(t *testing.T) {
	fn, err := simplyddns.SourceFuncByName("static")
	require.NoError(t, err)
	for _, tc := range []struct{ content, want string }{
		{"192.168.1.1", "192.168.1.1"},
		{" 2001:db8::1\n", "2001:db8::1"},
		{"", ""},
		{"not-an-address", ""},
	} {
		t.Run(tc.content, func(t *testing.T) {
			ip, err := fn(context.Background(), &simplyddns.SourceConfig{Content: tc.content})
			if tc.want == "" {
				require.Error(t, err)
				assert.Nil(t, ip)
			} else {
				require.NoError(t, err)
				require.NotNil(t, ip)
				assert.Equal(t, tc.want, ip.String())
			}
		})
	}
	_, err = fn(context.Background(), nil)
	require.Error(t, err)
}

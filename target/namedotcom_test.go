/**
 * File: namedotcom_test.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Saturday, December 26th 2020, 11:04:07 pm
 * Last Modified: Sunday, December 27th 2020, 8:59:14 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package target

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"github.com/mingcheng/simplyddns"
	"github.com/stretchr/testify/assert"
)

func TestTargetNameDotCom(t *testing.T) {
	const Address = "172.16.1.1"
	ip := net.ParseIP(Address)

	fn, err := simplyddns.TargetFunc("namedotcom")
	assert.NoError(t, err)
	assert.NotNil(t, fn)

	ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Second)
	defer cancel()
	go fn(ctx, &ip, &simplyddns.JobTarget{
		Key:     os.Getenv("NAME_COM_KEY"),
		Token:   os.Getenv("NAME_COM_TOKEN"),
		Proxy:   os.Getenv("NAME_COM_PROXY"),
		Domains: []string{"a.example.org", "b.example.org"},
	})

	select {
	case <-ctx.Done():
		// do nothing
	}
}

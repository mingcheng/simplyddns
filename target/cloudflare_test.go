/**
 * File: alidns_test.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Monday, December 28th 2020, 11:10:47 am
 * Last Modified: Monday, December 28th 2020, 2:47:07 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package target

import (
	"context"
	"os"
	"testing"

	"github.com/mingcheng/simplyddns"
	_ "github.com/mingcheng/simplyddns/source"

	"github.com/stretchr/testify/assert"
)

func TestNewCloudflareDNSClient(t *testing.T) {
	if _, exists := os.LookupEnv("CLOUDFLARE_DNS_TOKEN"); !exists {
		return
	}

	var err error

	source, err := simplyddns.SourceFuncByName("cloudflare")
	assert.NoError(t, err)
	assert.NotNil(t, source)

	target, err := simplyddns.TargetFuncByName("cloudflare")
	assert.NoError(t, err)

	ip, err := source(context.TODO(), nil)
	assert.NoError(t, err)

	err = target(context.TODO(), ip, &simplyddns.TargetConfig{
		Key:     os.Getenv("CLOUDFLARE_DNS_KEY"),
		Token:   os.Getenv("CLOUDFLARE_DNS_TOKEN"),
		Domains: []string{"1.intkd.com", "2.intkd.com"},
	})

	assert.NoError(t, err)
}

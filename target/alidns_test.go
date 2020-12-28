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

func TestNewAliDNS(t *testing.T) {
	var err error

	source, err := simplyddns.SourceFunc("myipip")
	assert.NoError(t, err)
	assert.NotNil(t, source)

	target, err := simplyddns.TargetFunc("alidns")
	assert.NoError(t, err)

	ip, err := source(context.TODO(), nil)
	assert.NoError(t, err)

	err = target(context.TODO(), ip, &simplyddns.TargetConfig{
		Key:     os.Getenv("ALIYUN_DNS_KEY"),
		Token:   os.Getenv("ALIYUN_DNS_TOKEN"),
		Domains: []string{"1.example.com", "2.example.com"},
	})

	// throw error because token and key is not set
	assert.Error(t, err)
}

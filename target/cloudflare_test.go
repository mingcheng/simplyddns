/*!*
 * Copyright (c) 2025-2026 Ming Lyu, aka mingcheng
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: cloudflare_test.go
 * Author: mingcheng <mingcheng@apache.org>
 * File Created: Monday, December 28th 2020, 11:10:47 am
 *
 * Modified By: mingcheng <mingcheng@apache.org>
 * Last Modified: 2026-05-12 12:24:11
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

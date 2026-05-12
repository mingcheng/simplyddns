/*!*
 * Copyright (c) 2025-2026 Ming Lyu, aka mingcheng
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: ipinfo_test.go
 * Author: mingcheng <mingcheng@apache.org>
 * File Created: 2025-11-27 17:04:40
 *
 * Modified By: mingcheng <mingcheng@apache.org>
 * Last Modified: 2026-05-12 12:13:21
 */

package source

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mingcheng/simplyddns"
)

func TestSourceIPInfo(t *testing.T) {
	fn, err := simplyddns.SourceFuncByName("ipinfo")
	assert.NoError(t, err)

	ip, err := fn(context.TODO(), nil)
	if err != nil {
		t.Skipf("ipinfo lookup failed (probably rate-limited): %v", err)
	}
	if !assert.NotNil(t, ip) {
		return
	}

	assert.False(t, ip.IsLoopback())
	assert.NotEqual(t, ip.String(), "")

	log.Info(ip)
}

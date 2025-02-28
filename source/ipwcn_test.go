/*!*
 * Copyright (c) 2024-2025 Hangzhou Guanwaii Technology Co,.Ltd.
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: ipwcn_test.go
 * Author: mingcheng (mingcheng@apache.org)
 * File Created: 2024-11-15 14:46:48
 *
 * Modified By: mingcheng (mingcheng@apache.org)
 * Last Modified: 2025-02-28 10:48:19
 */

package source

import (
	"context"
	"github.com/mingcheng/simplyddns"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSourceIPWcn(t *testing.T) {
	fn, err := simplyddns.SourceFuncByName("ipwcn")
	assert.NoError(t, err)

	ip, err := fn(context.TODO(), nil)
	assert.NoError(t, err)

	assert.False(t, ip.IsLoopback())
	assert.NotEqual(t, ip.String(), "")

	log.Info(ip)
}

/*!*
 * Copyright (c) 2025 Hangzhou Guanwaii Technology Co,.Ltd.
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: func_test.go
 * Author: mingcheng (mingcheng@apache.org)
 * File Created: 2025-03-12 13:23:45
 *
 * Modified By: mingcheng (mingcheng@apache.org)
 * Last Modified: 2025-03-12 13:39:01
 */

package simplyddns

import (
	"context"
	"fmt"
	"net"
	"slices"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFunctionRegistry(t *testing.T) {
	config := testJobConfig(t)
	fn, err := SourceFuncByName("  " + config.Source.Type + "  ")
	require.NoError(t, err)
	require.NotNil(t, fn)
	target, err := TargetFuncByName("  " + config.Target.Type + "  ")
	require.NoError(t, err)
	require.NotNil(t, target)
	require.Error(t, RegisterSourceFunc(config.Source.Type, fn))
	require.Error(t, RegisterTargetFunc(config.Target.Type, target))
	require.Error(t, RegisterSourceFunc("", fn))
	require.Error(t, RegisterTargetFunc(" ", target))
	require.Error(t, RegisterSourceFunc("nil-source", nil))
	require.Error(t, RegisterTargetFunc("nil-target", nil))
	_, err = SourceFuncByName("unknown-source")
	require.Error(t, err)
	_, err = TargetFuncByName("unknown-target")
	require.Error(t, err)
	assert.True(t, slices.IsSorted(GetAllSupportSourceFunc()))
}

func TestFunctionRegistryConcurrentAccess(t *testing.T) {
	var wg sync.WaitGroup
	names := make([]string, 32)
	for i := range names {
		names[i] = fmt.Sprintf("concurrent-%d", i)
		wg.Go(func() {
			name := names[i]
			err := RegisterSourceFunc(name, func(context.Context, *SourceConfig) (*net.IP, error) {
				ip := net.ParseIP("192.0.2.1")
				return &ip, nil
			})
			assert.NoError(t, err)
			_, err = SourceFuncByName(name)
			assert.NoError(t, err)
			assert.True(t, slices.IsSorted(GetAllSupportSourceFunc()))
		})
	}
	wg.Wait()
	t.Cleanup(func() {
		registryMu.Lock()
		defer registryMu.Unlock()
		for _, name := range names {
			delete(sourceFuncs, name)
		}
	})
}

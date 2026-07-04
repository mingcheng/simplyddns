/*!*
 * Copyright (c) 2025-2026 Ming Lyu, aka mingcheng
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: namedotcom_test.go
 * Author: mingcheng <mingcheng@apache.org>
 * File Created:  Saturday, December 26th 2020, 11:04:07 pm
 *
 * Modified By: mingcheng <mingcheng@apache.org>
 * Last Modified: 2026-05-12 12:15:30
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
	if _, exists := os.LookupEnv("NAME_COM_PROXY"); !exists {
		return
	}

	const Address = "172.16.1.1"
	ip := net.ParseIP(Address)

	fn, err := simplyddns.TargetFuncByName("namedotcom")
	assert.NoError(t, err)
	assert.NotNil(t, fn)

	ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Second)
	defer cancel()
	go func() {
		_ = fn(ctx, &ip, &simplyddns.TargetConfig{
			Key:     os.Getenv("NAME_COM_KEY"),
			Token:   os.Getenv("NAME_COM_TOKEN"),
			Proxy:   os.Getenv("NAME_COM_PROXY"),
			Domains: []string{"a.example.org", "b.example.org"},
		})
	}()

	<-ctx.Done()
}

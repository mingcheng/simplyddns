/*!*
 * Copyright (c) 2025-2026 Ming Lyu, aka mingcheng
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: lo.go
 * Author: mingcheng <mingcheng@apache.org>
 * File Created: Saturday, December 26th 2020, 3:01:39 pm
 *
 * Modified By: mingcheng <mingcheng@apache.org>
 * Last Modified: 2026-05-12 12:25:20
 */

package source

import (
	"context"
	"fmt"
	"net"

	ddns "github.com/mingcheng/simplyddns"
)

func init() {
	fn := func(_ context.Context, _ *ddns.SourceConfig) (*net.IP, error) {
		ip := net.ParseIP("127.0.0.1")
		if !ip.IsLoopback() {
			return nil, fmt.Errorf("%v is not a loopback address", ip)
		}

		return &ip, nil
	}

	_ = ddns.RegisterSourceFunc("lo", fn)
}

/*!*
 * Copyright (c) 2025-2026 Ming Lyu, aka mingcheng
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: ipinfo.go
 * Author: mingcheng <mingcheng@apache.org>
 * File Created: 2025-11-27 17:04:40
 *
 * Modified By: mingcheng <mingcheng@apache.org>
 * Last Modified: 2026-05-12 12:13:28
 */

package source

import (
	"context"
	"net"

	"github.com/ipinfo/go/v2/ipinfo"
	"github.com/mingcheng/simplyddns"
)

func init() {
	const (
		Name = "ipinfo"
	)

	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		result, err := ipinfo.GetIPAddr()
		if err != nil {
			return nil, err
		}

		ip := net.ParseIP(result)
		return &ip, nil
	}

	_ = simplyddns.RegisterSourceFunc(Name, fn)
}

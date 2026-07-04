/*!*
 * Copyright (c) 2025-2026 Ming Lyu, aka mingcheng
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: myip.go
 * Author: mingcheng <mingcheng@apache.org>
 * File Created: 2025-11-27 17:04:40
 *
 * Modified By: mingcheng <mingcheng@apache.org>
 * Last Modified: 2026-05-12 12:25:00
 */

package source

import (
	"context"
	"net"

	"github.com/mingcheng/simplyddns"
)

func init() {
	const Name = "myip"

	fn := func(_ context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		data, err := RawIPByURL("https://api.my-ip.io/ip.txt")
		if err != nil {
			return nil, err
		}
		return &data, nil
	}

	_ = simplyddns.RegisterSourceFunc(Name, fn)
}

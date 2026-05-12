/*!*
 * Copyright (c) 2025-2026 Ming Lyu, aka mingcheng
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: static.go
 * Author: mingcheng <mingcheng@apache.org>
 * File Created: Sunday, December 27th 2020, 2:07:45 pm
 *
 * Modified By: mingcheng <mingcheng@apache.org>
 * Last Modified: 2026-05-12 12:24:47
 */

package source

import (
	"context"
	"net"
	"strings"

	ddns "github.com/mingcheng/simplyddns"
)

func init() {
	fn := func(_ context.Context, source *ddns.SourceConfig) (*net.IP, error) {
		addr := net.ParseIP(strings.TrimSpace(source.Content))
		if addr == nil {
			return nil, nil
		}
		log.Debugf("parsed address is %s", addr.String())
		return &addr, nil
	}

	_ = ddns.RegisterSourceFunc("static", fn)
}

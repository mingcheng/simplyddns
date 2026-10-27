/*!*
 * Copyright (c) 2025-2026 Ming Lyu, aka mingcheng
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: cloudflare.go
 * Author: mingcheng <mingcheng@apache.org>
 * File Created: 2025-11-27 17:04:40
 *
 * Modified By: mingcheng <mingcheng@apache.org>
 * Last Modified: 2026-05-12 12:13:14
 */

package source

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/mingcheng/simplyddns"
)

func init() {
	const Name = "cloudflare"

	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		data, err := RawStrByURL(ctx, "https://www.cloudflare.com/cdn-cgi/trace", nil)
		if err != nil {
			return nil, err
		}

		for _, line := range strings.Split(data, "\n") {
			key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
			if ok && key == "ip" {
				return parseIP(value)
			}
		}

		return nil, fmt.Errorf("cloudflare trace does not contain an IP address")
	}

	registerSource(Name, fn)
}

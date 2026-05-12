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
	"net"
	"regexp"

	"github.com/mingcheng/simplyddns"
)

func init() {
	const Name = "cloudflare"

	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		data, err := RawStrByURL(ctx, "https://www.cloudflare.com/cdn-cgi/trace", nil)
		if err != nil {
			return nil, err
		}

		var re = regexp.MustCompile(`(?m)ip=([\d.]+)`)

		for _, match := range re.FindAllStringSubmatch(data, -1) {
			if match[1] != "" {
				addr := net.ParseIP(match[1])
				return &addr, nil
			}
		}

		return nil, nil
	}

	_ = simplyddns.RegisterSourceFunc(Name, fn)
}

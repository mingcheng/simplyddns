/*!*
 * Copyright (c) 2026 Ming Lyu, aka mingcheng
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: ipsb.go
 * Author: mingcheng <mingcheng@apache.org>
 * File Created: 2026-07-04 22:59:02
 *
 * Modified By: mingcheng <mingcheng@apache.org>
 * Last Modified: 2026-07-04 23:01:35
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
	const Name = "ipsb"

	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		log.Debugf("%s start requests", Name)

		resp, err := RawStrByURL(ctx, "https://api-ipv4.ip.sb/ip", map[string]string{
			"User-Agent": UserAgentCurl,
		})
		if err != nil {
			return nil, err
		}

		addr := net.ParseIP(strings.TrimSpace(resp))
		if addr == nil {
			return nil, fmt.Errorf("%s returned invalid IP %q", Name, resp)
		}

		log.Debugf("%s remote address is %s", Name, addr.String())
		return &addr, nil
	}

	_ = simplyddns.RegisterSourceFunc(Name, fn)
}

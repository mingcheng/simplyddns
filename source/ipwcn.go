/*!*
 * Copyright (c) 2025 Hangzhou Guanwaii Technology Co,.Ltd.
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: ipwcn.go
 * Author: mingcheng (mingcheng@apache.org)
 * File Created: 2024-11-15 14:46:48
 *
 * Modified By: mingcheng (mingcheng@apache.org)
 * Last Modified: 2025-07-14 14:18:40
 */

package source

import (
	"context"
	"fmt"
	"net"

	"github.com/mingcheng/simplyddns"
)

func init() {
	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		log.Debugf("start requests from %s", "4.ipw.cn")
		resp, err := RawStrByURL(ctx, "https://4.ipw.cn/", map[string]string{
			"User-Agent": UserAgentCurl,
		})
		if err != nil {
			log.Debug(err)
			return nil, err
		}

		addr := net.ParseIP(resp)
		if addr == nil {
			return nil, fmt.Errorf("failed to parse IP address: %s", resp)
		}

		log.Debugf("remote address is %s", addr.String())
		return &addr, nil
	}

	_ = simplyddns.RegisterSourceFunc("ipwcn", fn)
}

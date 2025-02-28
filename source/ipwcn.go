/*!*
 * Copyright (c) 2024-2025 Hangzhou Guanwaii Technology Co,.Ltd.
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: ipwcn.go
 * Author: mingcheng (mingcheng@apache.org)
 * File Created: 2024-11-15 14:46:48
 *
 * Modified By: mingcheng (mingcheng@apache.org)
 * Last Modified: 2025-02-28 10:48:25
 */

package source

import (
	"context"
	"github.com/mingcheng/simplyddns"
	"net"
)

func init() {
	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		log.Debugf("start requests from %s", "4.ipw.cn")
		resp, err := RawStrByURL(context.Background(), "https://4.ipw.cn/", map[string]string{
			"User-Agent": UserAgentCurl,
		})
		if err != nil {
			log.Debug(err)
			return nil, err
		}

		if err != nil {
			log.Debug(err)
			return nil, err
		}

		addr := net.ParseIP(resp)
		log.Debugf("remote address is %s", addr.String())
		return &addr, nil
	}

	_ = simplyddns.RegisterSourceFunc("ipwcn", fn)
}

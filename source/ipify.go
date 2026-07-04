/*!*
 * Copyright (c) 2025-2026 Ming Lyu, aka mingcheng
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: ipify.go
 * Author: mingcheng <mingcheng@apache.org>
 * File Created: Saturday, December 26th 2020, 10:41:38 pm
 *
 * Modified By: mingcheng <mingcheng@apache.org>
 * Last Modified: 2026-05-12 12:25:56
 */

package source

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/mingcheng/simplyddns"
)

func init() {
	const Name = "ipify.org"
	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		log.Debugf("%s start requests", Name)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.ipify.org?format=text", nil)
		if err != nil {
			return nil, err
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("%s returned status %s", Name, resp.Status)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		addr := net.ParseIP(strings.TrimSpace(string(body)))
		if addr == nil {
			return nil, fmt.Errorf("%s returned invalid IP %q", Name, string(body))
		}
		log.Debugf("%s remote address is %s", Name, addr.String())
		return &addr, nil
	}

	_ = simplyddns.RegisterSourceFunc("ipify", fn)
}

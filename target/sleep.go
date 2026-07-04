/*!*
 * Copyright (c) 2025-2026 Ming Lyu, aka mingcheng
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: sleep.go
 * Author: mingcheng <mingcheng@apache.org>
 * File Created: Saturday, December 26th 2020, 3:01:44 pm
 *
 * Modified By: mingcheng <mingcheng@apache.org>
 * Last Modified: 2026-05-12 12:15:18
 */

package target

import (
	"context"
	"net"
	"time"

	ddns "github.com/mingcheng/simplyddns"
)

func init() {
	log.Tracef("register target function which name is sleep")

	_ = ddns.RegisterTargetFunc("sleep", func(_ context.Context, addr *net.IP, _ *ddns.TargetConfig) error {
		log.Debugf("sleep target, received address %s", addr.String())
		time.Sleep(10 * time.Second)
		return nil
	})
}

/*!*
 * Copyright (c) 2026 Ming Lyu, aka mingcheng
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: constants.go
 * Author: mingcheng <mingcheng@apache.org>
 * File Created: 2025-07-14
 *
 * Modified By: mingcheng <mingcheng@apache.org>
 * Last Modified: 2026-05-12 12:12:02
 */

package simplyddns

import "time"

const (
	// DefaultInterval is the default check interval in seconds
	DefaultInterval = 60

	// MinInterval is the minimum allowed check interval in seconds
	MinInterval = 30

	// MaxRetries is the maximum number of retries for failed requests
	MaxRetries = 3

	// DefaultTimeout is the default timeout for HTTP requests
	DefaultTimeout = 30 * time.Second

	// DefaultUserAgent is the default user agent for HTTP requests
	DefaultUserAgent = "SimpleDDNS/1.0"
)

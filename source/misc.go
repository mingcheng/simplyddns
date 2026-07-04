/*!*
 * Copyright (c) 2026 Ming Lyu, aka mingcheng
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: misc.go
 * Author: mingcheng <mingcheng@apache.org>
 * File Created: Sunday, December 27th 2020, 7:01:38 pm
 *
 * Modified By: mingcheng <mingcheng@apache.org>
 * Last Modified: 2026-05-12 12:25:10
 */

package source

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/go-resty/resty/v2"
	ddns "github.com/mingcheng/simplyddns"
)

var log = ddns.NewLogger()

const (
	// UserAgent mimics a modern desktop browser for endpoints that filter
	// common HTTP clients.
	UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36"

	// UserAgentCurl mimics curl for endpoints that return plain text bodies
	// only when the User-Agent looks like a CLI tool.
	UserAgentCurl = "curl/7.54.1"
)

// RawIPByURL fetches the response body of url and parses it as a single IP.
func RawIPByURL(url string) (net.IP, error) {
	data, err := RawStrByURL(context.Background(), url, nil)
	if err != nil {
		return nil, err
	}

	addr := net.ParseIP(strings.TrimSpace(data))
	if addr == nil {
		return nil, fmt.Errorf("failed to parse IP address from %q", data)
	}
	return addr, nil
}

// RawStrByURL performs an HTTP GET against url and returns the body as a
// string. Additional headers may be provided to override the defaults.
func RawStrByURL(ctx context.Context, url string, headers map[string]string) (string, error) {
	req := resty.New().R().
		SetContext(ctx).
		SetHeaders(map[string]string{
			"Referer":    url,
			"User-Agent": UserAgent,
		})

	if headers != nil {
		req.SetHeaders(headers)
	}

	resp, err := req.Get(url)
	if err != nil {
		return "", err
	}

	if !resp.IsSuccess() {
		return "", fmt.Errorf("HTTP request failed with status %s", resp.Status())
	}

	return string(resp.Body()), nil
}

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
	"io"
	"net"
	"net/http"
	"strings"

	ddns "github.com/mingcheng/simplyddns"
	"github.com/tidwall/gjson"
)

var httpClient = &http.Client{Timeout: ddns.DefaultTimeout}

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
	addr, err := rawIPByURL(context.Background(), url, nil)
	if err != nil {
		return nil, err
	}
	return *addr, nil
}

func parseIP(data string) (*net.IP, error) {
	addr := net.ParseIP(strings.TrimSpace(data))
	if addr == nil {
		return nil, fmt.Errorf("failed to parse IP address from %q", data)
	}
	return &addr, nil
}

func rawIPByURL(ctx context.Context, url string, headers map[string]string) (*net.IP, error) {
	data, err := RawStrByURL(ctx, url, headers)
	if err != nil {
		return nil, err
	}
	return parseIP(data)
}

type jsonStatus struct {
	path, value string
}

func jsonIPByURL(ctx context.Context, url, path string, headers map[string]string, statuses ...jsonStatus) (*net.IP, error) {
	data, err := RawStrByURL(ctx, url, headers)
	if err != nil {
		return nil, err
	}
	if !gjson.Valid(data) {
		return nil, fmt.Errorf("source returned invalid JSON")
	}
	for _, status := range statuses {
		result := gjson.Get(data, status.path)
		if !result.Exists() || result.String() != status.value {
			return nil, fmt.Errorf("source returned %s=%q, expected %q", status.path, result.String(), status.value)
		}
	}
	return parseIP(gjson.Get(data, path).Str)
}

// RawStrByURL performs an HTTP GET against url and returns the body as a
// string. Additional headers may be provided to override the defaults.
func RawStrByURL(ctx context.Context, url string, headers map[string]string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Referer", url)
	req.Header.Set("User-Agent", UserAgent)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP request failed with status %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	return string(body), err
}

func registerSource(name string, fn ddns.SourceFunc) {
	if err := ddns.RegisterSourceFunc(name, fn); err != nil {
		panic(err)
	}
}

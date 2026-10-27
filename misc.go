/*!*
 * Copyright (c) 2025-2026 Ming Lyu, aka mingcheng
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: misc.go
 * Author: mingcheng <mingcheng@apache.org>
 * File Created: 2022-07-22 23:37:43
 *
 * Modified By: mingcheng <mingcheng@apache.org>
 * Last Modified: 2026-05-12 12:23:04
 */

package simplyddns

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	tld "github.com/jpillora/go-tld"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/idna"
	"golang.org/x/net/proxy"
)

// ProxyHttpClient creates an http.Client backed by a SOCKS5 proxy at addr.
func ProxyHttpClient(addr string) (*http.Client, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid SOCKS5 proxy address: %w", err)
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || host == "" || portNumber < 1 || portNumber > 65535 {
		return nil, fmt.Errorf("SOCKS5 proxy requires a host and a port between 1 and 65535")
	}
	dialer, err := proxy.SOCKS5("tcp", addr, nil, &net.Dialer{Timeout: DefaultTimeout})
	if err != nil {
		return nil, err
	}

	contextDialer, ok := dialer.(proxy.ContextDialer)
	if !ok {
		return nil, fmt.Errorf("SOCKS5 proxy dialer does not support cancellation")
	}

	return &http.Client{
		Timeout: DefaultTimeout,
		Transport: &http.Transport{
			DialContext:           contextDialer.DialContext,
			TLSHandshakeTimeout:   DefaultTimeout,
			ResponseHeaderTimeout: DefaultTimeout,
			IdleConnTimeout:       DefaultTimeout,
			ForceAttemptHTTP2:     true,
		},
	}, nil
}

// ParseDomain parses a bare domain name into its components using go-tld.
// See https://github.com/jpillora/go-tld for details.
func ParseDomain(domain string) (*tld.URL, error) {
	name := strings.TrimSuffix(domain, ".")
	wildcard := strings.HasPrefix(name, "*.")
	name = strings.TrimPrefix(name, "*.")
	name, err := idna.Lookup.ToASCII(name)
	if err != nil {
		return nil, fmt.Errorf("invalid domain %q: %w", domain, err)
	}
	if name == "" || len(name) > 253 || net.ParseIP(name) != nil {
		return nil, fmt.Errorf("%q is not a valid domain", domain)
	}
	for _, label := range strings.Split(name, ".") {
		if label == "" || len(label) > 63 || strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return nil, fmt.Errorf("%q is not a valid domain", domain)
		}
	}
	if wildcard {
		name = "*." + name
		if len(name) > 253 {
			return nil, fmt.Errorf("%q is not a valid domain", domain)
		}
	}
	u, err := tld.Parse("http://" + name)
	if err != nil {
		return nil, err
	}
	if u.Domain == "" || u.TLD == "" {
		return nil, fmt.Errorf("%q is not a valid domain", domain)
	}

	return u, nil
}

var log = logrus.New()
var lookupIP = func(ctx context.Context, network, host string) ([]net.IP, error) {
	return net.DefaultResolver.LookupIP(ctx, network, host)
}

// NewLogger returns the shared logger instance.
func NewLogger() *logrus.Logger {
	return log
}

// ValidateRecords validates each domain in the list against the given address.
// It returns nil only when every domain already resolves to addr.
func ValidateRecords(domains []string, addr *net.IP) error {
	return ValidateRecordsContext(context.Background(), domains, addr)
}

// ValidateRecordsContext validates records with a cancellable DNS lookup.
func ValidateRecordsContext(ctx context.Context, domains []string, addr *net.IP) error {
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()
	for _, domain := range domains {
		if err := ValidateRecordContext(ctx, domain, addr); err != nil {
			return err
		}
	}

	return nil
}

// ValidateRecord checks whether the given domain currently resolves to addr.
func ValidateRecord(domain string, addr *net.IP) error {
	return ValidateRecordContext(context.Background(), domain, addr)
}

// ValidateRecordContext checks whether a domain resolves to addr.
func ValidateRecordContext(ctx context.Context, domain string, addr *net.IP) error {
	if addr == nil || addr.To16() == nil {
		return fmt.Errorf("invalid IP address")
	}
	if _, err := ParseDomain(domain); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()
	records, err := lookupIP(ctx, "ip", domain)
	if err != nil {
		return err
	}

	for _, record := range records {
		if record.Equal(*addr) {
			return nil
		}
	}

	return fmt.Errorf("domain %s does not point to address %s", domain, addr.String())
}

// ValidateConfig performs basic sanity checks on a JobConfig.
func ValidateConfig(config *JobConfig) error {
	if config == nil {
		return fmt.Errorf("configure is nil")
	}
	if strings.TrimSpace(config.Source.Type) == "" || strings.TrimSpace(config.Target.Type) == "" {
		return fmt.Errorf("source or target type cannot be empty")
	}
	if uint64(config.Source.Interval) > uint64((1<<63-1)/time.Second) {
		return fmt.Errorf("source interval is too large")
	}

	for _, domain := range config.Target.Domains {
		if _, err := ParseDomain(domain); err != nil {
			return err
		}
	}
	if config.WebHook.Url != "" {
		u, err := url.Parse(config.WebHook.Url)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
			return fmt.Errorf("webhook URL must be an absolute HTTP or HTTPS URL")
		}
	}

	return nil
}

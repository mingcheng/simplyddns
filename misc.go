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
	"fmt"
	"net"
	"net/http"
	"sync"

	tld "github.com/jpillora/go-tld"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
)

// ProxyHttpClient creates an http.Client backed by a SOCKS5 proxy at addr.
func ProxyHttpClient(addr string) (*http.Client, error) {
	// setup a http client
	httpTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}

	// create a socks5 dialer
	dialer, err := proxy.SOCKS5("tcp", addr, nil, proxy.Direct)
	if err != nil {
		return nil, err
	}

	// set our socks5 as the dialer
	if contextDialer, ok := dialer.(proxy.ContextDialer); ok {
		httpTransport.DialContext = contextDialer.DialContext
	}

	return &http.Client{
		Transport: httpTransport,
	}, nil
}

// ParseDomain parses a bare domain name into its components using go-tld.
// See https://github.com/jpillora/go-tld for details.
func ParseDomain(domain string) (*tld.URL, error) {
	u, err := tld.Parse(fmt.Sprintf("http://%s/foo", domain))
	if err != nil {
		return nil, err
	}

	if !u.ICANN && u.Domain == "" && u.TLD == "" {
		return nil, fmt.Errorf("%q is not a valid domain", domain)
	}

	return u, nil
}

var (
	log  *logrus.Logger
	once sync.Once
)

// NewLogger returns the shared logger instance, initialising it on first use.
func NewLogger() *logrus.Logger {
	once.Do(func() {
		log = logrus.New()
	})

	return log
}

// ValidateRecords validates each domain in the list against the given address.
// It returns nil only when every domain already resolves to addr.
func ValidateRecords(domains []string, addr *net.IP) error {
	for _, domain := range domains {
		if _, err := ParseDomain(domain); err != nil {
			return err
		}
		if err := ValidateRecord(domain, addr); err != nil {
			return err
		}
	}

	return nil
}

// ValidateRecord checks whether the given domain currently resolves to addr.
func ValidateRecord(domain string, addr *net.IP) error {
	records, err := net.LookupIP(domain)
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

	for _, domain := range config.Target.Domains {
		if _, err := ParseDomain(domain); err != nil {
			return err
		}
	}

	return nil
}

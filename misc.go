/**
 * File: misc.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Saturday, December 26th 2020, 7:21:58 pm
 * Last Modified: Sunday, December 27th 2020, 8:40:05 pm
 *
 * http://www.opensource.org/licenses/MIT
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

// ProxyHttpClient to create http client with socks5 proxy
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

// https://github.com/jpillora/go-tld
func ParseDomain(domain string) (*tld.URL, error) {
	if u, err := tld.Parse(fmt.Sprintf("http://%s/foo", domain)); err != nil {
		return nil, err
	} else {
		if !u.ICANN && (u.Domain == "" && u.TLD == "") {
			return nil, fmt.Errorf("%v is not a vaildate domain", domain)
		}

		return u, nil
	}
}

// NewLogger to return logger instance
var (
	log  *logrus.Logger
	once sync.Once
)

func NewLogger() *logrus.Logger {
	once.Do(func() {
		log = logrus.New()
	})

	return log
}

// func init() {
// 	log = NewLogger()
// }

// ValidateRecords 批量验证 DNS 域名是否已经是对应的 IP 地址
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

// ValidateRecord 批量验证 DNS 域名是否已经是对应的 IP 地址
func ValidateRecord(domain string, addr *net.IP) error {
	found := false

	if records, err := net.LookupIP(domain); err != nil {
		return err
	} else {
		for _, record := range records {
			if record.Equal(*addr) {
				found = true
			}
		}

		if found {
			return nil
		}
	}

	return fmt.Errorf("domain %s is not found address %s", domain, addr.String())
}

// ValidateConfig 验证配置对象是否合适
func ValidateConfig(config *JobConfig) error {
	if config == nil {
		return fmt.Errorf("configure is nil")
	}

	// check domain
	for _, domain := range config.Target.Domains {
		if _, err := ParseDomain(domain); err != nil {
			return err
		}
	}

	return nil
}

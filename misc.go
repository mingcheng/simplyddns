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
	"net/http"

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
var log *logrus.Logger

func NewLogger() *logrus.Logger {
	if log == nil {
		log = logrus.New()
	}

	return log
}

func init() {
	log = NewLogger()
}

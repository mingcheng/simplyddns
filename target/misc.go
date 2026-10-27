/**
 * File: misc.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Sunday, December 27th 2020, 7:01:07 pm
 * Last Modified: Sunday, December 27th 2020, 8:39:11 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package target

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"

	ddns "github.com/mingcheng/simplyddns"
)

var (
	log = ddns.NewLogger()
)

func recordType(addr *net.IP) (string, error) {
	if addr == nil || addr.To16() == nil {
		return "", fmt.Errorf("invalid IP address")
	}
	if addr.To4() != nil {
		return "A", nil
	}
	return "AAAA", nil
}

func newHTTPClient(config *ddns.TargetConfig) (*http.Client, error) {
	if config == nil {
		return nil, fmt.Errorf("target configuration is nil")
	}
	if config.Key == "" || config.Token == "" {
		return nil, fmt.Errorf("target key and token are required")
	}
	if config.Proxy != "" {
		return ddns.ProxyHttpClient(config.Proxy)
	}
	return &http.Client{Transport: http.DefaultTransport, Timeout: ddns.DefaultTimeout}, nil
}

// The Aliyun SDK lacks request contexts. Keep both its timeout and the job's
// cancellation active until the response body has been consumed.
type contextTransport struct {
	ctx  context.Context
	base http.RoundTripper
}

func (t contextTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if err := t.ctx.Err(); err != nil {
		if r.Body != nil {
			r.Body.Close()
		}
		return nil, err
	}
	ctx, cancel := context.WithCancelCause(r.Context())
	stop := context.AfterFunc(t.ctx, func() { cancel(context.Cause(t.ctx)) })
	cleanup := func() {
		stop()
		cancel(context.Canceled)
	}
	resp, err := t.base.RoundTrip(r.Clone(ctx))
	if err != nil {
		if cause := context.Cause(ctx); cause != nil {
			err = cause
		}
		cleanup()
		return nil, err
	}
	resp.Body = &contextBody{ReadCloser: resp.Body, cleanup: cleanup}
	return resp, nil
}

type contextBody struct {
	io.ReadCloser
	cleanup func()
}

func (b *contextBody) Close() error {
	defer b.cleanup()
	return b.ReadCloser.Close()
}

func registerTarget(name string, fn ddns.TargetFunc) {
	if err := ddns.RegisterTargetFunc(name, fn); err != nil {
		panic(err)
	}
}

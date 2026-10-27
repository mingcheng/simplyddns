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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	ddns "github.com/mingcheng/simplyddns"
)

var (
	log = ddns.NewLogger()
)

const defaultRecordTTL = 600

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
	return targetHTTPClient(config)
}

func targetHTTPClient(config *ddns.TargetConfig) (*http.Client, error) {
	if config == nil {
		return nil, fmt.Errorf("target configuration is nil")
	}
	if config.Proxy != "" {
		return ddns.ProxyHttpClient(config.Proxy)
	}
	return &http.Client{Transport: http.DefaultTransport, Timeout: ddns.DefaultTimeout}, nil
}

type dnsName struct {
	domain, zone, name string
}

func dnsNames(config *ddns.TargetConfig, zone string) ([]dnsName, error) {
	if config == nil || len(config.Domains) == 0 {
		return nil, fmt.Errorf("target requires at least one domain")
	}
	if zone != "" {
		parsed, err := ddns.ParseDomain(strings.TrimSuffix(zone, "."))
		if err != nil {
			return nil, fmt.Errorf("invalid DNS zone: %w", err)
		}
		zone = parsed.Hostname()
	}
	names := make([]dnsName, 0, len(config.Domains))
	seen := make(map[string]bool)
	for _, domain := range config.Domains {
		parsed, err := ddns.ParseDomain(domain)
		if err != nil {
			return nil, err
		}
		currentZone := zone
		if currentZone == "" {
			currentZone = parsed.Domain + "." + parsed.TLD
		}
		host := parsed.Hostname()
		name := "@"
		if host != currentZone {
			if !strings.HasSuffix(host, "."+currentZone) {
				return nil, fmt.Errorf("domain %s does not belong to zone %s", host, currentZone)
			}
			name = strings.TrimSuffix(host, "."+currentZone)
		}
		if seen[host] {
			continue
		}
		seen[host] = true
		names = append(names, dnsName{domain: host, zone: currentZone, name: name})
	}
	return names, nil
}

func requestJSON(ctx context.Context, client *http.Client, method, endpoint string, headers map[string]string, body, result any, status int) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", ddns.DefaultUserAgent)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != status {
		return fmt.Errorf("HTTP request failed with status %s (expected %d)", resp.Status, status)
	}
	if result == nil {
		return nil
	}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(result); err != nil {
		return fmt.Errorf("invalid JSON response: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return fmt.Errorf("invalid JSON response: %w", err)
		}
		return fmt.Errorf("response contains multiple JSON values")
	}
	return nil
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

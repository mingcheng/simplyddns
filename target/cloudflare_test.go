/*!*
 * Copyright (c) 2025-2026 Ming Lyu, aka mingcheng
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: cloudflare_test.go
 * Author: mingcheng <mingcheng@apache.org>
 * File Created: Monday, December 28th 2020, 11:10:47 am
 *
 * Modified By: mingcheng <mingcheng@apache.org>
 * Last Modified: 2026-05-12 12:24:11
 */

package target

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"testing"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/mingcheng/simplyddns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCloudflareDNSClient(t *testing.T) {
	mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("constructing a client must not make an extra API request")
	})
	client, err := NewCloudflareDNSClient(&simplyddns.TargetConfig{Key: "zone", Token: "token"})
	require.NoError(t, err)
	assert.Equal(t, "zone", client.zoneIdentifier.Identifier)
	assert.Equal(t, simplyddns.DefaultTimeout, client.httpClient.Timeout)
	_, err = NewCloudflareDNSClient(nil)
	require.Error(t, err)
}

func TestCloudflareCreateOrUpdateRecord(t *testing.T) {
	for _, tc := range []struct {
		name, address, old, kind, writeMethod string
	}{
		{"update IPv4", "192.0.2.1", "192.0.2.2", "A", http.MethodPatch},
		{"update IPv6", "2001:db8::1", "2001:db8::2", "AAAA", http.MethodPatch},
		{"create IPv4", "192.0.2.1", "", "A", http.MethodPost},
		{"create IPv6", "2001:db8::1", "", "AAAA", http.MethodPost},
		{"unchanged", "2001:db8::1", "2001:db8:0:0::1", "AAAA", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var methods []string
			proxied := true
			mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
				methods = append(methods, r.Method)
				assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))
				if r.Method == http.MethodGet {
					assert.Equal(t, tc.kind, r.URL.Query().Get("type"))
					assert.Equal(t, "home.example.com", r.URL.Query().Get("name"))
					records := []cloudflare.DNSRecord{}
					if tc.old != "" {
						records = append(records, cloudflare.DNSRecord{
							ID: "record", Name: "home.example.com", Type: tc.kind, Content: tc.old,
							TTL: 3600, Proxied: &proxied, Comment: "keep me", Tags: []string{"managed"},
						})
					}
					writeJSON(t, w, http.StatusOK, map[string]any{
						"success": true, "result": records,
						"result_info": map[string]int{"page": 1, "total_pages": 1},
					})
					return
				}
				assert.Equal(t, tc.writeMethod, r.Method, "updates must never delete records")
				var body map[string]any
				assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				assert.Equal(t, tc.address, body["content"])
				if r.Method == http.MethodPatch {
					assert.Equal(t, "/client/v4/zones/zone/dns_records/record", r.URL.Path)
					assert.Equal(t, float64(3600), body["ttl"])
					assert.Equal(t, true, body["proxied"])
					assert.Equal(t, "keep me", body["comment"])
					assert.Equal(t, []any{"managed"}, body["tags"])
				} else {
					assert.Equal(t, tc.kind, body["type"])
					assert.Equal(t, "home.example.com", body["name"])
				}
				writeJSON(t, w, http.StatusOK, map[string]any{"success": true, "result": map[string]string{"id": "record"}})
			})
			client, err := NewCloudflareDNSClient(&simplyddns.TargetConfig{Key: "zone", Token: "token"})
			require.NoError(t, err)
			ip := net.ParseIP(tc.address)
			require.NoError(t, client.CreateOrUpdateRecord(context.Background(), "home.example.com", &ip))
			want := []string{http.MethodGet}
			if tc.writeMethod != "" {
				want = append(want, tc.writeMethod)
			}
			assert.Equal(t, want, methods)
		})
	}
}

func TestCloudflareDoesNotCreateAfterLookupFailure(t *testing.T) {
	calls := 0
	mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		assert.Equal(t, http.MethodGet, r.Method)
		writeJSON(t, w, http.StatusForbidden, map[string]any{
			"success": false, "errors": []map[string]any{{"code": 10000, "message": "denied"}},
		})
	})
	fn, err := simplyddns.TargetFuncByName("cloudflare")
	require.NoError(t, err)
	ip := net.ParseIP("192.0.2.1")
	err = fn(context.Background(), &ip, &simplyddns.TargetConfig{
		Key: "zone", Token: "token", Domains: []string{"a.example.com", "b.example.com"},
	})
	require.Error(t, err)
	assert.Equal(t, 1, calls)
}

func TestCloudflareUpdatesAllPages(t *testing.T) {
	var updated []string
	mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			page := r.URL.Query().Get("page")
			writeJSON(t, w, http.StatusOK, map[string]any{
				"success": true,
				"result": []map[string]any{{
					"id": "record-" + page, "name": "home.example.com", "type": "A", "content": "192.0.2.2", "ttl": 60,
				}},
				"result_info": map[string]any{"page": json.Number(page), "total_pages": 2, "total_count": 2, "per_page": 1},
			})
			return
		}
		assert.Equal(t, http.MethodPatch, r.Method)
		updated = append(updated, r.URL.Path)
		writeJSON(t, w, http.StatusOK, map[string]any{"success": true, "result": map[string]string{"id": "record"}})
	})
	client, err := NewCloudflareDNSClient(&simplyddns.TargetConfig{Key: "zone", Token: "token"})
	require.NoError(t, err)
	ip := net.ParseIP("192.0.2.1")
	require.NoError(t, client.CreateOrUpdateRecord(context.Background(), "home.example.com", &ip))
	assert.Equal(t, []string{
		"/client/v4/zones/zone/dns_records/record-1",
		"/client/v4/zones/zone/dns_records/record-2",
	}, updated)
}

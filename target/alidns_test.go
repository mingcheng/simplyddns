/*!*
 * Copyright (c) 2025-2026 Ming Lyu, aka mingcheng
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: alidns_test.go
 * Author: mingcheng <mingcheng@apache.org>
 * File Created: Monday, December 28th 2020, 11:10:47 am
 *
 * Modified By: mingcheng <mingcheng@apache.org>
 * Last Modified: 2026-05-12 12:24:34
 */

package target

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"testing"

	"github.com/mingcheng/simplyddns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAliDNS(t *testing.T) {
	mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("constructing a client must not fetch every domain")
	})
	client, err := NewAliDNS(&simplyddns.TargetConfig{Key: "key", Token: "token"})
	require.NoError(t, err)
	assert.NotNil(t, client.Client)
	_, err = NewAliDNS(nil)
	require.Error(t, err)
}

func TestAliDNSCreateAndUpdate(t *testing.T) {
	for _, tc := range []struct {
		name, domain, host, address, kind string
		existing                          bool
	}{
		{"create IPv4 apex", "example.com", "@", "192.0.2.1", "A", false},
		{"create IPv6", "home.example.com", "home", "2001:db8::1", "AAAA", false},
		{"update IPv4", "home.example.com", "home", "192.0.2.1", "A", true},
		{"update IPv6 apex", "example.com", "@", "2001:db8::1", "AAAA", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var actions []string
			mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
				assert.NoError(t, r.ParseForm())
				assert.Equal(t, "https", r.URL.Scheme)
				action := r.Form.Get("Action")
				actions = append(actions, action)
				assert.Equal(t, tc.kind, r.Form.Get("Type"))
				if action == "DescribeSubDomainRecords" {
					assert.Equal(t, tc.domain, r.Form.Get("SubDomain"))
					assert.Equal(t, "example.com", r.Form.Get("DomainName"))
					records := []map[string]any{}
					if tc.existing {
						records = append(records, map[string]any{
							"RecordId": "record", "Type": tc.kind, "Value": "192.0.2.2", "TTL": 3600, "Line": "telecom",
						})
					}
					writeJSON(t, w, http.StatusOK, map[string]any{
						"TotalCount": len(records), "DomainRecords": map[string]any{"Record": records},
					})
					return
				}
				assert.Equal(t, tc.host, r.Form.Get("RR"))
				assert.Equal(t, tc.address, r.Form.Get("Value"))
				if tc.existing {
					assert.Equal(t, "UpdateDomainRecord", action)
					assert.Equal(t, "record", r.Form.Get("RecordId"))
					assert.Equal(t, "3600", r.Form.Get("TTL"))
					assert.Equal(t, "telecom", r.Form.Get("Line"))
				} else {
					assert.Equal(t, "AddDomainRecord", action)
					assert.Equal(t, "example.com", r.Form.Get("DomainName"))
				}
				writeJSON(t, w, http.StatusOK, map[string]string{"RecordId": "record"})
			})
			fn, err := simplyddns.TargetFuncByName("alidns")
			require.NoError(t, err)
			ip := net.ParseIP(tc.address)
			require.NoError(t, fn(context.Background(), &ip, &simplyddns.TargetConfig{
				Key: "key", Token: "token", Domains: []string{tc.domain},
			}))
			assert.Len(t, actions, 2)
		})
	}
}

func TestAliDNSErrorsAreNotOverwritten(t *testing.T) {
	for _, failAction := range []string{"DescribeSubDomainRecords", "UpdateDomainRecord", "AddDomainRecord"} {
		t.Run(failAction, func(t *testing.T) {
			calls := 0
			mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				assert.NoError(t, r.ParseForm())
				action := r.Form.Get("Action")
				if action == failAction {
					writeJSON(t, w, http.StatusForbidden, map[string]string{"Code": "Forbidden", "Message": "denied"})
					return
				}
				assert.Equal(t, "DescribeSubDomainRecords", action)
				records := []map[string]any{}
				if failAction == "UpdateDomainRecord" {
					records = append(records, map[string]any{"RecordId": "record", "Type": "A", "Value": "192.0.2.2"})
				}
				writeJSON(t, w, http.StatusOK, map[string]any{
					"TotalCount": len(records), "DomainRecords": map[string]any{"Record": records},
				})
			})
			fn, err := simplyddns.TargetFuncByName("alidns")
			require.NoError(t, err)
			ip := net.ParseIP("192.0.2.1")
			err = fn(context.Background(), &ip, &simplyddns.TargetConfig{
				Key: "key", Token: "token", Domains: []string{"a.example.com", "b.example.com"},
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), "a.example.com")
			if failAction == "DescribeSubDomainRecords" {
				assert.Equal(t, 1, calls)
			} else {
				assert.Equal(t, 2, calls)
			}
		})
	}
}

func TestAliDNSAllPages(t *testing.T) {
	pages := 0
	mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
		pages++
		assert.NoError(t, r.ParseForm())
		assert.Equal(t, "DescribeDomains", r.Form.Get("Action"))
		assert.Equal(t, "100", r.Form.Get("PageSize"))
		domain := "example.com"
		if pages == 2 {
			domain = "example.org"
		}
		writeJSON(t, w, http.StatusOK, map[string]any{
			"TotalCount": 2, "Domains": map[string]any{"Domain": []map[string]string{{"DomainName": domain}}},
		})
		assert.Equal(t, strconv.Itoa(pages), r.Form.Get("PageNumber"))
	})
	client, err := NewAliDNS(&simplyddns.TargetConfig{Key: "key", Token: "token"})
	require.NoError(t, err)
	domains, err := client.All()
	require.NoError(t, err)
	require.Len(t, domains, 2)
	assert.Equal(t, "example.org", domains[1].DomainName)
	assert.Equal(t, 2, pages)
}

func TestAliDNSRejectsInconsistentRecords(t *testing.T) {
	calls := 0
	mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		writeJSON(t, w, http.StatusOK, map[string]any{
			"TotalCount": 1, "DomainRecords": map[string]any{"Record": []any{}},
		})
	})
	fn, err := simplyddns.TargetFuncByName("alidns")
	require.NoError(t, err)
	ip := net.ParseIP("192.0.2.1")
	require.Error(t, fn(context.Background(), &ip, &simplyddns.TargetConfig{
		Key: "key", Token: "token", Domains: []string{"home.example.com"},
	}))
	assert.Equal(t, 1, calls)
}

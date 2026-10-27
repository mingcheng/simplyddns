package target

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"testing"

	ddns "github.com/mingcheng/simplyddns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func huaweiTestConfig() ddns.TargetConfig {
	return ddns.TargetConfig{
		Key: "key", Token: "token", Region: "cn-north-4", ProjectID: "project", SessionToken: "session",
		Zone: "zone", Domains: []string{"home.example.com"},
	}
}

func TestHuaweiCloudRecords(t *testing.T) {
	for _, tc := range []struct{ name, address, old, kind string }{
		{"create IPv4", "192.0.2.1", "", "A"},
		{"create IPv6", "2001:db8::1", "", "AAAA"},
		{"update IPv4", "192.0.2.1", "192.0.2.2", "A"},
		{"update IPv6", "2001:db8::1", "2001:db8::2", "AAAA"},
		{"unchanged", "2001:db8::1", "2001:db8:0:0::1", "AAAA"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			writes := 0
			mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "dns.cn-north-4.myhuaweicloud.com", r.URL.Host)
				assert.True(t, strings.HasPrefix(r.Header.Get("Authorization"), "SDK-HMAC-SHA256 Access=key,"))
				assert.Equal(t, "project", r.Header.Get("X-Project-Id"))
				assert.Equal(t, "session", r.Header.Get("X-Security-Token"))
				if r.URL.Path == "/v2/zones/zone" {
					writeJSON(t, w, http.StatusOK, map[string]string{"id": "zone", "name": "example.com."})
					return
				}
				if r.Method == http.MethodGet {
					assert.Equal(t, "home.example.com.", r.URL.Query().Get("name"))
					assert.Equal(t, "equal", r.URL.Query().Get("search_mode"))
					assert.Equal(t, tc.kind, r.URL.Query().Get("type"))
					records := []map[string]any{}
					if tc.old != "" {
						records = append(records, map[string]any{
							"id": "record", "name": "home.example.com.", "type": tc.kind, "ttl": 3600,
							"records": []string{tc.old}, "description": "keep", "status": "DISABLE",
						})
					}
					writeJSON(t, w, http.StatusOK, map[string]any{
						"recordsets": records, "metadata": map[string]int{"total_count": len(records)},
					})
					return
				}
				writes++
				var body map[string]any
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				assert.Equal(t, []any{tc.address}, body["records"])
				if tc.old == "" {
					assert.Equal(t, http.MethodPost, r.Method)
					assert.Equal(t, "/v2/zones/zone/recordsets", r.URL.Path)
					assert.Equal(t, "home.example.com.", body["name"])
					assert.Equal(t, tc.kind, body["type"])
					assert.EqualValues(t, defaultRecordTTL, body["ttl"])
					writeJSON(t, w, http.StatusCreated, map[string]string{"id": "record", "status": "PENDING_CREATE"})
				} else {
					assert.Equal(t, http.MethodPut, r.Method)
					assert.Equal(t, "/v2/zones/zone/recordsets/record", r.URL.Path)
					assert.EqualValues(t, 3600, body["ttl"])
					assert.Equal(t, "keep", body["description"])
					assert.NotContains(t, body, "status")
					writeJSON(t, w, http.StatusOK, map[string]string{"id": "record", "status": "PENDING_UPDATE"})
				}
			})
			config := huaweiTestConfig()
			ip := net.ParseIP(tc.address)
			require.NoError(t, updateHuaweiCloud(context.Background(), &ip, &config))
			if net.ParseIP(tc.old).Equal(ip) {
				assert.Zero(t, writes)
			} else {
				assert.Equal(t, 1, writes)
			}
		})
	}
}

func TestHuaweiCloudPagination(t *testing.T) {
	var markers, updates []string
	mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/zones/zone" {
			writeJSON(t, w, http.StatusOK, map[string]string{"name": "example.com."})
			return
		}
		if r.Method == http.MethodGet {
			marker := r.URL.Query().Get("marker")
			markers = append(markers, marker)
			id := "first"
			if marker != "" {
				id = "second"
			}
			writeJSON(t, w, http.StatusOK, map[string]any{
				"recordsets": []map[string]any{{
					"id": id, "name": "home.example.com.", "type": "A", "ttl": 600, "records": []string{"192.0.2.2"},
				}},
				"metadata": map[string]int{"total_count": 2},
			})
			return
		}
		assert.Equal(t, http.MethodPut, r.Method)
		id := strings.TrimPrefix(r.URL.Path, "/v2/zones/zone/recordsets/")
		updates = append(updates, id)
		writeJSON(t, w, http.StatusOK, map[string]string{"id": id})
	})
	config := huaweiTestConfig()
	ip := net.ParseIP("192.0.2.1")
	require.NoError(t, updateHuaweiCloud(context.Background(), &ip, &config))
	assert.Equal(t, []string{"", "first"}, markers)
	assert.Equal(t, []string{"first", "second"}, updates)
}

func TestHuaweiCloudErrorsAndCancellation(t *testing.T) {
	for _, name := range []string{"lookup denied", "missing metadata", "truncated list", "missing create ID", "write failure", "outside zone", "failed create state"} {
		t.Run(name, func(t *testing.T) {
			config := huaweiTestConfig()
			if name == "outside zone" {
				config.Domains = []string{"badexample.com"}
			}
			mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/v2/zones/zone" {
					writeJSON(t, w, http.StatusOK, map[string]string{"name": "example.com."})
					return
				}
				if name != "write failure" && name != "missing create ID" && name != "failed create state" {
					assert.Equal(t, http.MethodGet, r.Method, "lookup failures must not cause writes")
				}
				if name == "lookup denied" || (name == "write failure" && r.Method == http.MethodPost) {
					writeJSON(t, w, http.StatusForbidden, map[string]string{"error_code": "DNS.0001", "error_msg": "denied"})
					return
				}
				if r.Method == http.MethodPost {
					body := map[string]string{}
					if name == "failed create state" {
						body = map[string]string{"id": "record", "status": "ERROR"}
					}
					writeJSON(t, w, http.StatusCreated, body)
					return
				}
				body := map[string]any{"recordsets": []any{}, "metadata": map[string]int{"total_count": 0}}
				if name == "missing metadata" {
					delete(body, "metadata")
				} else if name == "truncated list" {
					body["metadata"] = map[string]int{"total_count": 1}
				}
				writeJSON(t, w, http.StatusOK, body)
			})
			ip := net.ParseIP("192.0.2.1")
			require.Error(t, updateHuaweiCloud(context.Background(), &ip, &config))
		})
	}
	t.Run("invalid region", func(t *testing.T) {
		mockAPI(t, func(http.ResponseWriter, *http.Request) { t.Error("invalid region must not make requests") })
		config := huaweiTestConfig()
		config.Region = "unknown-region"
		ip := net.ParseIP("192.0.2.1")
		require.Error(t, updateHuaweiCloud(context.Background(), &ip, &config))
	})
	t.Run("cancellation", func(t *testing.T) {
		testTargetCancellation(t, "huaweicloud", huaweiTestConfig())
	})
}

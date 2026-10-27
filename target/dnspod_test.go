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

func TestDNSPodRecords(t *testing.T) {
	for _, tc := range []struct{ name, address, old, kind string }{
		{"create IPv4", "192.0.2.1", "", "A"},
		{"create IPv6", "2001:db8::1", "", "AAAA"},
		{"update IPv4", "192.0.2.1", "192.0.2.2", "A"},
		{"update IPv6", "2001:db8::1", "2001:db8::2", "AAAA"},
		{"unchanged", "2001:db8::1", "2001:db8:0:0::1", "AAAA"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var actions []string
			mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "dnspod.tencentcloudapi.com", r.URL.Host)
				assert.Equal(t, http.MethodPost, r.Method)
				assert.True(t, strings.HasPrefix(r.Header.Get("Authorization"), "TC3-HMAC-SHA256 Credential=key/"))
				assert.Equal(t, "session", r.Header.Get("X-TC-Token"))
				var body map[string]any
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				assert.Equal(t, "example.com", body["Domain"])
				assert.Equal(t, "home", body["SubDomain"])
				assert.Equal(t, tc.kind, body["RecordType"])
				action := r.Header.Get("X-TC-Action")
				actions = append(actions, action)
				if action == "DescribeRecordList" {
					assert.Equal(t, "no", body["ErrorOnEmpty"])
					records := []map[string]any{}
					if tc.old != "" {
						records = append(records, map[string]any{
							"RecordId": 123, "Name": "home", "Type": tc.kind, "Value": tc.old,
							"Line": "default", "LineId": "0", "TTL": 3600, "Status": "DISABLE", "Weight": 5, "Remark": "keep",
						})
					}
					writeJSON(t, w, http.StatusOK, map[string]any{"Response": map[string]any{
						"RecordCountInfo": map[string]int{"TotalCount": len(records)}, "RecordList": records,
					}})
					return
				}
				assert.Equal(t, tc.address, body["Value"])
				if tc.old == "" {
					assert.Equal(t, "CreateRecord", action)
					assert.Equal(t, "0", body["RecordLineId"])
					assert.EqualValues(t, defaultRecordTTL, body["TTL"])
				} else {
					assert.Equal(t, "ModifyRecord", action)
					assert.EqualValues(t, 123, body["RecordId"])
					assert.EqualValues(t, 3600, body["TTL"])
					assert.Equal(t, "DISABLE", body["Status"])
					assert.Equal(t, "keep", body["Remark"])
					assert.EqualValues(t, 5, body["Weight"])
				}
				writeJSON(t, w, http.StatusOK, map[string]any{"Response": map[string]any{"RecordId": 123}})
			})
			ip := net.ParseIP(tc.address)
			require.NoError(t, updateDNSPod(context.Background(), &ip, &ddns.TargetConfig{
				Key: "key", Token: "token", SessionToken: "session", Domains: []string{"home.example.com"},
			}))
			if net.ParseIP(tc.old).Equal(ip) {
				assert.Equal(t, []string{"DescribeRecordList"}, actions)
			} else {
				assert.Len(t, actions, 2)
			}
		})
	}
}

func TestDNSPodPagination(t *testing.T) {
	var offsets, updated []float64
	mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		if r.Header.Get("X-TC-Action") == "DescribeRecordList" {
			offset := body["Offset"].(float64)
			offsets = append(offsets, offset)
			writeJSON(t, w, http.StatusOK, map[string]any{"Response": map[string]any{
				"RecordCountInfo": map[string]int{"TotalCount": 2},
				"RecordList": []map[string]any{{
					"RecordId": offset + 1, "Name": "home", "Type": "A", "Value": "192.0.2.2",
					"Line": "default", "LineId": "0", "TTL": 600, "Status": "ENABLE",
				}},
			}})
			return
		}
		assert.Equal(t, "ModifyRecord", r.Header.Get("X-TC-Action"))
		updated = append(updated, body["RecordId"].(float64))
		writeJSON(t, w, http.StatusOK, map[string]any{"Response": map[string]any{"RecordId": body["RecordId"]}})
	})
	ip := net.ParseIP("192.0.2.1")
	require.NoError(t, updateDNSPod(context.Background(), &ip, &ddns.TargetConfig{
		Key: "key", Token: "token", Domains: []string{"home.example.com"},
	}))
	assert.Equal(t, []float64{0, 1}, offsets)
	assert.Equal(t, []float64{1, 2}, updated)
}

func TestDNSPodErrorsAndCancellation(t *testing.T) {
	for _, name := range []string{"API failure", "missing count", "truncated list", "missing create ID", "write failure"} {
		t.Run(name, func(t *testing.T) {
			mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
				action := r.Header.Get("X-TC-Action")
				response := map[string]any{
					"RecordCountInfo": map[string]int{"TotalCount": 0}, "RecordList": []any{},
				}
				if name == "API failure" || (name == "write failure" && action == "CreateRecord") {
					response = map[string]any{"Error": map[string]string{"Code": "AuthFailure", "Message": "denied"}, "RequestId": "request"}
				} else if name == "missing count" || action == "CreateRecord" {
					response = map[string]any{}
				} else if name == "truncated list" {
					response["RecordCountInfo"] = map[string]int{"TotalCount": 1}
				}
				if name != "write failure" && name != "missing create ID" {
					assert.Equal(t, "DescribeRecordList", action)
				}
				writeJSON(t, w, http.StatusOK, map[string]any{"Response": response})
			})
			ip := net.ParseIP("192.0.2.1")
			require.Error(t, updateDNSPod(context.Background(), &ip, &ddns.TargetConfig{
				Key: "key", Token: "token", Domains: []string{"home.example.com", "other.example.com"},
			}))
		})
	}
	t.Run("cancellation", func(t *testing.T) {
		testTargetCancellation(t, "dnspod", ddns.TargetConfig{Key: "key", Token: "token", Domains: []string{"home.example.com"}})
	})
}

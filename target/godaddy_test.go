package target

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"

	ddns "github.com/mingcheng/simplyddns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoDaddyRecords(t *testing.T) {
	for _, tc := range []struct {
		name, domain, host, address, old, kind string
	}{
		{"create IPv4", "home.example.com", "home", "192.0.2.1", "", "A"},
		{"create IPv6", "home.example.com", "home", "2001:db8::1", "", "AAAA"},
		{"update IPv4", "home.example.com", "home", "192.0.2.1", "192.0.2.2", "A"},
		{"update IPv6", "home.example.com", "home", "2001:db8::1", "2001:db8::2", "AAAA"},
		{"unchanged", "home.example.com", "home", "2001:db8::1", "2001:db8:0:0::1", "AAAA"},
		{"apex", "example.com", "@", "192.0.2.1", "", "A"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			writes := 0
			mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "api.godaddy.com", r.URL.Host)
				assert.Equal(t, "sso-key key:token", r.Header.Get("Authorization"))
				assert.Equal(t, "/v1/domains/example.com/records/"+tc.kind+"/"+tc.host, r.URL.Path)
				if r.Method == http.MethodGet {
					assert.Equal(t, "100", r.URL.Query().Get("limit"))
					records := []goDaddyRecord{}
					if tc.old != "" {
						records = append(records, goDaddyRecord{Type: tc.kind, Name: tc.host, Data: tc.old, TTL: 3600})
					}
					writeJSON(t, w, http.StatusOK, records)
					return
				}
				writes++
				assert.Equal(t, http.MethodPut, r.Method)
				var records []goDaddyRecord
				require.NoError(t, json.NewDecoder(r.Body).Decode(&records))
				require.Len(t, records, 1)
				assert.Equal(t, tc.address, records[0].Data)
				assert.Empty(t, records[0].Type)
				assert.Empty(t, records[0].Name)
				ttl := int64(defaultRecordTTL)
				if tc.old != "" {
					ttl = 3600
				}
				assert.Equal(t, ttl, records[0].TTL)
				w.WriteHeader(http.StatusOK)
			})
			fn, err := ddns.TargetFuncByName("godaddy")
			require.NoError(t, err)
			ip := net.ParseIP(tc.address)
			require.NoError(t, fn(context.Background(), &ip, &ddns.TargetConfig{
				Key: "key", Token: "token", Domains: []string{tc.domain},
			}))
			if net.ParseIP(tc.old).Equal(ip) {
				assert.Zero(t, writes)
			} else {
				assert.Equal(t, 1, writes)
			}
		})
	}
}

func TestGoDaddyLookupFailures(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		status     int
	}{
		{"denied", `{"message":"denied"}`, http.StatusForbidden},
		{"not found", `{"message":"missing domain"}`, http.StatusNotFound},
		{"null", `null`, http.StatusOK},
		{"malformed", `[`, http.StatusOK},
		{"incomplete", `[{"data":"192.0.2.2"}]`, http.StatusOK},
		{"wrong name", `[{"type":"A","name":"other","data":"192.0.2.1","ttl":600}]`, http.StatusOK},
		{"wrong type", `[{"type":"AAAA","name":"home","data":"2001:db8::1","ttl":600}]`, http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method, "lookup failures must not cause writes")
				w.WriteHeader(tc.status)
				_, err := fmt.Fprint(w, tc.body)
				assert.NoError(t, err)
			})
			ip := net.ParseIP("192.0.2.1")
			require.Error(t, updateGoDaddy(context.Background(), &ip, &ddns.TargetConfig{
				Key: "key", Token: "token", Domains: []string{"home.example.com"},
			}))
		})
	}
}

func TestGoDaddyPaginationAndWriteFailure(t *testing.T) {
	var offsets []string
	mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			offset := r.URL.Query().Get("offset")
			offsets = append(offsets, offset)
			records := []goDaddyRecord{{Type: "A", Name: "home", Data: "192.0.2.200", TTL: 600}}
			if offset == "0" {
				records = nil
				for i := 1; i <= 100; i++ {
					records = append(records, goDaddyRecord{Type: "A", Name: "home", Data: fmt.Sprintf("192.0.2.%d", i), TTL: 600})
				}
			}
			writeJSON(t, w, http.StatusOK, records)
			return
		}
		assert.Equal(t, http.MethodPut, r.Method)
		writeJSON(t, w, http.StatusForbidden, map[string]string{"message": "denied"})
	})
	ip := net.ParseIP("198.51.100.1")
	require.ErrorContains(t, updateGoDaddy(context.Background(), &ip, &ddns.TargetConfig{
		Key: "key", Token: "token", Domains: []string{"home.example.com", "other.example.com"},
	}), "403")
	assert.Equal(t, []string{"0", "100"}, offsets)
}

func TestGoDaddyCancellation(t *testing.T) {
	testTargetCancellation(t, "godaddy", ddns.TargetConfig{Key: "key", Token: "token", Domains: []string{"home.example.com"}})
}

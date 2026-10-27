package target

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"testing"

	ddns "github.com/mingcheng/simplyddns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func powerDNSTestConfig() ddns.TargetConfig {
	return ddns.TargetConfig{
		Token: "api-key", Zone: "example.com.", Endpoint: "https://dns.example.net/api/v1/servers/localhost",
		Domains: []string{"home.example.com"},
	}
}

func TestPowerDNSRecords(t *testing.T) {
	for _, tc := range []struct{ name, address, old, kind string }{
		{"create IPv4", "192.0.2.1", "", "A"},
		{"create IPv6", "2001:db8::1", "", "AAAA"},
		{"update IPv4", "192.0.2.1", "192.0.2.2", "A"},
		{"update IPv6", "2001:db8::1", "2001:db8::2", "AAAA"},
		{"unchanged", "2001:db8::1", "2001:db8:0:0::1", "AAAA"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			writes := 0
			comments := []powerDNSComment{{Content: "keep me", Account: "owner", ModifiedAt: 123}}
			mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "dns.example.net", r.URL.Host)
				assert.Equal(t, "/api/v1/servers/localhost/zones/example.com.", r.URL.Path)
				assert.Equal(t, "api-key", r.Header.Get("X-API-Key"))
				if r.Method == http.MethodGet {
					zone := powerDNSZone{Name: "example.com.", RRSets: []powerDNSRecordSet{{
						Name: "other.example.com.", Type: "TXT", TTL: 600, Records: []powerDNSRecord{{Content: `"keep"`}},
					}}}
					if tc.old != "" {
						zone.RRSets = append(zone.RRSets, powerDNSRecordSet{
							Name: "home.example.com.", Type: tc.kind, TTL: 3600,
							Records: []powerDNSRecord{{Content: tc.old, Disabled: true}}, Comments: comments,
						})
					}
					writeJSON(t, w, http.StatusOK, zone)
					return
				}
				writes++
				assert.Equal(t, http.MethodPatch, r.Method)
				var body powerDNSZone
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				require.Len(t, body.RRSets, 1, "unrelated RRsets must not be sent")
				record := body.RRSets[0]
				assert.Equal(t, "REPLACE", record.ChangeType)
				assert.Equal(t, "home.example.com.", record.Name)
				assert.Equal(t, tc.kind, record.Type)
				require.Len(t, record.Records, 1)
				assert.Equal(t, tc.address, record.Records[0].Content)
				if tc.old != "" {
					assert.EqualValues(t, 3600, record.TTL)
					assert.True(t, record.Records[0].Disabled)
					assert.Equal(t, comments, record.Comments)
				} else {
					assert.EqualValues(t, defaultRecordTTL, record.TTL)
					assert.False(t, record.Records[0].Disabled)
				}
				w.WriteHeader(http.StatusNoContent)
			})
			ip := net.ParseIP(tc.address)
			config := powerDNSTestConfig()
			require.NoError(t, updatePowerDNS(context.Background(), &ip, &config))
			if net.ParseIP(tc.old).Equal(ip) {
				assert.Zero(t, writes)
			} else {
				assert.Equal(t, 1, writes)
			}
		})
	}
}

func TestPowerDNSRejectsUnsafeOrInvalidUpdates(t *testing.T) {
	for _, name := range []string{"outside zone", "missing zone", "missing rrset type", "mixed disabled states", "lookup denied", "redirect"} {
		t.Run(name, func(t *testing.T) {
			config := powerDNSTestConfig()
			if name == "outside zone" {
				config.Domains = []string{"badexample.com"}
			}
			mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "dns.example.net", r.URL.Host, "API keys must not follow redirects")
				assert.Equal(t, http.MethodGet, r.Method, "invalid responses must not cause writes")
				zone := powerDNSZone{Name: "example.com.", RRSets: []powerDNSRecordSet{}}
				switch name {
				case "missing zone":
					zone.Name = ""
				case "missing rrset type":
					zone.RRSets = append(zone.RRSets, powerDNSRecordSet{Name: "home.example.com."})
				case "mixed disabled states":
					zone.RRSets = append(zone.RRSets, powerDNSRecordSet{
						Name: "home.example.com.", Type: "A", TTL: 600,
						Records: []powerDNSRecord{{Content: "192.0.2.2"}, {Content: "192.0.2.3", Disabled: true}},
					})
				case "lookup denied":
					w.WriteHeader(http.StatusForbidden)
					return
				case "redirect":
					w.Header().Set("Location", "https://other.example.net/")
					w.WriteHeader(http.StatusFound)
					return
				}
				writeJSON(t, w, http.StatusOK, zone)
			})
			ip := net.ParseIP("192.0.2.1")
			require.Error(t, updatePowerDNS(context.Background(), &ip, &config))
		})
	}
}

func TestPowerDNSBatchAndWriteFailure(t *testing.T) {
	for _, status := range []int{http.StatusNoContent, http.StatusUnprocessableEntity} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			writes := 0
			mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					writeJSON(t, w, http.StatusOK, powerDNSZone{Name: "example.com.", RRSets: []powerDNSRecordSet{}})
					return
				}
				writes++
				var body powerDNSZone
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				require.Len(t, body.RRSets, 2)
				assert.Equal(t, "home.example.com.", body.RRSets[0].Name)
				assert.Equal(t, "example.com.", body.RRSets[1].Name)
				w.WriteHeader(status)
			})
			ip := net.ParseIP("192.0.2.1")
			config := powerDNSTestConfig()
			config.Domains = []string{"home.example.com", "example.com", "HOME.example.com"}
			err := updatePowerDNS(context.Background(), &ip, &config)
			if status == http.StatusNoContent {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, "422")
			}
			assert.Equal(t, 1, writes)
		})
	}
}

func TestPowerDNSConfigurationAndCancellation(t *testing.T) {
	t.Run("invalid endpoint", func(t *testing.T) {
		mockAPI(t, func(http.ResponseWriter, *http.Request) { t.Error("invalid endpoint must not make requests") })
		for _, endpoint := range []string{"", "file:///tmp/api", "https://dns.example.net/?key=x", "https://user:password@dns.example.net/"} {
			config := powerDNSTestConfig()
			config.Endpoint = endpoint
			ip := net.ParseIP("192.0.2.1")
			require.Error(t, updatePowerDNS(context.Background(), &ip, &config))
		}
	})
	t.Run("cancellation", func(t *testing.T) {
		testTargetCancellation(t, "powerdns", powerDNSTestConfig())
	})
}

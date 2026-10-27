package target

import (
	"context"
	"encoding/xml"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"

	ddns "github.com/mingcheng/simplyddns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const route53ZoneResponse = `<GetHostedZoneResponse><HostedZone><Id>/hostedzone/Z1</Id><Name>example.com.</Name><CallerReference>test</CallerReference></HostedZone></GetHostedZoneResponse>`
const route53ChangeResponse = `<ChangeResourceRecordSetsResponse><ChangeInfo><Id>/change/C1</Id><Status>PENDING</Status><SubmittedAt>2026-10-27T00:00:00Z</SubmittedAt></ChangeInfo></ChangeResourceRecordSetsResponse>`

type route53TestChange struct {
	Action string
	Record struct {
		Name, Type, SetIdentifier, HealthCheckId string
		TTL                                      int64
		Weight                                   *int64
		Values                                   []string `xml:"ResourceRecords>ResourceRecord>Value"`
	} `xml:"ResourceRecordSet"`
}

func writeRoute53XML(t *testing.T, w http.ResponseWriter, value string) {
	t.Helper()
	w.Header().Set("Content-Type", "text/xml")
	_, err := fmt.Fprint(w, value)
	assert.NoError(t, err)
}

func TestRoute53Records(t *testing.T) {
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
				assert.Equal(t, "route53.amazonaws.com", r.URL.Host)
				assert.True(t, strings.HasPrefix(r.Header.Get("Authorization"), "AWS4-HMAC-SHA256 Credential=key/"))
				assert.Equal(t, "session", r.Header.Get("X-Amz-Security-Token"))
				if r.URL.Path == "/2013-04-01/hostedzone/Z1" {
					writeRoute53XML(t, w, route53ZoneResponse)
					return
				}
				assert.Equal(t, "/2013-04-01/hostedzone/Z1/rrset", r.URL.Path)
				if r.Method == http.MethodGet {
					assert.Equal(t, "home.example.com.", r.URL.Query().Get("name"))
					assert.Equal(t, tc.kind, r.URL.Query().Get("type"))
					record := ""
					if tc.old != "" {
						record = fmt.Sprintf(`<ResourceRecordSet><Name>home.example.com.</Name><Type>%s</Type><TTL>3600</TTL><ResourceRecords><ResourceRecord><Value>%s</Value></ResourceRecord></ResourceRecords></ResourceRecordSet>`, tc.kind, tc.old)
					}
					writeRoute53XML(t, w, `<ListResourceRecordSetsResponse><ResourceRecordSets>`+record+`</ResourceRecordSets><IsTruncated>false</IsTruncated><MaxItems>100</MaxItems></ListResourceRecordSetsResponse>`)
					return
				}
				writes++
				assert.Equal(t, http.MethodPost, r.Method)
				var body struct {
					Changes []route53TestChange `xml:"ChangeBatch>Changes>Change"`
				}
				require.NoError(t, xml.NewDecoder(r.Body).Decode(&body))
				require.Len(t, body.Changes, 1)
				change := body.Changes[0]
				assert.Equal(t, "UPSERT", change.Action)
				assert.Equal(t, "home.example.com.", change.Record.Name)
				assert.Equal(t, tc.kind, change.Record.Type)
				assert.Equal(t, []string{tc.address}, change.Record.Values)
				if tc.old == "" {
					assert.EqualValues(t, defaultRecordTTL, change.Record.TTL)
				} else {
					assert.EqualValues(t, 3600, change.Record.TTL)
				}
				writeRoute53XML(t, w, route53ChangeResponse)
			})
			ip := net.ParseIP(tc.address)
			require.NoError(t, updateRoute53(context.Background(), &ip, &ddns.TargetConfig{
				Key: "key", Token: "token", SessionToken: "session", Zone: "Z1", Domains: []string{"home.example.com"},
			}))
			if net.ParseIP(tc.old).Equal(ip) {
				assert.Zero(t, writes)
			} else {
				assert.Equal(t, 1, writes)
			}
		})
	}
}

func TestRoute53PreservesRoutingAcrossPages(t *testing.T) {
	pages := 0
	mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/2013-04-01/hostedzone/Z1" {
			writeRoute53XML(t, w, route53ZoneResponse)
			return
		}
		if r.Method == http.MethodGet {
			pages++
			id, continuation := "first", `<IsTruncated>true</IsTruncated><NextRecordName>home.example.com.</NextRecordName><NextRecordType>A</NextRecordType><NextRecordIdentifier>second</NextRecordIdentifier>`
			if pages == 2 {
				id, continuation = "second", `<IsTruncated>false</IsTruncated>`
				assert.Equal(t, "second", r.URL.Query().Get("identifier"))
			}
			record := fmt.Sprintf(`<ResourceRecordSet><Name>home.example.com.</Name><Type>A</Type><SetIdentifier>%s</SetIdentifier><Weight>10</Weight><HealthCheckId>health</HealthCheckId><TTL>3600</TTL><ResourceRecords><ResourceRecord><Value>192.0.2.2</Value></ResourceRecord></ResourceRecords></ResourceRecordSet>`, id)
			writeRoute53XML(t, w, `<ListResourceRecordSetsResponse><ResourceRecordSets>`+record+`</ResourceRecordSets>`+continuation+`<MaxItems>100</MaxItems></ListResourceRecordSetsResponse>`)
			return
		}
		var body struct {
			Changes []route53TestChange `xml:"ChangeBatch>Changes>Change"`
		}
		require.NoError(t, xml.NewDecoder(r.Body).Decode(&body))
		require.Len(t, body.Changes, 2)
		assert.Equal(t, "first", body.Changes[0].Record.SetIdentifier)
		assert.Equal(t, "second", body.Changes[1].Record.SetIdentifier)
		for _, change := range body.Changes {
			require.NotNil(t, change.Record.Weight)
			assert.EqualValues(t, 10, *change.Record.Weight)
			assert.Equal(t, "health", change.Record.HealthCheckId)
			assert.EqualValues(t, 3600, change.Record.TTL)
			assert.Equal(t, []string{"192.0.2.1"}, change.Record.Values)
		}
		writeRoute53XML(t, w, route53ChangeResponse)
	})
	ip := net.ParseIP("192.0.2.1")
	require.NoError(t, updateRoute53(context.Background(), &ip, &ddns.TargetConfig{
		Key: "key", Token: "token", Zone: "Z1", Domains: []string{"home.example.com"},
	}))
	assert.Equal(t, 2, pages)
}

func TestRoute53FailuresAndCancellation(t *testing.T) {
	for _, name := range []string{"lookup denied", "alias", "missing continuation", "outside zone", "write failure", "missing change ID", "failed change state"} {
		t.Run(name, func(t *testing.T) {
			config := ddns.TargetConfig{Key: "key", Token: "token", Zone: "Z1", Domains: []string{"home.example.com"}}
			if name == "outside zone" {
				config.Domains = []string{"badexample.com"}
			}
			mockAPI(t, func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/2013-04-01/hostedzone/Z1" {
					writeRoute53XML(t, w, route53ZoneResponse)
					return
				}
				if name == "lookup denied" || (name == "write failure" && r.Method == http.MethodPost) {
					w.WriteHeader(http.StatusForbidden)
					writeRoute53XML(t, w, `<ErrorResponse><Error><Code>AccessDenied</Code><Message>denied</Message></Error><RequestId>request</RequestId></ErrorResponse>`)
					return
				}
				if name != "write failure" && name != "missing change ID" && name != "failed change state" {
					assert.Equal(t, http.MethodGet, r.Method)
				}
				if r.Method == http.MethodPost {
					body := `<ChangeResourceRecordSetsResponse/>`
					if name == "failed change state" {
						body = strings.Replace(route53ChangeResponse, "PENDING", "FAILED", 1)
					}
					writeRoute53XML(t, w, body)
					return
				}
				record, continuation := "", `<IsTruncated>false</IsTruncated>`
				if name == "alias" {
					record = `<ResourceRecordSet><Name>home.example.com.</Name><Type>A</Type><AliasTarget><HostedZoneId>Z2</HostedZoneId><DNSName>alias.example.net.</DNSName><EvaluateTargetHealth>false</EvaluateTargetHealth></AliasTarget></ResourceRecordSet>`
				} else if name == "missing continuation" {
					continuation = `<IsTruncated>true</IsTruncated>`
				}
				writeRoute53XML(t, w, `<ListResourceRecordSetsResponse><ResourceRecordSets>`+record+`</ResourceRecordSets>`+continuation+`<MaxItems>100</MaxItems></ListResourceRecordSetsResponse>`)
			})
			ip := net.ParseIP("192.0.2.1")
			require.Error(t, updateRoute53(context.Background(), &ip, &config))
		})
	}
	t.Run("cancellation", func(t *testing.T) {
		testTargetCancellation(t, "route53", ddns.TargetConfig{Key: "key", Token: "token", Zone: "Z1", Domains: []string{"home.example.com"}})
	})
}

package source

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"testing/synctest"

	"github.com/mingcheng/simplyddns"
	_ "github.com/mingcheng/simplyddns/target"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutomaticHTTPSourceSelection(t *testing.T) {
	for _, sourceType := range []string{"", " AUTO ", "auto-ipv4", "auto-ipv6"} {
		for _, count := range []int{0, 3, 5, 7} {
			t.Run(fmt.Sprintf("%s/count=%d", sourceType, count), func(t *testing.T) {
				t.Setenv(simplyddns.AutoSourceCountEnv, strconv.Itoa(count))
				wantCount := count
				if count == 0 {
					require.NoError(t, os.Unsetenv(simplyddns.AutoSourceCountEnv))
					wantCount = 3
				}
				synctest.Test(t, func(t *testing.T) {
					allowed := []string{
						"ipify", "ipsb", "opendns", "icanhazip-ipv4",
						"ident-ipv4", "myexternalip-ipv4", "wtfismyip-ipv4",
					}
					address := "192.0.2.10"
					if sourceType == "auto-ipv6" {
						allowed = []string{
							"ipify-ipv6", "ipsb-ipv6", "opendns-ipv6", "icanhazip-ipv6",
							"ident-ipv6", "ipinfo-ipv6", "wtfismyip-ipv6",
						}
						address = "2001:db8::10"
					}
					responses := make(map[string]string)
					for _, fixture := range httpSourceTests {
						if slices.Contains(allowed, fixture.name) {
							responses[fixture.url] = fmt.Sprintf(fixture.body, address)
						}
					}
					require.Len(t, responses, 7)
					requests := make(chan string, len(httpSourceTests))
					release := make(chan struct{})
					mockHTTP(t, func(r *http.Request) (*http.Response, error) {
						requests <- r.URL.String()
						select {
						case <-release:
						case <-r.Context().Done():
							return nil, r.Context().Err()
						}
						body, ok := responses[r.URL.String()]
						if !ok {
							t.Errorf("unexpected automatic source request: %s", r.URL)
							return nil, fmt.Errorf("unexpected source: %s", r.URL)
						}
						return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
					})
					job, err := simplyddns.NewJob(simplyddns.JobConfig{
						Source: simplyddns.SourceConfig{Type: sourceType},
						Target: simplyddns.TargetConfig{Type: "sleep"},
					})
					require.NoError(t, err)
					defer job.Stop()
					type outcome struct {
						ip  *net.IP
						err error
					}
					done := make(chan outcome, 1)
					go func() {
						ip, err := job.Source(context.Background(), &job.Config.Source)
						done <- outcome{ip, err}
					}()
					synctest.Wait()
					assert.Len(t, requests, wantCount, "configured sources must be queried concurrently")
					seen := make(map[string]bool)
					for range len(requests) {
						url := <-requests
						assert.Contains(t, responses, url)
						assert.False(t, seen[url], "duplicate vote from the same endpoint")
						seen[url] = true
					}
					close(release)
					result := <-done
					require.NoError(t, result.err)
					require.NotNil(t, result.ip)
					assert.Equal(t, address, result.ip.String())
				})
			})
		}
	}
}

func TestAutomaticHTTPSourceRejectsInvalidCount(t *testing.T) {
	t.Setenv(simplyddns.AutoSourceCountEnv, "4")
	mockHTTP(t, func(r *http.Request) (*http.Response, error) {
		t.Errorf("invalid automatic configuration must not make requests: %s", r.URL)
		return nil, fmt.Errorf("unexpected request")
	})
	for _, name := range []string{"auto", "auto-ipv4", "auto-ipv6"} {
		fn, err := simplyddns.SourceFuncByName(name)
		require.NoError(t, err)
		ip, err := fn(context.Background(), nil)
		require.ErrorContains(t, err, simplyddns.AutoSourceCountEnv)
		assert.Nil(t, ip)
	}
}

func TestAutomaticHTTPJobsRemainIndependent(t *testing.T) {
	t.Setenv(simplyddns.AutoSourceCountEnv, "5")
	mockHTTP(t, func(r *http.Request) (*http.Response, error) {
		for _, fixture := range httpSourceTests {
			if fixture.url != r.URL.String() {
				continue
			}
			address := "192.0.2.10"
			if strings.HasSuffix(fixture.name, "-ipv6") {
				address = "2001:db8::10"
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(fmt.Sprintf(fixture.body, address))),
			}, nil
		}
		return nil, fmt.Errorf("unexpected request: %s", r.URL)
	})
	var workers sync.WaitGroup
	for _, tc := range []struct {
		sourceType, address string
	}{
		{"auto", "192.0.2.10"},
		{"auto-ipv6", "2001:db8::10"},
	} {
		job, err := simplyddns.NewJob(simplyddns.JobConfig{
			Source: simplyddns.SourceConfig{Type: tc.sourceType},
			Target: simplyddns.TargetConfig{Type: "sleep"},
		})
		require.NoError(t, err)
		defer job.Stop()
		workers.Go(func() {
			for range 2 {
				ip, err := job.Source(context.Background(), &job.Config.Source)
				if assert.NoError(t, err) && assert.NotNil(t, ip) {
					assert.Equal(t, tc.address, ip.String())
				}
			}
		})
	}
	workers.Wait()
}

func TestExplicitHTTPSourceRequestScope(t *testing.T) {
	t.Setenv(simplyddns.AutoSourceCountEnv, "invalid")
	for _, tc := range []struct {
		name, sourceType string
		failures         []string
		disagree         bool
		wantSources      []string
		wantIP           string
		newJobError      bool
	}{
		{"single server", "ipify", nil, false, []string{"ipify"}, "192.0.2.10", false},
		{"single failure", "ipify", []string{"ipify"}, false, []string{"ipify"}, "", false},
		{"explicit list", "ipify,ipsb", nil, false, []string{"ipify", "ipsb"}, "192.0.2.10", false},
		{"normalized names", " IPIFY , IPSB ", nil, false, []string{"ipify", "ipsb"}, "192.0.2.10", false},
		{"disagreement", "ipify,ipsb", nil, true, []string{"ipify", "ipsb"}, "", false},
		{"all failed without expansion", "ipify,ipsb", []string{"ipify", "ipsb"}, false, []string{"ipify", "ipsb"}, "", false},
		{"fallback stays in list", "ipify,ipsb", []string{"ipify"}, false, []string{"ipify", "ipsb"}, "192.0.2.10", false},
		{"explicit IPv6", "icanhazip-ipv6,ident-ipv6", nil, false, []string{"icanhazip-ipv6", "ident-ipv6"}, "2001:db8::10", false},
		{"loopback stays offline", "lo", nil, false, nil, "127.0.0.1", false},
		{"static stays offline", "static", nil, false, nil, "192.0.2.50", false},
		{"unknown name", "ipify,unknown-source", nil, false, nil, "", true},
		{"empty name", "ipify,", nil, false, nil, "", true},
		{"mixed automatic and explicit", "auto,ipify", nil, false, nil, "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var mu sync.Mutex
			var requested []string
			mockHTTP(t, func(r *http.Request) (*http.Response, error) {
				mu.Lock()
				requested = append(requested, r.URL.String())
				mu.Unlock()
				for _, fixture := range httpSourceTests {
					if fixture.url != r.URL.String() {
						continue
					}
					if slices.Contains(tc.failures, fixture.name) {
						return &http.Response{
							StatusCode: http.StatusServiceUnavailable,
							Status:     "503 Service Unavailable",
							Body:       http.NoBody,
						}, nil
					}
					address := "192.0.2.10"
					if strings.Contains(tc.sourceType, "-ipv6") {
						address = "2001:db8::10"
					}
					if tc.disagree && fixture.name == "ipsb" {
						address = "198.51.100.20"
					}
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(fmt.Sprintf(fixture.body, address))),
					}, nil
				}
				return nil, fmt.Errorf("unexpected request: %s", r.URL)
			})
			job, err := simplyddns.NewJob(simplyddns.JobConfig{
				Source: simplyddns.SourceConfig{Type: tc.sourceType, Content: "192.0.2.50"},
				Target: simplyddns.TargetConfig{Type: "sleep"},
			})
			if tc.newJobError {
				require.Error(t, err)
				assert.Nil(t, job)
			} else {
				require.NoError(t, err)
				defer job.Stop()
				ip, err := job.Source(context.Background(), &job.Config.Source)
				if tc.wantIP == "" {
					require.Error(t, err)
					assert.Nil(t, ip)
				} else {
					require.NoError(t, err)
					require.NotNil(t, ip)
					assert.Equal(t, tc.wantIP, ip.String())
				}
			}
			var expected []string
			for _, fixture := range httpSourceTests {
				if slices.Contains(tc.wantSources, fixture.name) {
					expected = append(expected, fixture.url)
				}
			}
			assert.ElementsMatch(t, expected, requested, "explicit sources must never query unspecified servers")
		})
	}
}

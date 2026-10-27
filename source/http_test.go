package source

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mingcheng/simplyddns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return fn(r) }

func mockHTTP(t *testing.T, fn roundTripFunc) {
	t.Helper()
	original := httpClient
	httpClient = &http.Client{Transport: fn, Timeout: simplyddns.DefaultTimeout}
	t.Cleanup(func() { httpClient = original })
}

func TestHTTPSources(t *testing.T) {
	for _, tc := range []struct {
		name, url, body, userAgent string
	}{
		{"cloudflare", "https://www.cloudflare.com/cdn-cgi/trace", "fl=abc\nip=%s\nwarp=off\n", UserAgent},
		{"cz88", "https://update.cz88.net/api/cz88/ip/geo?ip=", `{"code":200,"data":{"ip":"%s"}}`, UserAgentCurl},
		{"ipify", "https://api.ipify.org?format=text", " %s\n", UserAgent},
		{"ipinfo", "https://ipinfo.io/ip", "%s\n", UserAgent},
		{"ipip", "https://myip.ipip.net/json", `{"data":{"ip":"%s"}}`, UserAgent},
		{"ipplus360", "https://www.ipplus360.com/getIP", `{"data":"%s"}`, UserAgent},
		{"ipsb", "https://api-ipv4.ip.sb/ip", "%s\n", UserAgentCurl},
		{"myip", "https://api.my-ip.io/ip.txt", "%s\n", UserAgent},
		{"opendns", "https://myipv4.p1.opendns.com/get_my_ip", `{"ip":"%s"}`, UserAgent},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fn, err := simplyddns.SourceFuncByName(tc.name)
			require.NoError(t, err)
			for _, address := range []string{"192.0.2.10", "2001:db8::1", "invalid", ""} {
				t.Run(address, func(t *testing.T) {
					mockHTTP(t, func(r *http.Request) (*http.Response, error) {
						assert.Equal(t, http.MethodGet, r.Method)
						assert.Equal(t, tc.url, r.URL.String())
						assert.Equal(t, tc.userAgent, r.UserAgent())
						return &http.Response{
							StatusCode: http.StatusOK,
							Header:     make(http.Header),
							Body:       io.NopCloser(strings.NewReader(fmt.Sprintf(tc.body, address))),
						}, nil
					})
					ip, err := fn(context.Background(), nil)
					if address == "invalid" || address == "" {
						require.Error(t, err)
						assert.Nil(t, ip)
						return
					}
					require.NoError(t, err)
					require.NotNil(t, ip)
					assert.Equal(t, address, ip.String())
				})
			}
			t.Run("missing address", func(t *testing.T) {
				mockHTTP(t, func(*http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("{}"))}, nil
				})
				_, err := fn(context.Background(), nil)
				require.Error(t, err)
			})
			t.Run("HTTP error", func(t *testing.T) {
				mockHTTP(t, func(*http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusForbidden, Status: "403 Forbidden", Body: http.NoBody}, nil
				})
				_, err := fn(context.Background(), nil)
				require.ErrorContains(t, err, "403")
			})
			t.Run("cancellation", func(t *testing.T) {
				mockHTTP(t, func(r *http.Request) (*http.Response, error) {
					return nil, r.Context().Err()
				})
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				_, err := fn(ctx, nil)
				require.ErrorIs(t, err, context.Canceled)
			})
		})
	}
}

func TestRawHTTPHelpers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "custom", r.UserAgent())
		assert.Equal(t, "https://example.com", r.Referer())
		fmt.Fprintln(w, "192.0.2.1")
	}))
	defer server.Close()
	body, err := RawStrByURL(context.Background(), server.URL, map[string]string{
		"User-Agent": "custom", "Referer": "https://example.com",
	})
	require.NoError(t, err)
	assert.Equal(t, "192.0.2.1\n", body)

	ipServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "2001:db8::1")
	}))
	defer ipServer.Close()
	ip, err := RawIPByURL(ipServer.URL)
	require.NoError(t, err)
	assert.Equal(t, "2001:db8::1", ip.String())

	_, err = RawStrByURL(context.Background(), "://invalid", nil)
	require.Error(t, err)
}

func TestHTTPTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err := RawStrByURL(ctx, server.URL, nil)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

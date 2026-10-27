package source

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
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

type httpSourceTest struct {
	name, url, body, userAgent string
	family                     int
	apiErrors                  []string
}

var httpSourceTests = []httpSourceTest{cloudflareTestCase, cz88TestCase, ipifyTestCase, ipinfoTestCase, ipipTestCase, ipplus360TestCase, ipsbTestCase, myipTestCase, opendnsTestCase, countryisTestCase, curlmyipTestCase, eth0TestCase, freeipapiTestCase, httpbinTestCase, httpbingoTestCase, icanhazipTestCase, icanhazipIPv4TestCase, icanhazipIPv6TestCase, identTestCase, identIPv4TestCase, identIPv6TestCase, ifconfigcoTestCase, ifconfigioTestCase, ifconfigmeTestCase, ipaddressTestCase, ipconfigioTestCase, ipechoTestCase, ipguideTestCase, ipifyDualTestCase, ipifyIPv6TestCase, ipinfoIPv6TestCase, iplocationTestCase, ipqueryTestCase, ipsbDualTestCase, ipsbIPv6TestCase, ipwhoisTestCase, myexternalipTestCase, myexternalipIPv4TestCase, opendnsIPv6TestCase, tencentTestCase, tnediTestCase, tnediIPv4TestCase, tnediIPv6TestCase, tykTestCase, ustcTestCase, wtfismyipTestCase, wtfismyipIPv4TestCase, wtfismyipIPv6TestCase, zxincTestCase}

func testHTTPSource(t *testing.T, tc httpSourceTest) {
	t.Helper()
	fn, err := simplyddns.SourceFuncByName(tc.name)
	require.NoError(t, err)
	for _, address := range []struct {
		name, value string
		family      int
	}{
		{"IPv4", "192.0.2.10", 4},
		{"IPv6", "2001:db8::1", 6},
		{"mapped IPv4", "::ffff:192.0.2.10", 4},
		{"invalid", "invalid", 0},
		{"empty", "", 0},
		{"address list", "192.0.2.10, 198.51.100.20", 0},
		{"HTML", "<html>192.0.2.10</html>", 0},
	} {
		t.Run(address.name, func(t *testing.T) {
			mockHTTP(t, func(r *http.Request) (*http.Response, error) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, tc.url, r.URL.String())
				assert.Equal(t, tc.userAgent, r.UserAgent())
				if strings.HasPrefix(tc.body, "{") {
					assert.Equal(t, "application/json", r.Header.Get("Accept"))
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(fmt.Sprintf(tc.body, address.value))),
				}, nil
			})
			ip, err := fn(context.Background(), nil)
			if address.family == 0 || tc.family != 0 && tc.family != address.family {
				require.Error(t, err)
				assert.Nil(t, ip)
				if address.family != 0 {
					assert.ErrorContains(t, err, fmt.Sprintf("expected IPv%d", tc.family))
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, ip)
			assert.Equal(t, net.ParseIP(address.value).String(), ip.String())
		})
	}
	t.Run("missing address", func(t *testing.T) {
		mockHTTP(t, func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("{}"))}, nil
		})
		_, err := fn(context.Background(), nil)
		require.Error(t, err)
	})
	if strings.HasPrefix(tc.body, "{") {
		t.Run("invalid JSON", func(t *testing.T) {
			mockHTTP(t, func(*http.Request) (*http.Response, error) {
				body := fmt.Sprintf(tc.body, "192.0.2.10") + "trailing garbage"
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
			})
			ip, err := fn(context.Background(), nil)
			require.ErrorContains(t, err, "invalid JSON")
			assert.Nil(t, ip)
		})
	}
	for _, code := range []int{http.StatusForbidden, http.StatusTooManyRequests, http.StatusServiceUnavailable} {
		t.Run(fmt.Sprintf("HTTP %d", code), func(t *testing.T) {
			mockHTTP(t, func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: code,
					Status:     fmt.Sprintf("%d %s", code, http.StatusText(code)),
					Body:       http.NoBody,
				}, nil
			})
			ip, err := fn(context.Background(), nil)
			require.ErrorContains(t, err, fmt.Sprint(code))
			assert.Nil(t, ip)
		})
	}
	t.Run("cancellation", func(t *testing.T) {
		mockHTTP(t, func(r *http.Request) (*http.Response, error) {
			return nil, r.Context().Err()
		})
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := fn(ctx, nil)
		require.ErrorIs(t, err, context.Canceled)
	})
	for i, body := range tc.apiErrors {
		t.Run(fmt.Sprintf("API error %d", i+1), func(t *testing.T) {
			mockHTTP(t, func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
			})
			ip, err := fn(context.Background(), nil)
			require.Error(t, err)
			assert.Nil(t, ip)
		})
	}

}

func TestHTTPSourceDocumentation(t *testing.T) {
	for _, name := range []string{"README.md", "README_ZH.md"} {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile("../" + name)
			require.NoError(t, err)
			for _, tc := range httpSourceTests {
				assert.Contains(t, string(data), "`"+tc.name+"`", "missing source %s", tc.name)
			}
		})
	}
}

func TestOnlineHTTPSources(t *testing.T) {
	selected := os.Getenv("SIMPLYDDNS_LIVE_SOURCES")
	if selected == "" {
		t.Skip("set SIMPLYDDNS_LIVE_SOURCES to a comma-separated list or all to enable live API checks")
	}
	names := strings.Split(selected, ",")
	if selected == "all" {
		names = nil
		for _, tc := range httpSourceTests {
			names = append(names, tc.name)
		}
	}
	for _, name := range names {
		name = strings.TrimSpace(name)
		require.NotEmpty(t, name)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			fn, err := simplyddns.SourceFuncByName(name)
			require.NoError(t, err)
			ctx, cancel := context.WithTimeout(context.Background(), simplyddns.DefaultTimeout)
			defer cancel()
			ip, err := fn(ctx, nil)
			require.NoError(t, err)
			require.NotNil(t, ip)
			assert.NotNil(t, ip.To16(), "source must return a valid IP address")
			family := 6
			if ip.To4() != nil {
				family = 4
			}
			t.Logf("returned IPv%d", family)
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

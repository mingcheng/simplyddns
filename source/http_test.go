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

var httpSourceTests = []struct {
	name, url, body, userAgent string
	family                     int
}{
	{"cloudflare", "https://www.cloudflare.com/cdn-cgi/trace", "fl=abc\nip=%s\nwarp=off\n", UserAgent, 0},
	{"cz88", "https://update.cz88.net/api/cz88/ip/geo?ip=", `{"code":200,"data":{"ip":"%s"}}`, UserAgentCurl, 0},
	{"ipify", "https://api.ipify.org?format=text", " %s\n", UserAgent, 0},
	{"ipinfo", "https://ipinfo.io/ip", "%s\n", UserAgent, 0},
	{"ipip", "https://myip.ipip.net/json", `{"data":{"ip":"%s"}}`, UserAgent, 0},
	{"ipplus360", "https://www.ipplus360.com/getIP", `{"data":"%s"}`, UserAgent, 0},
	{"ipsb", "https://api-ipv4.ip.sb/ip", "%s\n", UserAgentCurl, 0},
	{"myip", "https://api.my-ip.io/ip.txt", "%s\n", UserAgent, 0},
	{"opendns", "https://myipv4.p1.opendns.com/get_my_ip", `{"ip":"%s"}`, UserAgent, 0},
	{"countryis", "https://api.country.is/", `{"ip":"%s","country":"US"}`, UserAgentCurl, 0},
	{"curlmyip", "https://curlmyip.net/", "%s\n", UserAgentCurl, 0},
	{"eth0", "https://eth0.me/", "%s\n", UserAgentCurl, 0},
	{"freeipapi", "https://free.freeipapi.com/", "%s\n", UserAgentCurl, 0},
	{"httpbin", "https://httpbin.org/ip", `{"origin":"%s"}`, UserAgentCurl, 0},
	{"httpbingo", "https://httpbingo.org/ip", `{"origin":"%s"}`, UserAgentCurl, 0},
	{"icanhazip", "https://icanhazip.com/", "%s\n", UserAgentCurl, 0},
	{"icanhazip-ipv4", "https://ipv4.icanhazip.com/", "%s\n", UserAgentCurl, 4},
	{"icanhazip-ipv6", "https://ipv6.icanhazip.com/", "%s\n", UserAgentCurl, 6},
	{"ident", "https://ident.me/", "%s\n", UserAgentCurl, 0},
	{"ident-ipv4", "https://4.ident.me/", "%s\n", UserAgentCurl, 4},
	{"ident-ipv6", "https://6.ident.me/", "%s\n", UserAgentCurl, 6},
	{"ifconfigco", "https://ifconfig.co/ip", "%s\n", UserAgentCurl, 0},
	{"ifconfigio", "https://ifconfig.io/ip", "%s\n", UserAgentCurl, 0},
	{"ifconfigme", "https://ifconfig.me/ip", "%s\n", UserAgentCurl, 0},
	{"ipaddress", "https://ipaddress.sh/", "%s\n", UserAgentCurl, 0},
	{"ipconfigio", "https://ipconfig.io/ip", "%s\n", UserAgentCurl, 0},
	{"ipecho", "https://ipecho.net/plain", "%s\n", UserAgentCurl, 0},
	{"ipguide", "https://ip.guide/", `{"ip":"%s","network":{"cidr":"192.0.2.0/24"}}`, UserAgentCurl, 0},
	{"ipify-dual", "https://api64.ipify.org?format=text", "%s\n", UserAgentCurl, 0},
	{"ipify-ipv6", "https://api6.ipify.org?format=text", "%s\n", UserAgentCurl, 6},
	{"ipinfo-ipv6", "https://v6.ipinfo.io/ip", "%s\n", UserAgentCurl, 6},
	{"iplocation", "https://api.iplocation.net/?cmd=get-ip", `{"ip":"%s","response_code":200,"response_message":"OK"}`, UserAgentCurl, 0},
	{"ipquery", "https://api.ipquery.io/", "%s\n", UserAgentCurl, 0},
	{"ipsb-dual", "https://api.ip.sb/ip", "%s\n", UserAgentCurl, 0},
	{"ipsb-ipv6", "https://api-ipv6.ip.sb/ip", "%s\n", UserAgentCurl, 6},
	{"ipwhois", "https://ipwho.is/?fields=ip,success,message", `{"ip":"%s","success":true}`, UserAgentCurl, 0},
	{"myexternalip", "https://myexternalip.com/raw", "%s\n", UserAgentCurl, 0},
	{"myexternalip-ipv4", "https://ipv4.myexternalip.com/raw", "%s\n", UserAgentCurl, 4},
	{"opendns-ipv6", "https://myipv6.p1.opendns.com/get_my_ip", `{"ip":"%s"}`, UserAgentCurl, 6},
	{"tencent", "https://r.inews.qq.com/api/ip2city", `{"ret":0,"errMsg":"","ip":"%s"}`, UserAgentCurl, 0},
	{"tnedi", "https://tnedi.me/", "%s\n", UserAgentCurl, 0},
	{"tnedi-ipv4", "https://4.tnedi.me/", "%s\n", UserAgentCurl, 4},
	{"tnedi-ipv6", "https://6.tnedi.me/", "%s\n", UserAgentCurl, 6},
	{"tyk", "https://ip.tyk.nu/", "%s\n", UserAgentCurl, 0},
	{"ustc", "https://test.ustc.edu.cn/backend/getIP.php", `{"processedString":"%s","rawIspInfo":""}`, UserAgentCurl, 0},
	{"wtfismyip", "https://wtfismyip.com/text", "%s\n", UserAgentCurl, 0},
	{"wtfismyip-ipv4", "https://ipv4.wtfismyip.com/text", "%s\n", UserAgentCurl, 4},
	{"wtfismyip-ipv6", "https://ipv6.wtfismyip.com/text", "%s\n", UserAgentCurl, 6},
	{"zxinc", "https://ip.zxinc.org/api.php?type=json", `{"code":0,"data":{"myip":"%s"}}`, UserAgentCurl, 0},
}

func TestHTTPSources(t *testing.T) {
	for _, tc := range httpSourceTests {
		t.Run(tc.name, func(t *testing.T) {
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
		})
	}
}

func TestHTTPSourceAPIErrors(t *testing.T) {
	for _, tc := range []struct {
		name, body string
	}{
		{"cz88", `{"code":403,"data":{"ip":"192.0.2.10"}}`},
		{"iplocation", `{"response_code":403,"ip":"192.0.2.10"}`},
		{"iplocation", `{"ip":"192.0.2.10"}`},
		{"ipwhois", `{"success":false,"ip":"192.0.2.10","message":"rate limited"}`},
		{"ipwhois", `{"ip":"192.0.2.10"}`},
		{"tencent", `{"ret":-1,"ip":"192.0.2.10"}`},
		{"tencent", `{"ip":"192.0.2.10"}`},
		{"zxinc", `{"code":1,"data":{"myip":"192.0.2.10"}}`},
		{"zxinc", `{"data":{"myip":"192.0.2.10"}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mockHTTP(t, func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(tc.body))}, nil
			})
			fn, err := simplyddns.SourceFuncByName(tc.name)
			require.NoError(t, err)
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

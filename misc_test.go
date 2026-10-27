package simplyddns

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDomain(t *testing.T) {
	for _, tc := range []struct {
		domain    string
		root      string
		subdomain string
	}{
		{"example.com", "example.com", ""},
		{"a.b.example.co.uk", "example.co.uk", "a.b"},
		{"*.example.com", "example.com", "*"},
	} {
		t.Run(tc.domain, func(t *testing.T) {
			domain, err := ParseDomain(tc.domain)
			require.NoError(t, err)
			assert.Equal(t, tc.root, domain.Domain+"."+domain.TLD)
			assert.Equal(t, tc.subdomain, domain.Subdomain)
		})
	}
}

func useTestDNS(t *testing.T) {
	t.Helper()
	original := lookupIP
	lookupIP = func(ctx context.Context, _, host string) ([]net.IP, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if host == "home.example.com" {
			return []net.IP{net.ParseIP("192.0.2.1")}, nil
		}
		return nil, &net.DNSError{Err: "no such host", Name: host, IsNotFound: true}
	}
	t.Cleanup(func() { lookupIP = original })
}

func TestValidateRecords(t *testing.T) {
	useTestDNS(t)
	ip := net.ParseIP("192.0.2.1")
	require.NoError(t, ValidateRecord("home.example.com", &ip))
	require.NoError(t, ValidateRecords([]string{"home.example.com"}, &ip))
	require.Error(t, ValidateRecords([]string{"home.example.com", "missing.example.com"}, &ip))
	other := net.ParseIP("192.0.2.2")
	require.Error(t, ValidateRecord("home.example.com", &other))
	require.Error(t, ValidateRecord("home.example.com", nil))
	require.Error(t, ValidateRecord("not a domain", &ip))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.Error(t, ValidateRecordContext(ctx, "home.example.com", &ip))
}

func TestJobSkipsMatchingDNSRecords(t *testing.T) {
	useTestDNS(t)
	config := testJobConfig(t)
	config.Target.Domains = []string{"home.example.com"}
	job, err := NewJob(config)
	require.NoError(t, err)
	defer job.Stop()
	calls := 0
	job.TargetFunc = func(context.Context, *net.IP, *TargetConfig) error { calls++; return nil }
	require.NoError(t, job.run(context.Background()))
	assert.Zero(t, calls)
	job.Config.Target.Domains = []string{"missing.example.com"}
	require.NoError(t, job.run(context.Background()))
	assert.Equal(t, 1, calls)
}

func TestProxyHTTPClient(t *testing.T) {
	for _, addr := range []string{"", "localhost", ":1080", "localhost:invalid", "localhost:0", "localhost:65536"} {
		_, err := ProxyHttpClient(addr)
		require.Error(t, err)
	}
	client, err := ProxyHttpClient("127.0.0.1:1080")
	require.NoError(t, err)
	assert.Equal(t, DefaultTimeout, client.Timeout)
	assert.NotNil(t, client.Transport)
}

func TestParseDomainRejectsInvalidNames(t *testing.T) {
	for _, domain := range []string{
		"", "com", "localhost", "192.0.2.1", "https://example.com", "example.com/path",
		"example.com:80", "user@example.com", "example.com?query=value", "example.com#fragment",
		".example.com", "a..example.com", "-a.example.com", "a-.example.com", " example.com",
	} {
		t.Run(domain, func(t *testing.T) {
			_, err := ParseDomain(domain)
			assert.Error(t, err)
		})
	}
}

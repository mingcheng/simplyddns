package source

import "testing"

var cloudflareTestCase = httpSourceTest{
	name:      "cloudflare",
	url:       "https://www.cloudflare.com/cdn-cgi/trace",
	body:      "fl=abc\nip=%s\nwarp=off\n",
	userAgent: UserAgent,
	family:    0,
}

func TestSourceCloudflare(t *testing.T) {
	testHTTPSource(t, cloudflareTestCase)
}

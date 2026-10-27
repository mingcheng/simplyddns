package source

import "testing"

var httpbinTestCase = httpSourceTest{
	name:      "httpbin",
	url:       "https://httpbin.org/ip",
	body:      `{"origin":"%s"}`,
	userAgent: UserAgentCurl,
	family:    0,
}

func TestSourceHTTPBin(t *testing.T) {
	testHTTPSource(t, httpbinTestCase)
}

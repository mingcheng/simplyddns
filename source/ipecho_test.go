package source

import "testing"

var ipechoTestCase = httpSourceTest{
	name:      "ipecho",
	url:       "https://ipecho.net/plain",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    0,
}

func TestSourceIPEcho(t *testing.T) {
	testHTTPSource(t, ipechoTestCase)
}

package source

import "testing"

var ipifyIPv6TestCase = httpSourceTest{
	name:      "ipify-ipv6",
	url:       "https://api6.ipify.org?format=text",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    6,
}

func TestSourceIPifyIPv6(t *testing.T) {
	testHTTPSource(t, ipifyIPv6TestCase)
}

package source

import "testing"

var ipsbTestCase = httpSourceTest{
	name:      "ipsb",
	url:       "https://api-ipv4.ip.sb/ip",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    0,
}

func TestSourceIPSB(t *testing.T) {
	testHTTPSource(t, ipsbTestCase)
}

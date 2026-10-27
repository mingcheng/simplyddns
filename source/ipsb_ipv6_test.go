package source

import "testing"

var ipsbIPv6TestCase = httpSourceTest{
	name:      "ipsb-ipv6",
	url:       "https://api-ipv6.ip.sb/ip",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    6,
}

func TestSourceIPSBIPv6(t *testing.T) {
	testHTTPSource(t, ipsbIPv6TestCase)
}

package source

import "testing"

var opendnsIPv6TestCase = httpSourceTest{
	name:      "opendns-ipv6",
	url:       "https://myipv6.p1.opendns.com/get_my_ip",
	body:      `{"ip":"%s"}`,
	userAgent: UserAgentCurl,
	family:    6,
}

func TestSourceOpenDNSIPv6(t *testing.T) {
	testHTTPSource(t, opendnsIPv6TestCase)
}

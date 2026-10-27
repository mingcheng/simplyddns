package source

import "testing"

var icanhazipIPv6TestCase = httpSourceTest{
	name:      "icanhazip-ipv6",
	url:       "https://ipv6.icanhazip.com/",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    6,
}

func TestSourceICanHazIPIPv6(t *testing.T) {
	testHTTPSource(t, icanhazipIPv6TestCase)
}

package source

import "testing"

var icanhazipIPv4TestCase = httpSourceTest{
	name:      "icanhazip-ipv4",
	url:       "https://ipv4.icanhazip.com/",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    4,
}

func TestSourceICanHazIPIPv4(t *testing.T) {
	testHTTPSource(t, icanhazipIPv4TestCase)
}

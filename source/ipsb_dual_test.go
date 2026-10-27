package source

import "testing"

var ipsbDualTestCase = httpSourceTest{
	name:      "ipsb-dual",
	url:       "https://api.ip.sb/ip",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    0,
}

func TestSourceIPSBDual(t *testing.T) {
	testHTTPSource(t, ipsbDualTestCase)
}

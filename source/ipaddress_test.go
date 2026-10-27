package source

import "testing"

var ipaddressTestCase = httpSourceTest{
	name:      "ipaddress",
	url:       "https://ipaddress.sh/",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    0,
}

func TestSourceIPAddress(t *testing.T) {
	testHTTPSource(t, ipaddressTestCase)
}

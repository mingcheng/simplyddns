package source

import "testing"

var ipinfoIPv6TestCase = httpSourceTest{
	name:      "ipinfo-ipv6",
	url:       "https://v6.ipinfo.io/ip",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    6,
}

func TestSourceIPInfoIPv6(t *testing.T) {
	testHTTPSource(t, ipinfoIPv6TestCase)
}

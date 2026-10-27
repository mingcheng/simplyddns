package source

import "testing"

var tnediIPv6TestCase = httpSourceTest{
	name:      "tnedi-ipv6",
	url:       "https://6.tnedi.me/",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    6,
}

func TestSourceTnediIPv6(t *testing.T) {
	testHTTPSource(t, tnediIPv6TestCase)
}

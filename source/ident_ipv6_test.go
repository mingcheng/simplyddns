package source

import "testing"

var identIPv6TestCase = httpSourceTest{
	name:      "ident-ipv6",
	url:       "https://6.ident.me/",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    6,
}

func TestSourceIdentIPv6(t *testing.T) {
	testHTTPSource(t, identIPv6TestCase)
}

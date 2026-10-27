package source

import "testing"

var identIPv4TestCase = httpSourceTest{
	name:      "ident-ipv4",
	url:       "https://4.ident.me/",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    4,
}

func TestSourceIdentIPv4(t *testing.T) {
	testHTTPSource(t, identIPv4TestCase)
}

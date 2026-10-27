package source

import "testing"

var tnediIPv4TestCase = httpSourceTest{
	name:      "tnedi-ipv4",
	url:       "https://4.tnedi.me/",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    4,
}

func TestSourceTnediIPv4(t *testing.T) {
	testHTTPSource(t, tnediIPv4TestCase)
}

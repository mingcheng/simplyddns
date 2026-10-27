package source

import "testing"

var ipifyDualTestCase = httpSourceTest{
	name:      "ipify-dual",
	url:       "https://api64.ipify.org?format=text",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    0,
}

func TestSourceIPifyDual(t *testing.T) {
	testHTTPSource(t, ipifyDualTestCase)
}

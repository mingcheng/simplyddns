package source

import "testing"

var icanhazipTestCase = httpSourceTest{
	name:      "icanhazip",
	url:       "https://icanhazip.com/",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    0,
}

func TestSourceICanHazIP(t *testing.T) {
	testHTTPSource(t, icanhazipTestCase)
}

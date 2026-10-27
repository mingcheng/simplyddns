package source

import "testing"

var ipguideTestCase = httpSourceTest{
	name:      "ipguide",
	url:       "https://ip.guide/",
	body:      `{"ip":"%s","network":{"cidr":"192.0.2.0/24"}}`,
	userAgent: UserAgentCurl,
	family:    0,
}

func TestSourceIPGuide(t *testing.T) {
	testHTTPSource(t, ipguideTestCase)
}

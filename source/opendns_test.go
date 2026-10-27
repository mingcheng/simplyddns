package source

import "testing"

var opendnsTestCase = httpSourceTest{
	name:      "opendns",
	url:       "https://myipv4.p1.opendns.com/get_my_ip",
	body:      `{"ip":"%s"}`,
	userAgent: UserAgent,
	family:    0,
}

func TestSourceOpenDNS(t *testing.T) {
	testHTTPSource(t, opendnsTestCase)
}

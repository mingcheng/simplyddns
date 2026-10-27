package source

import "testing"

var ipipTestCase = httpSourceTest{
	name:      "ipip",
	url:       "https://myip.ipip.net/json",
	body:      `{"data":{"ip":"%s"}}`,
	userAgent: UserAgent,
	family:    0,
}

func TestSourceIPIP(t *testing.T) {
	testHTTPSource(t, ipipTestCase)
}

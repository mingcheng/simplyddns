package source

import "testing"

var tnediTestCase = httpSourceTest{
	name:      "tnedi",
	url:       "https://tnedi.me/",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    0,
}

func TestSourceTnedi(t *testing.T) {
	testHTTPSource(t, tnediTestCase)
}

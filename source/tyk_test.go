package source

import "testing"

var tykTestCase = httpSourceTest{
	name:      "tyk",
	url:       "https://ip.tyk.nu/",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    0,
}

func TestSourceTyk(t *testing.T) {
	testHTTPSource(t, tykTestCase)
}

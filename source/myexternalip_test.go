package source

import "testing"

var myexternalipTestCase = httpSourceTest{
	name:      "myexternalip",
	url:       "https://myexternalip.com/raw",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    0,
}

func TestSourceMyExternalIP(t *testing.T) {
	testHTTPSource(t, myexternalipTestCase)
}

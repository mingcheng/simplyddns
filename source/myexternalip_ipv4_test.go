package source

import "testing"

var myexternalipIPv4TestCase = httpSourceTest{
	name:      "myexternalip-ipv4",
	url:       "https://ipv4.myexternalip.com/raw",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    4,
}

func TestSourceMyExternalIPIPv4(t *testing.T) {
	testHTTPSource(t, myexternalipIPv4TestCase)
}

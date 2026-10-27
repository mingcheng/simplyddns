package source

import "testing"

var ifconfigmeTestCase = httpSourceTest{
	name:      "ifconfigme",
	url:       "https://ifconfig.me/ip",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    0,
}

func TestSourceIfconfigMe(t *testing.T) {
	testHTTPSource(t, ifconfigmeTestCase)
}

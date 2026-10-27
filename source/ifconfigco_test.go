package source

import "testing"

var ifconfigcoTestCase = httpSourceTest{
	name:      "ifconfigco",
	url:       "https://ifconfig.co/ip",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    0,
}

func TestSourceIfconfigCo(t *testing.T) {
	testHTTPSource(t, ifconfigcoTestCase)
}

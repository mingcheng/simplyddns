package source

import "testing"

var ifconfigioTestCase = httpSourceTest{
	name:      "ifconfigio",
	url:       "https://ifconfig.io/ip",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    0,
}

func TestSourceIfconfigIO(t *testing.T) {
	testHTTPSource(t, ifconfigioTestCase)
}

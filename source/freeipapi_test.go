package source

import "testing"

var freeipapiTestCase = httpSourceTest{
	name:      "freeipapi",
	url:       "https://free.freeipapi.com/",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    0,
}

func TestSourceFreeIPAPI(t *testing.T) {
	testHTTPSource(t, freeipapiTestCase)
}

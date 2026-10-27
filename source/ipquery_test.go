package source

import "testing"

var ipqueryTestCase = httpSourceTest{
	name:      "ipquery",
	url:       "https://api.ipquery.io/",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    0,
}

func TestSourceIPQuery(t *testing.T) {
	testHTTPSource(t, ipqueryTestCase)
}

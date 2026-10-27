package source

import "testing"

var wtfismyipTestCase = httpSourceTest{
	name:      "wtfismyip",
	url:       "https://wtfismyip.com/text",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    0,
}

func TestSourceWTFIsMyIP(t *testing.T) {
	testHTTPSource(t, wtfismyipTestCase)
}

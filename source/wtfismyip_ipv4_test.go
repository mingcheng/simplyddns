package source

import "testing"

var wtfismyipIPv4TestCase = httpSourceTest{
	name:      "wtfismyip-ipv4",
	url:       "https://ipv4.wtfismyip.com/text",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    4,
}

func TestSourceWTFIsMyIPIPv4(t *testing.T) {
	testHTTPSource(t, wtfismyipIPv4TestCase)
}

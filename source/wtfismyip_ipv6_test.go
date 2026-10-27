package source

import "testing"

var wtfismyipIPv6TestCase = httpSourceTest{
	name:      "wtfismyip-ipv6",
	url:       "https://ipv6.wtfismyip.com/text",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    6,
}

func TestSourceWTFIsMyIPIPv6(t *testing.T) {
	testHTTPSource(t, wtfismyipIPv6TestCase)
}

package source

import "testing"

var curlmyipTestCase = httpSourceTest{
	name:      "curlmyip",
	url:       "https://curlmyip.net/",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    0,
}

func TestSourceCurlMyIP(t *testing.T) {
	testHTTPSource(t, curlmyipTestCase)
}

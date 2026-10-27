package source

import "testing"

var eth0TestCase = httpSourceTest{
	name:      "eth0",
	url:       "https://eth0.me/",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    0,
}

func TestSourceEth0(t *testing.T) {
	testHTTPSource(t, eth0TestCase)
}

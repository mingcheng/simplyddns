package source

import "testing"

var ipinfoTestCase = httpSourceTest{
	name:      "ipinfo",
	url:       "https://ipinfo.io/ip",
	body:      "%s\n",
	userAgent: UserAgent,
	family:    0,
}

func TestSourceIPInfo(t *testing.T) {
	testHTTPSource(t, ipinfoTestCase)
}

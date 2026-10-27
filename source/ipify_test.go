package source

import "testing"

var ipifyTestCase = httpSourceTest{
	name:      "ipify",
	url:       "https://api.ipify.org?format=text",
	body:      " %s\n",
	userAgent: UserAgent,
	family:    0,
}

func TestSourceIPify(t *testing.T) {
	testHTTPSource(t, ipifyTestCase)
}

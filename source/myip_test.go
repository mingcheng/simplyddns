package source

import "testing"

var myipTestCase = httpSourceTest{
	name:      "myip",
	url:       "https://api.my-ip.io/ip.txt",
	body:      "%s\n",
	userAgent: UserAgent,
	family:    0,
}

func TestSourceMyIP(t *testing.T) {
	testHTTPSource(t, myipTestCase)
}

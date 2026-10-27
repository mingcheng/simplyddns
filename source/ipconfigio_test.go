package source

import "testing"

var ipconfigioTestCase = httpSourceTest{
	name:      "ipconfigio",
	url:       "https://ipconfig.io/ip",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    0,
}

func TestSourceIPConfigIO(t *testing.T) {
	testHTTPSource(t, ipconfigioTestCase)
}

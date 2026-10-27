package source

import "testing"

var ipplus360TestCase = httpSourceTest{
	name:      "ipplus360",
	url:       "https://www.ipplus360.com/getIP",
	body:      `{"data":"%s"}`,
	userAgent: UserAgent,
	family:    0,
}

func TestSourceIPPlus360(t *testing.T) {
	testHTTPSource(t, ipplus360TestCase)
}

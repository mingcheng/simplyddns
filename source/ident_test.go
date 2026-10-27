package source

import "testing"

var identTestCase = httpSourceTest{
	name:      "ident",
	url:       "https://ident.me/",
	body:      "%s\n",
	userAgent: UserAgentCurl,
	family:    0,
}

func TestSourceIdent(t *testing.T) {
	testHTTPSource(t, identTestCase)
}

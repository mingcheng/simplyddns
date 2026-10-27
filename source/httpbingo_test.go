package source

import "testing"

var httpbingoTestCase = httpSourceTest{
	name:      "httpbingo",
	url:       "https://httpbingo.org/ip",
	body:      `{"origin":"%s"}`,
	userAgent: UserAgentCurl,
	family:    0,
}

func TestSourceHTTPBingo(t *testing.T) {
	testHTTPSource(t, httpbingoTestCase)
}

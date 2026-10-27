package source

import "testing"

var ipwhoisTestCase = httpSourceTest{
	name:      "ipwhois",
	url:       "https://ipwho.is/?fields=ip,success,message",
	body:      `{"ip":"%s","success":true}`,
	userAgent: UserAgentCurl,
	family:    0,
	apiErrors: []string{
		`{"success":false,"ip":"192.0.2.10","message":"rate limited"}`,
		`{"ip":"192.0.2.10"}`,
	},
}

func TestSourceIPWhois(t *testing.T) {
	testHTTPSource(t, ipwhoisTestCase)
}

package source

import "testing"

var iplocationTestCase = httpSourceTest{
	name:      "iplocation",
	url:       "https://api.iplocation.net/?cmd=get-ip",
	body:      `{"ip":"%s","response_code":200,"response_message":"OK"}`,
	userAgent: UserAgentCurl,
	family:    0,
	apiErrors: []string{
		`{"response_code":403,"ip":"192.0.2.10"}`,
		`{"ip":"192.0.2.10"}`,
	},
}

func TestSourceIPLocation(t *testing.T) {
	testHTTPSource(t, iplocationTestCase)
}

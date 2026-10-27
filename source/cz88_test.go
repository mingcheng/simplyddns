package source

import "testing"

var cz88TestCase = httpSourceTest{
	name:      "cz88",
	url:       "https://update.cz88.net/api/cz88/ip/geo?ip=",
	body:      `{"code":200,"data":{"ip":"%s"}}`,
	userAgent: UserAgentCurl,
	family:    0,
	apiErrors: []string{
		`{"code":403,"data":{"ip":"192.0.2.10"}}`,
	},
}

func TestSourceCZ88(t *testing.T) {
	testHTTPSource(t, cz88TestCase)
}

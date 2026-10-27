package source

import "testing"

var tencentTestCase = httpSourceTest{
	name:      "tencent",
	url:       "https://r.inews.qq.com/api/ip2city",
	body:      `{"ret":0,"errMsg":"","ip":"%s"}`,
	userAgent: UserAgentCurl,
	family:    0,
	apiErrors: []string{
		`{"ret":-1,"ip":"192.0.2.10"}`,
		`{"ip":"192.0.2.10"}`,
	},
}

func TestSourceTencent(t *testing.T) {
	testHTTPSource(t, tencentTestCase)
}

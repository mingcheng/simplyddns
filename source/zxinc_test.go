package source

import "testing"

var zxincTestCase = httpSourceTest{
	name:      "zxinc",
	url:       "https://ip.zxinc.org/api.php?type=json",
	body:      `{"code":0,"data":{"myip":"%s"}}`,
	userAgent: UserAgentCurl,
	family:    0,
	apiErrors: []string{
		`{"code":1,"data":{"myip":"192.0.2.10"}}`,
		`{"data":{"myip":"192.0.2.10"}}`,
	},
}

func TestSourceZXInc(t *testing.T) {
	testHTTPSource(t, zxincTestCase)
}

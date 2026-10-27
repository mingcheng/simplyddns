package source

import "testing"

var ustcTestCase = httpSourceTest{
	name:      "ustc",
	url:       "https://test.ustc.edu.cn/backend/getIP.php",
	body:      `{"processedString":"%s","rawIspInfo":""}`,
	userAgent: UserAgentCurl,
	family:    0,
}

func TestSourceUSTC(t *testing.T) {
	testHTTPSource(t, ustcTestCase)
}

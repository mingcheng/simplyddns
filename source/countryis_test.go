package source

import "testing"

var countryisTestCase = httpSourceTest{
	name:      "countryis",
	url:       "https://api.country.is/",
	body:      `{"ip":"%s","country":"US"}`,
	userAgent: UserAgentCurl,
	family:    0,
}

func TestSourceCountryIS(t *testing.T) {
	testHTTPSource(t, countryisTestCase)
}

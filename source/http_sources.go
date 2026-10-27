package source

import (
	"context"
	"fmt"
	"net"

	"github.com/mingcheng/simplyddns"
)

func init() {
	for _, endpoint := range []struct {
		name, url, jsonPath string
		family              int
		statuses            []jsonStatus
	}{
		{"countryis", "https://api.country.is/", "ip", 0, nil},
		{"curlmyip", "https://curlmyip.net/", "", 0, nil},
		{"eth0", "https://eth0.me/", "", 0, nil},
		{"freeipapi", "https://free.freeipapi.com/", "", 0, nil},
		{"httpbin", "https://httpbin.org/ip", "origin", 0, nil},
		{"httpbingo", "https://httpbingo.org/ip", "origin", 0, nil},
		{"icanhazip", "https://icanhazip.com/", "", 0, nil},
		{"icanhazip-ipv4", "https://ipv4.icanhazip.com/", "", 4, nil},
		{"icanhazip-ipv6", "https://ipv6.icanhazip.com/", "", 6, nil},
		{"ident", "https://ident.me/", "", 0, nil},
		{"ident-ipv4", "https://4.ident.me/", "", 4, nil},
		{"ident-ipv6", "https://6.ident.me/", "", 6, nil},
		{"ifconfigco", "https://ifconfig.co/ip", "", 0, nil},
		{"ifconfigio", "https://ifconfig.io/ip", "", 0, nil},
		{"ifconfigme", "https://ifconfig.me/ip", "", 0, nil},
		{"ipaddress", "https://ipaddress.sh/", "", 0, nil},
		{"ipconfigio", "https://ipconfig.io/ip", "", 0, nil},
		{"ipecho", "https://ipecho.net/plain", "", 0, nil},
		{"ipguide", "https://ip.guide/", "ip", 0, nil},
		{"ipify-dual", "https://api64.ipify.org?format=text", "", 0, nil},
		{"ipify-ipv6", "https://api6.ipify.org?format=text", "", 6, nil},
		{"ipinfo-ipv6", "https://v6.ipinfo.io/ip", "", 6, nil},
		{"iplocation", "https://api.iplocation.net/?cmd=get-ip", "ip", 0, []jsonStatus{{"response_code", "200"}}},
		{"ipquery", "https://api.ipquery.io/", "", 0, nil},
		{"ipsb-dual", "https://api.ip.sb/ip", "", 0, nil},
		{"ipsb-ipv6", "https://api-ipv6.ip.sb/ip", "", 6, nil},
		{"ipwhois", "https://ipwho.is/?fields=ip,success,message", "ip", 0, []jsonStatus{{"success", "true"}}},
		{"myexternalip", "https://myexternalip.com/raw", "", 0, nil},
		{"myexternalip-ipv4", "https://ipv4.myexternalip.com/raw", "", 4, nil},
		{"opendns-ipv6", "https://myipv6.p1.opendns.com/get_my_ip", "ip", 6, nil},
		{"tencent", "https://r.inews.qq.com/api/ip2city", "ip", 0, []jsonStatus{{"ret", "0"}}},
		{"tnedi", "https://tnedi.me/", "", 0, nil},
		{"tnedi-ipv4", "https://4.tnedi.me/", "", 4, nil},
		{"tnedi-ipv6", "https://6.tnedi.me/", "", 6, nil},
		{"tyk", "https://ip.tyk.nu/", "", 0, nil},
		{"ustc", "https://test.ustc.edu.cn/backend/getIP.php", "processedString", 0, nil},
		{"wtfismyip", "https://wtfismyip.com/text", "", 0, nil},
		{"wtfismyip-ipv4", "https://ipv4.wtfismyip.com/text", "", 4, nil},
		{"wtfismyip-ipv6", "https://ipv6.wtfismyip.com/text", "", 6, nil},
		{"zxinc", "https://ip.zxinc.org/api.php?type=json", "data.myip", 0, []jsonStatus{{"code", "0"}}},
	} {
		registerSource(endpoint.name, func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
			headers := map[string]string{"User-Agent": UserAgentCurl}
			var (
				addr *net.IP
				err  error
			)
			if endpoint.jsonPath == "" {
				addr, err = rawIPByURL(ctx, endpoint.url, headers)
			} else {
				headers["Accept"] = "application/json"
				addr, err = jsonIPByURL(ctx, endpoint.url, endpoint.jsonPath, headers, endpoint.statuses...)
			}
			if err != nil {
				return nil, fmt.Errorf("%s: %w", endpoint.name, err)
			}

			family := 6
			if addr.To4() != nil {
				family = 4
			}
			if endpoint.family != 0 && family != endpoint.family {
				return nil, fmt.Errorf("%s: expected IPv%d, got IPv%d", endpoint.name, endpoint.family, family)
			}
			return addr, nil
		})
	}
}

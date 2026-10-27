package source

import (
	"context"
	"net"

	"github.com/mingcheng/simplyddns"
)

func init() {
	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		return jsonIPByURL(ctx, "https://ip.zxinc.org/api.php?type=json", "data.myip", map[string]string{
			"Accept":     "application/json",
			"User-Agent": UserAgentCurl,
		}, jsonStatus{path: "code", value: "0"})
	}
	registerHTTPSource("zxinc", 0, fn)
}

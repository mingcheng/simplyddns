package source

import (
	"context"
	"net"

	"github.com/mingcheng/simplyddns"
)

func init() {
	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		return jsonIPByURL(ctx, "https://r.inews.qq.com/api/ip2city", "ip", map[string]string{
			"Accept":     "application/json",
			"User-Agent": UserAgentCurl,
		}, jsonStatus{path: "ret", value: "0"})
	}
	registerHTTPSource("tencent", 0, fn)
}

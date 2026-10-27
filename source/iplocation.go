package source

import (
	"context"
	"net"

	"github.com/mingcheng/simplyddns"
)

func init() {
	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		return jsonIPByURL(ctx, "https://api.iplocation.net/?cmd=get-ip", "ip", map[string]string{
			"Accept":     "application/json",
			"User-Agent": UserAgentCurl,
		}, jsonStatus{path: "response_code", value: "200"})
	}
	registerHTTPSource("iplocation", 0, fn)
}

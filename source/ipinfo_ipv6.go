package source

import (
	"context"
	"net"

	"github.com/mingcheng/simplyddns"
)

func init() {
	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		return rawIPByURL(ctx, "https://v6.ipinfo.io/ip", map[string]string{
			"User-Agent": UserAgentCurl,
		})
	}
	registerHTTPSource("ipinfo-ipv6", 6, fn)
}

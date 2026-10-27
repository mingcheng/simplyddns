package source

import (
	"context"
	"net"

	"github.com/mingcheng/simplyddns"
)

func init() {
	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		return rawIPByURL(ctx, "https://api-ipv6.ip.sb/ip", map[string]string{
			"User-Agent": UserAgentCurl,
		})
	}
	registerHTTPSource("ipsb-ipv6", 6, fn)
}

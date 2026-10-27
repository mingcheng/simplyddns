package source

import (
	"context"
	"net"

	"github.com/mingcheng/simplyddns"
)

func init() {
	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		return rawIPByURL(ctx, "https://6.tnedi.me/", map[string]string{
			"User-Agent": UserAgentCurl,
		})
	}
	registerHTTPSource("tnedi-ipv6", 6, fn)
}

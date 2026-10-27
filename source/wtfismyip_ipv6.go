package source

import (
	"context"
	"net"

	"github.com/mingcheng/simplyddns"
)

func init() {
	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		return rawIPByURL(ctx, "https://ipv6.wtfismyip.com/text", map[string]string{
			"User-Agent": UserAgentCurl,
		})
	}
	registerHTTPSource("wtfismyip-ipv6", 6, fn)
}

package source

import (
	"context"
	"net"

	"github.com/mingcheng/simplyddns"
)

func init() {
	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		return rawIPByURL(ctx, "https://ipv4.myexternalip.com/raw", map[string]string{
			"User-Agent": UserAgentCurl,
		})
	}
	registerHTTPSource("myexternalip-ipv4", 4, fn)
}

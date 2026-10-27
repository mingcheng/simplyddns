package source

import (
	"context"
	"net"

	"github.com/mingcheng/simplyddns"
)

func init() {
	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		return rawIPByURL(ctx, "https://ipaddress.sh/", map[string]string{
			"User-Agent": UserAgentCurl,
		})
	}
	registerHTTPSource("ipaddress", 0, fn)
}

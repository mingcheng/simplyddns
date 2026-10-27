package source

import (
	"context"
	"net"

	"github.com/mingcheng/simplyddns"
)

func init() {
	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		return rawIPByURL(ctx, "https://api.ipquery.io/", map[string]string{
			"User-Agent": UserAgentCurl,
		})
	}
	registerHTTPSource("ipquery", 0, fn)
}

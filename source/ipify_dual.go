package source

import (
	"context"
	"net"

	"github.com/mingcheng/simplyddns"
)

func init() {
	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		return rawIPByURL(ctx, "https://api64.ipify.org?format=text", map[string]string{
			"User-Agent": UserAgentCurl,
		})
	}
	registerHTTPSource("ipify-dual", 0, fn)
}

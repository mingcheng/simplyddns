package source

import (
	"context"
	"net"

	"github.com/mingcheng/simplyddns"
)

func init() {
	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		return rawIPByURL(ctx, "https://free.freeipapi.com/", map[string]string{
			"User-Agent": UserAgentCurl,
		})
	}
	registerHTTPSource("freeipapi", 0, fn)
}

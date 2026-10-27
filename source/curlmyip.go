package source

import (
	"context"
	"net"

	"github.com/mingcheng/simplyddns"
)

func init() {
	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		return rawIPByURL(ctx, "https://curlmyip.net/", map[string]string{
			"User-Agent": UserAgentCurl,
		})
	}
	registerHTTPSource("curlmyip", 0, fn)
}

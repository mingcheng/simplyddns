package source

import (
	"context"
	"net"

	"github.com/mingcheng/simplyddns"
)

func init() {
	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		return rawIPByURL(ctx, "https://ipconfig.io/ip", map[string]string{
			"User-Agent": UserAgentCurl,
		})
	}
	registerHTTPSource("ipconfigio", 0, fn)
}

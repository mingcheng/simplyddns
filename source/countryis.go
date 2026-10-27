package source

import (
	"context"
	"net"

	"github.com/mingcheng/simplyddns"
)

func init() {
	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		return jsonIPByURL(ctx, "https://api.country.is/", "ip", map[string]string{
			"Accept":     "application/json",
			"User-Agent": UserAgentCurl,
		})
	}
	registerHTTPSource("countryis", 0, fn)
}

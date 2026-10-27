package source

import (
	"context"
	"net"

	"github.com/mingcheng/simplyddns"
)

func init() {
	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		return jsonIPByURL(ctx, "https://ipwho.is/?fields=ip,success,message", "ip", map[string]string{
			"Accept":     "application/json",
			"User-Agent": UserAgentCurl,
		}, jsonStatus{path: "success", value: "true"})
	}
	registerHTTPSource("ipwhois", 0, fn)
}

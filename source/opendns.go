package source

import (
	"context"
	"net"

	"github.com/mingcheng/simplyddns"
)

func init() {
	const Name = "opendns"

	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		return jsonIPByURL(ctx, "https://myipv4.p1.opendns.com/get_my_ip", "ip", map[string]string{
			"Accept":     "application/json",
			"User-Agent": UserAgent,
		})
	}

	registerSource(Name, fn)
}

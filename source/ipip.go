package source

import (
	"context"
	"net"

	"github.com/mingcheng/simplyddns"
)

func init() {
	const Name = "ipip"

	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		return jsonIPByURL(ctx, "https://myip.ipip.net/json", "data.ip", map[string]string{
			"Accept": "application/json",
		})
	}

	registerSource(Name, fn)
}

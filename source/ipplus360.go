package source

import (
	"context"
	"net"

	"github.com/mingcheng/simplyddns"
)

func init() {
	const Name = "ipplus360"

	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		return jsonIPByURL(ctx, "https://www.ipplus360.com/getIP", "data", map[string]string{
			"Accept": "application/json",
		})
	}

	registerSource(Name, fn)
}

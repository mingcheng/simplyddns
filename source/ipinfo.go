package source

import (
	"context"
	"github.com/mingcheng/simplyddns"
	"net"

	"github.com/ipinfo/go/v2/ipinfo"
)

func init() {
	const (
		Name = "ipinfo"
	)

	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		result, err := ipinfo.GetIPAddr()
		if err != nil {
			return nil, err
		}

		ip := net.ParseIP(result)
		return &ip, nil
	}

	_ = simplyddns.RegisterSourceFunc(Name, fn)
}

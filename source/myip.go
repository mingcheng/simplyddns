package source

import (
	"context"
	"github.com/mingcheng/simplyddns"
	"net"
)

func init() {
	const Name = "myip"

	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		if data, err := RawIPByURL("https://api.my-ip.io/ip.txt"); err != nil {
			return nil, err
		} else {
			return &data, nil
		}
	}

	_ = simplyddns.RegisterSourceFunc(Name, fn)
}

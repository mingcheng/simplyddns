package source

import (
	"context"
	"github.com/mingcheng/simplyddns"
	"net"
)

func init() {
	const Name = "ipezy"

	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		if data, err := RawIPByURL("https://api.ipezy.com"); err != nil {
			return nil, err
		} else {
			return &data, nil
		}
	}

	_ = simplyddns.RegisterSourceFunc(Name, fn)
}

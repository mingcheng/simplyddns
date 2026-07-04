package source

import (
	"context"
	"fmt"
	"net"

	"github.com/mingcheng/simplyddns"
	"github.com/tidwall/gjson"
)

func init() {
	const Name = "ipip"

	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		data, err := RawStrByURL(ctx, "https://myip.ipip.net/json", map[string]string{
			"Accept": "application/json",
		})

		if err != nil {
			return nil, err
		}

		if result := gjson.Get(data, "data.ip").Str; result != "" {
			ip := net.ParseIP(result)
			return &ip, nil
		}

		return nil, fmt.Errorf("can not found address from %s", Name)
	}

	_ = simplyddns.RegisterSourceFunc(Name, fn)
}

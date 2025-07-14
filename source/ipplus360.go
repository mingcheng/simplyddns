package source

import (
	"context"
	"fmt"
	"net"

	"github.com/mingcheng/simplyddns"
	"github.com/tidwall/gjson"
)

func init() {
	const Name = "ipplus360"

	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		data, err := RawStrByURL(ctx, "https://www.ipplus360.com/getIP", map[string]string{
			"Accept": "application/json",
		})

		if err != nil {
			return nil, err
		}

		if result := gjson.Get(data, "data").Str; result != "" {
			ip := net.ParseIP(result)
			return &ip, nil
		}

		return nil, fmt.Errorf("can not found address from %s", Name)
	}

	_ = simplyddns.RegisterSourceFunc(Name, fn)
}

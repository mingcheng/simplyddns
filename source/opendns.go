package source

import (
	"context"
	"fmt"
	"net"

	"github.com/mingcheng/simplyddns"
	"github.com/tidwall/gjson"
)

//ip=103.135.249.59

func init() {
	const Name = "opendns"

	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {

		data, err := RawStrByURL(ctx, "https://myipv4.p1.opendns.com/get_my_ip", map[string]string{
			"Accept":     "application/json",
			"User-Agent": UserAgent,
		})

		if err != nil {
			return nil, err
		}

		if result := gjson.Get(data, "ip").Str; result != "" {
			ip := net.ParseIP(result)
			return &ip, nil
		}

		return nil, fmt.Errorf("can not found address from %s", Name)
	}

	_ = simplyddns.RegisterSourceFunc(Name, fn)
}

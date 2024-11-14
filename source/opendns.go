package source

import (
	"context"
	"fmt"
	"github.com/mingcheng/simplyddns"
	"github.com/tidwall/gjson"
	"net"
)

//ip=103.135.249.59

func init() {
	const Name = "opendns"

	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {

		data, err := RawStrByURL(context.Background(), "https://myipv4.p1.opendns.com/get_my_ip", map[string]string{
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

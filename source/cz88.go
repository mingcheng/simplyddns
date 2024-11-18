package source

import (
	"context"
	"fmt"
	"github.com/mingcheng/simplyddns"
	"github.com/tidwall/gjson"
	"net"
	"net/http"
)

func init() {
	const (
		Name = "cz88"
	)

	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		data, err := RawStrByURL(context.Background(), `https://update.cz88.net/api/cz88/ip/geo?ip=`, map[string]string{
			"Accept":     "application/json",
			"Referer":    "https://update.cz88.net/geo",
			"User-Agent": UserAgentCurl,
		})

		if err != nil {
			return nil, err
		}

		statusCode := gjson.Get(data, "code").Int()
		if statusCode != http.StatusOK {
			return nil, fmt.Errorf("status code is %d", statusCode)
		}

		if result := gjson.Get(data, "data.ip").Str; result != "" {
			ip := net.ParseIP(result)
			return &ip, nil
		}

		return nil, fmt.Errorf("can not found address from %s", Name)
	}

	_ = simplyddns.RegisterSourceFunc(Name, fn)
}

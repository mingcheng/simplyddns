package source

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/mingcheng/simplyddns"
	"github.com/tidwall/gjson"
)

func init() {
	const (
		Name = "cz88"
	)

	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		data, err := RawStrByURL(ctx, `https://update.cz88.net/api/cz88/ip/geo?ip=`, map[string]string{
			"Accept":     "application/json",
			"Referer":    "https://update.cz88.net/geo",
			"User-Agent": UserAgentCurl,
		})

		if err != nil {
			return nil, err
		}
		if !gjson.Valid(data) {
			return nil, fmt.Errorf("cz88 returned invalid JSON")
		}

		statusCode := gjson.Get(data, "code").Int()
		if statusCode != http.StatusOK {
			return nil, fmt.Errorf("status code is %d", statusCode)
		}

		return parseIP(gjson.Get(data, "data.ip").Str)
	}

	registerSource(Name, fn)
}

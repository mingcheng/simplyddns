package source

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mingcheng/simplyddns"
	"net"
)

func init() {
	const (
		Name  = "ipinfo"
		Token = "8fe8bfdbb0d459"
	)

	type Result struct {
		IP string `json:"ip"`
	}

	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		data, err := RawStrByURL(context.Background(), "https://ipinfo.io", map[string]string{
			"Accept":        "application/json",
			"Authorization": fmt.Sprintf("Bearer %s", Token),
		})

		if err != nil {
			return nil, err
		}

		result := Result{}
		err = json.Unmarshal([]byte(data), &result)
		if err != nil {
			return nil, err
		}

		ip := net.ParseIP(result.IP)
		return &ip, nil
	}

	_ = simplyddns.RegisterSourceFunc(Name, fn)
}

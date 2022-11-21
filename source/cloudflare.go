package source

import (
	"context"
	"github.com/mingcheng/simplyddns"
	"net"
	"regexp"
)

//ip=103.135.249.59

func init() {
	const Name = "cloudflare"

	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		data, err := RawStrByURL(ctx, "https://www.cloudflare.com/cdn-cgi/trace", nil)
		if err != nil {
			return nil, err
		}

		var re = regexp.MustCompile(`(?m)ip=([\d|\.]+)`)

		for _, match := range re.FindAllStringSubmatch(data, -1) {
			if match[1] != "" {
				addr := net.ParseIP(match[1])
				return &addr, nil
			}
		}

		return nil, nil
	}

	_ = simplyddns.RegisterSourceFunc(Name, fn)
}

package source

import (
	"context"
	"net"

	"github.com/mingcheng/simplyddns"
)

func init() {
	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		return jsonIPByURL(ctx, "https://test.ustc.edu.cn/backend/getIP.php", "processedString", map[string]string{
			"Accept":     "application/json",
			"User-Agent": UserAgentCurl,
		})
	}
	registerHTTPSource("ustc", 0, fn)
}

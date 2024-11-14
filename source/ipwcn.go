/**
 * File: ipwcn.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Saturday, December 26th 2020, 10:41:38 pm
 * Last Modified: Sunday, December 27th 2020, 7:35:13 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package source

import (
	"context"
	"github.com/mingcheng/simplyddns"
	"net"
)

func init() {
	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		log.Debugf("start requests from %s", "4.ipw.cn")
		resp, err := RawStrByURL(context.Background(), "https://4.ipw.cn/", map[string]string{
			"User-Agent": UserAgentCurl,
		})
		if err != nil {
			log.Debug(err)
			return nil, err
		}

		if err != nil {
			log.Debug(err)
			return nil, err
		}

		addr := net.ParseIP(resp)
		log.Debugf("remote address is %s", addr.String())
		return &addr, nil
	}

	_ = simplyddns.RegisterSourceFunc("ipwcn", fn)
}

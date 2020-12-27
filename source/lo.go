/**
 * File: lo.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Saturday, December 26th 2020, 3:01:39 pm
 * Last Modified: Sunday, December 27th 2020, 8:38:46 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package source

import (
	"context"
	"fmt"
	"net"

	ddns "github.com/mingcheng/simplyddns"
)

func init() {
	fn := func(_ context.Context, _ *ddns.JobSource) (*net.IP, error) {
		ip := net.ParseIP("127.0.0.1")
		if !ip.IsLoopback() {
			return nil, fmt.Errorf("%v is not lookback address", ip)
		}

		return &ip, nil
	}

	_ = ddns.RegisterSourceFunc("lo", fn)
}

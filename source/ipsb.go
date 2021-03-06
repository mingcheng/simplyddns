/**
 * File: myip.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Saturday, December 26th 2020, 10:41:38 pm
 * Last Modified: Saturday, February 13th 2021, 10:39:15 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package source

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/mingcheng/simplyddns"
)

func init() {
	const Name = "ip.sb"
	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {
		log.Debugf("%s start requests", Name)
		resp, err := http.Get("https://api.ip.sb/ip")
		if err != nil || resp.StatusCode != http.StatusOK {
			log.Debug(err)
			return nil, err
		}

		if respStr, err := ioutil.ReadAll(resp.Body); err != nil {
			log.Debug(err)
			return nil, err
		} else {
			ipAddress := strings.TrimSpace(string(respStr))
			addr := net.ParseIP(ipAddress)
			if len(ipAddress) <= 0 || addr == nil {
				return nil, fmt.Errorf("error parse response address %s", ipAddress)
			}

			log.Debugf("%s remote address is %s", Name, addr.String())
			return &addr, nil
		}
	}

	_ = simplyddns.RegisterSourceFunc("ipsb", fn)
}

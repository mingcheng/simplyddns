/**
 * File: myip.go
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
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/mingcheng/simplyddns"
	"github.com/valyala/fastjson"
)

func init() {
	const Name = "MyIPIP.net"
	fn := func(ctx context.Context, _ *simplyddns.JobSource) (*net.IP, error) {
		log.Debugf("%s start requests", Name)
		resp, err := http.Get("https://myip.ipip.net/json")
		if err != nil || resp.StatusCode != http.StatusOK {
			log.Debug(err)
			return nil, err
		}

		if jsonStr, err := ioutil.ReadAll(resp.Body); err != nil {
			log.Debug(err)
			return nil, err
		} else {
			// https://github.com/valyala/fastjson
			if ip := fastjson.GetString(jsonStr, "data", "ip"); ip != "" {
				addr := net.ParseIP(ip)
				log.Debugf("%s remote address is %s", Name, addr.String())
				return &addr, nil
			}
		}

		return nil, fmt.Errorf("canot get IP address from ipip.net")
	}

	_ = simplyddns.RegisterSourceFunc("myipip", fn)
}

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
	"io/ioutil"
	"net"
	"net/http"

	"github.com/mingcheng/simplyddns"
)

func init() {
	const Name = "whatismyipaddress.com"
	fn := func(ctx context.Context, _ *simplyddns.SourceConfig) (*net.IP, error) {

		log.Debugf("%s start requests", Name)
		resp, err := http.Get("https://ipv4bot.whatismyipaddress.com")
		if err != nil || resp.StatusCode != http.StatusOK {
			log.Debug(err)
			return nil, err
		}

		ip, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			log.Debug(err)
			return nil, err
		}

		addr := net.ParseIP(string(ip))
		log.Debugf("%s remote address is %s", Name, addr.String())
		return &addr, nil
	}

	_ = simplyddns.RegisterSourceFunc("whatis", fn)
}

/**
 * File: static.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Sunday, December 27th 2020, 2:07:45 pm
 * Last Modified: Sunday, December 27th 2020, 2:09:34 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package source

import (
	"context"
	ddns "github.com/mingcheng/simplyddns"
	"net"
	"strings"
)

func init() {
	fn := func(_ context.Context, source *ddns.SourceConfig) (*net.IP, error) {
		addr := net.ParseIP(strings.TrimSpace(source.Content))
		log.Debugf("parsed address is %s", addr.String())
		return &addr, nil
	}

	_ = ddns.RegisterSourceFunc("static", fn)
}

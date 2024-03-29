/**
 * File: sleep.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Saturday, December 26th 2020, 3:01:44 pm
 * Last Modified: Sunday, December 27th 2020, 8:40:58 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package target

import (
	"context"
	"net"
	"time"

	ddns "github.com/mingcheng/simplyddns"
)

func init() {
	log.Tracef("register target function which name is sleep")

	_ = ddns.RegisterTargetFunc("sleep", func(_ context.Context, addr *net.IP, _ *ddns.TargetConfig) error {
		log.Debugf("sleep target, recive address %s", addr.String())
		time.Sleep(10 * time.Second)
		return nil
	})
}

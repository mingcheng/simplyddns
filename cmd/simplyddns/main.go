/**
 * File: main.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Friday, December 25th 2020, 9:43:56 pm
 * Last Modified: Sunday, December 27th 2020, 8:38:36 pm
 *
 * http://www.opensource.org/licenses/MIT
 */

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	ddns "github.com/mingcheng/simplyddns"
	_ "github.com/mingcheng/simplyddns/source"
	_ "github.com/mingcheng/simplyddns/target"
	"github.com/sirupsen/logrus"
)

const AppName = "simplyddns"

var (
	BuildTime    = "unknown"
	BuildVersion = "unknown"
)

var (
	configure ddns.Config
	log       = ddns.NewLogger()
)

func init() {
	log.Printf("%s %s, %s", AppName, BuildVersion, BuildTime)
}

func main() {
	if err := ReadConfigure(&configure); err != nil {
		log.Panic(err)
	}

	if configure.LogFile != "" {
		fp, err := os.OpenFile(configure.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			log.Panic(err)
		}

		log.Out = fp
		defer fp.Close()
	}

	if configure.Debug {
		log.SetLevel(logrus.DebugLevel)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dispatch, err := ddns.NewDispatch(configure.Jobs)
	if err != nil {
		log.Panic(err)
	}

	// waiting for stop
	go func() {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Kill, os.Interrupt, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
		<-interrupt

		log.Debugf("stop disaptch")
		dispatch.Stop()
	}()

	log.Debugf("start dispatch")
	dispatch.Start(ctx)
}

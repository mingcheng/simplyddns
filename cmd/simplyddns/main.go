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
	"os"
	"syscall"

	"github.com/judwhite/go-svc"
	"github.com/mingcheng/simplyddns"
	_ "github.com/mingcheng/simplyddns/source"
	_ "github.com/mingcheng/simplyddns/target"
)

const AppName = "simplyddns"

var (
	BuildTime    = "unknown"
	BuildVersion = "unknown"
	BuildCommit  = "unknown"
	log          = simplyddns.NewLogger()
)

func main() {
	prg := &Program{
		Configure: &simplyddns.Config{},
	}

	// Call svc.Run to start your Program/service.
	if err := svc.Run(prg, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill); err != nil {
		log.Fatal(err)
	}
}

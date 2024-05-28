/*!*
 * Copyright (c) 2025 Hangzhou Guanwaii Technology Co,.Ltd.
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: main.go
 * Author: mingcheng@outlook.com
 * File Created: Friday, December 25th 2020, 9:43:56 pm
 *
 * Modified By: mingcheng (mingcheng@apache.org)
 * Last Modified: 2025-03-12 14:35:10
 */

// simplyddns is a simple dynamic DNS client
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

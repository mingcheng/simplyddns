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
	"errors"
	"flag"
	"fmt"
	"io"
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
	if err := run(os.Args[1:], os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func run(args []string, output io.Writer) error {
	flags := flag.NewFlagSet(AppName, flag.ContinueOnError)
	flags.SetOutput(output)
	version := flags.Bool("version", false, "print version and exit")
	configFile := flags.String("config", "", "path to a YAML configuration file")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %v", flags.Args())
	}
	if *version {
		_, err := fmt.Fprintf(output, "%s %s (commit %s, built %s)\n", AppName, BuildVersion, BuildCommit, BuildTime)
		return err
	}
	prg := &Program{Configure: &simplyddns.Config{}, ConfigFile: *configFile}
	return svc.Run(prg, syscall.SIGINT, syscall.SIGTERM)
}

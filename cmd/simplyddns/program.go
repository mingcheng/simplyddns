/*!*
 * Copyright (c) 2026 Ming Lyu, aka mingcheng
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: program.go
 * Author: mingcheng (mingcheng@apache.org)
 * File Created: Friday, December 25th 2020, 9:43:56 pm
 *
 * Modified By: mingcheng <mingcheng@apache.org>
 * Last Modified: 2026-07-04 22:32:16
 */

package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/judwhite/go-svc"
	"github.com/mingcheng/simplyddns"
	"github.com/sirupsen/logrus"
)

// Program interface for simplyddns
type Program struct {
	Configure  *simplyddns.Config
	ConfigFile string
	Dispatch   *simplyddns.Dispatch
	logFile    *os.File
}

// Init Program by svc library
func (p *Program) Init(env svc.Environment) error {
	log.Printf("%s %s, build on %s(%s)", AppName, BuildVersion, BuildTime, BuildCommit)

	log.Printf("supported source funcs is [%s]",
		strings.Join(simplyddns.GetAllSupportSourceFunc(), ", "))

	// read configure from file
	if err := readConfigure(p.Configure, p.ConfigFile); err != nil {
		return err
	}

	switch p.Configure.LogFile {
	case "", "/dev/stderr":
		log.SetOutput(os.Stderr)
	case "/dev/stdout":
		log.SetOutput(os.Stdout)
	default:
		fp, err := os.OpenFile(p.Configure.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}

		p.logFile = fp
		log.SetOutput(fp)
	}

	// detect if debug mode is on
	if p.Configure.Debug {
		log.SetLevel(logrus.DebugLevel)
		log.Debugf("set %v to debug mode", AppName)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	return nil
}

// Start the Program by svc library
func (p *Program) Start() error {
	if p.Configure == nil {
		return fmt.Errorf("program configuration is nil")
	}
	var err error

	p.Dispatch, err = simplyddns.NewDispatch(p.Configure.Jobs)
	if err != nil {
		return err
	}

	// start the dispatch
	go p.Dispatch.Start(context.Background())

	return nil
}

// Stop Program trigger by svc library
func (p *Program) Stop() error {
	log.Debug("stopping Program, bye~")
	if p.Dispatch != nil {
		p.Dispatch.Stop()
	}
	if p.logFile != nil {
		log.SetOutput(os.Stderr)
		err := p.logFile.Close()
		p.logFile = nil
		return err
	}
	return nil
}

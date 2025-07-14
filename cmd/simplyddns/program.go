/*!*
 * Copyright (c) 2025 Hangzhou Guanwaii Technology Co,.Ltd.
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: program.go
 * Author: mingcheng (mingcheng@apache.org)
 * File Created: Friday, December 25th 2020, 9:43:56 pm
 *
 * Modified By: mingcheng (mingcheng@apache.org)
 * Last Modified: 2025-03-12 14:35:38
 */

package main

import (
	"context"
	"os"
	"strings"

	"github.com/judwhite/go-svc"
	"github.com/mingcheng/simplyddns"
	"github.com/sirupsen/logrus"
)

// Program interface for simplyddns
type Program struct {
	Configure *simplyddns.Config
	Dispatch  *simplyddns.Dispatch
}

// Init Program by svc library
func (p *Program) Init(env svc.Environment) error {
	log.Printf("%s %s, build on %s(%s)", AppName, BuildVersion, BuildTime, BuildCommit)

	log.Printf("supported source funcs is [%s]",
		strings.Join(simplyddns.GetAllSupportSourceFunc(), ","))

	// read configure from file
	if err := ReadConfigure(p.Configure); err != nil {
		return err
	}

	// write to local file
	if p.Configure.LogFile != "" {
		log.Debugf("read configure file from %s", p.Configure.LogFile)
		fp, err := os.OpenFile(p.Configure.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			return err
		}

		log.Out = fp
	}

	// detect if debug mode is on
	if p.Configure.Debug {
		log.SetLevel(logrus.DebugLevel)
		log.Debugf("set %v to debug mode", AppName)
	}

	return nil
}

// Start the Program by svc library
func (p *Program) Start() error {
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

	return nil
}

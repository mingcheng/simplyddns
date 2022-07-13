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
	log.Printf("%s %s, %s", AppName, BuildVersion, BuildTime)

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
	p.Dispatch.Stop()

	return nil
}

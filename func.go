package simplyddns

import (
	"context"
	"fmt"
	"net"
)

type SourceFunc func(context.Context, *SourceConfig) (*net.IP, error)
type TargetFunc func(context.Context, *net.IP, *TargetConfig) error

var (
	sourceFuncs map[string]SourceFunc
	targetFuncs map[string]TargetFunc
)

func init() {
	sourceFuncs = make(map[string]SourceFunc)
	targetFuncs = make(map[string]TargetFunc)
}

func GetAllSupportSourceFunc() []string {
	var funcNames []string
	for k, _ := range sourceFuncs {
		funcNames = append(funcNames, k)
	}

	return funcNames
}

func SourceFuncByName(name string) (fn SourceFunc, err error) {
	if fn = sourceFuncs[name]; fn == nil {
		err = fmt.Errorf("the source function which name %s is not found", name)
		return
	}

	return
}

func RegisterSourceFunc(name string, fn SourceFunc) (err error) {
	if found, _ := SourceFuncByName(name); found != nil {
		return fmt.Errorf("source func %s is already registered", name)
	}

	sourceFuncs[name] = fn
	return nil
}

func TargetFuncByName(name string) (fn TargetFunc, err error) {
	if fn = targetFuncs[name]; fn == nil {
		err = fmt.Errorf("the target function which name %s is not found", name)
		return
	}

	return
}

func RegisterTargetFunc(name string, fn TargetFunc) (err error) {
	if found, _ := TargetFuncByName(name); found != nil {
		return fmt.Errorf("target func %s is already registered", name)
	}

	targetFuncs[name] = fn
	return nil
}

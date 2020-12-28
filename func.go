package simplyddns

import (
	"context"
	"fmt"
	"net"
)

var (
	sourceFunc = map[string]func(context.Context, *SourceConfig) (*net.IP, error){}
	targetFunc = map[string]func(context.Context, *net.IP, *TargetConfig) error{}
)

func SourceFunc(name string) (func(context.Context, *SourceConfig) (*net.IP, error), error) {
	if fn := sourceFunc[name]; fn != nil {
		return fn, nil
	}

	return nil, fmt.Errorf("source func %s is not found", name)
}

func RegisterSourceFunc(name string, fn func(context.Context, *SourceConfig) (*net.IP, error)) error {
	if fn, _ := SourceFunc(name); fn != nil {
		return fmt.Errorf("source func %s is already registered", name)
	}

	sourceFunc[name] = fn
	return nil
}

func TargetFunc(name string) (func(context.Context, *net.IP, *TargetConfig) error, error) {
	if fn := targetFunc[name]; fn != nil {
		return fn, nil
	}

	return nil, fmt.Errorf("target func %s is not found", name)
}

func RegisterTargetFunc(name string, fn func(context.Context, *net.IP, *TargetConfig) error) error {
	if fn, _ := TargetFunc(name); fn != nil {
		return fmt.Errorf("target func %s is already registered", name)
	}

	targetFunc[name] = fn
	return nil
}

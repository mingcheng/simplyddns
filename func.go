/*!*
 * Copyright (c) 2025 Hangzhou Guanwaii Technology Co,.Ltd.
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: func.go
 * Author: mingcheng (mingcheng@apache.org)
 * File Created: 2025-03-12 13:23:45
 *
 * Modified By: mingcheng (mingcheng@apache.org)
 * Last Modified: 2025-03-12 13:38:53
 */

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
	for k := range sourceFuncs {
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

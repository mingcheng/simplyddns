/*!*
 * Copyright (c) 2025-2026 Ming Lyu, aka mingcheng
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: func.go
 * Author: mingcheng <mingcheng@apache.org>
 * File Created: 2025-03-12 13:23:45
 *
 * Modified By: mingcheng <mingcheng@apache.org>
 * Last Modified: 2026-05-12 12:23:24
 */

package simplyddns

import (
	"context"
	"fmt"
	"net"
)

// SourceFunc resolves the current IP address from a configured source.
type SourceFunc func(context.Context, *SourceConfig) (*net.IP, error)

// TargetFunc applies the given IP address to a configured target.
type TargetFunc func(context.Context, *net.IP, *TargetConfig) error

var (
	sourceFuncs = map[string]SourceFunc{}
	targetFuncs = map[string]TargetFunc{}
)

// GetAllSupportSourceFunc returns the names of every registered source function.
func GetAllSupportSourceFunc() []string {
	names := make([]string, 0, len(sourceFuncs))
	for k := range sourceFuncs {
		names = append(names, k)
	}
	return names
}

// SourceFuncByName looks up a registered source function by name.
func SourceFuncByName(name string) (SourceFunc, error) {
	fn, ok := sourceFuncs[name]
	if !ok {
		return nil, fmt.Errorf("source function %q is not found", name)
	}
	return fn, nil
}

// RegisterSourceFunc registers a new source function under the given name.
// It returns an error if a function with the same name is already registered.
func RegisterSourceFunc(name string, fn SourceFunc) error {
	if _, exists := sourceFuncs[name]; exists {
		return fmt.Errorf("source func %q is already registered", name)
	}
	sourceFuncs[name] = fn
	return nil
}

// TargetFuncByName looks up a registered target function by name.
func TargetFuncByName(name string) (TargetFunc, error) {
	fn, ok := targetFuncs[name]
	if !ok {
		return nil, fmt.Errorf("target function %q is not found", name)
	}
	return fn, nil
}

// RegisterTargetFunc registers a new target function under the given name.
// It returns an error if a function with the same name is already registered.
func RegisterTargetFunc(name string, fn TargetFunc) error {
	if _, exists := targetFuncs[name]; exists {
		return fmt.Errorf("target func %q is already registered", name)
	}
	targetFuncs[name] = fn
	return nil
}

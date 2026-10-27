/*!*
 * Copyright (c) 2025-2026 Ming Lyu, aka mingcheng
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: config.go
 * Author: mingcheng <mingcheng@apache.org>
 * File Created: Saturday, December 26th 2020, 11:39:12 am
 *
 * Modified By: mingcheng <mingcheng@apache.org>
 * Last Modified: 2026-05-12 12:12:51
 */

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mingcheng/simplyddns"
	"github.com/spf13/viper"
)

// ReadConfigure loads the default configuration file.
func ReadConfigure(config *simplyddns.Config) error {
	return readConfigure(config, "")
}

func readConfigure(config *simplyddns.Config, filename string, paths ...string) error {
	if config == nil {
		return fmt.Errorf("configuration destination is nil")
	}
	reader := viper.New()
	reader.SetConfigType("yaml")
	if filename != "" {
		reader.SetConfigFile(filename)
	} else {
		reader.SetConfigName(AppName)
		if len(paths) == 0 {
			paths = append(paths, filepath.Join("/etc", AppName))
			home, err := os.UserHomeDir()
			if err != nil {
				log.Debugf("home configuration directory unavailable: %v", err)
			} else {
				paths = append(paths, filepath.Join(home, ".config", AppName))
			}
			paths = append(paths, ".")
		}
		for _, path := range paths {
			reader.AddConfigPath(path)
		}
	}
	if err := reader.ReadInConfig(); err != nil {
		return fmt.Errorf("read configuration: %w", err)
	}

	var decoded simplyddns.Config
	if err := reader.Unmarshal(&decoded); err != nil {
		return fmt.Errorf("decode configuration: %w", err)
	}
	if len(decoded.Jobs) == 0 {
		return fmt.Errorf("at least one DDNS job is required")
	}
	for i := range decoded.Jobs {
		if err := simplyddns.ValidateConfig(&decoded.Jobs[i]); err != nil {
			return fmt.Errorf("job %d: %w", i+1, err)
		}
	}
	*config = decoded
	return nil
}

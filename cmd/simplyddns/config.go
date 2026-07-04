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

	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigType("yaml")
	viper.SetConfigName(fmt.Sprintf("%s.yml", AppName))

	// configuration lookup paths
	viper.AddConfigPath(fmt.Sprintf("/etc/%s", AppName))
	viper.AddConfigPath(fmt.Sprintf("$HOME/.config/%s", AppName))
	viper.AddConfigPath(".")
}

// ReadConfigure loads the configuration file and decodes it into config.
func ReadConfigure(config interface{}) error {
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return viper.Unmarshal(config)
}

/*!*
 * Copyright (c) 2025 Hangzhou Guanwaii Technology Co,.Ltd.
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: config.go
 * Author: mingcheng@outlook.com
 * File Created: Saturday, December 26th 2020, 11:39:12 am
 *
 * Modified By: mingcheng (mingcheng@apache.org)
 * Last Modified: 2025-03-12 14:36:04
 */

package main

import (
	"fmt"

	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigType("yaml")
	// viper.AllowEmptyEnv(true)
	viper.SetConfigName(fmt.Sprintf("%s.yml", AppName))

	// lookup path
	viper.AddConfigPath(fmt.Sprintf("/etc/%s", AppName))
	viper.AddConfigPath(fmt.Sprintf("$HOME/.config/%s", AppName))
	viper.AddConfigPath(".")
}

func ReadConfigure(config interface{}) error {
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	if err := viper.Unmarshal(config); err != nil {
		return err
	}

	return nil
}

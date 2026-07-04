/*!*
 * Copyright (c) 2025-2026 Ming Lyu, aka mingcheng
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: config.go
 * Author: mingcheng <mingcheng@apache.org>
 * File Created: 2022-07-22 23:37:43
 *
 * Modified By: mingcheng <mingcheng@apache.org>
 * Last Modified: 2026-05-12 12:11:18
 */

package simplyddns

type Auth struct {
	User     string `yaml:"username" mapstructure:"username"`
	Password string `yaml:"password" mapstructure:"password"`
}

type Config struct {
	LogFile  string      `yaml:"logfile" mapstructure:"logfile"`
	Debug    bool        `yaml:"debug" mapstructure:"debug"`
	LoadUI   bool        `yaml:"ui" mapstructure:"ui"`
	BindAddr string      `yaml:"addr" mapstructure:"addr"`
	Auth     Auth        `yaml:"auth" mapstructure:"auth"`
	Jobs     []JobConfig `yaml:"ddns" mapstructure:"ddns"`
}

/*!*
 * Copyright (c) 2022-2025 Hangzhou Guanwaii Technology Co,.Ltd.
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: config.go
 * Author: mingcheng (mingcheng@apache.org)
 * File Created: 2022-07-22 23:37:43
 *
 * Modified By: mingcheng (mingcheng@apache.org)
 * Last Modified: 2025-02-28 10:47:46
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

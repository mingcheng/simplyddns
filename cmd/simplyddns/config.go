/**
 * File: config.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Saturday, December 26th 2020, 11:39:12 am
 * Last Modified: Sunday, December 27th 2020, 8:38:22 pm
 *
 * http://www.opensource.org/licenses/MIT
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

/**
 * File: config.go
 * Author: Ming Cheng<mingcheng@outlook.com>
 *
 * Created Date: Friday, December 25th 2020, 9:58:53 pm
 * Last Modified: Sunday, December 27th 2020, 8:39:35 pm
 *
 * http://www.opensource.org/licenses/MIT
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

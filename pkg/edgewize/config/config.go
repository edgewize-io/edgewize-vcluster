package config

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

const (
	ActionAllow = "allow"
	ActionDeny  = "deny"
)

type Config struct {
	Rules []Rule `yaml:"rules"`
}

type Rule struct {
	Selector string `yaml:"selector"`
	Name     string `yaml:"name"`
	Action   string `yaml:"action"`
}

func (r Rule) DoAction() bool {
	switch r.Action {
	case ActionAllow:
		return true
	case ActionDeny:
		return false
	default:
		return false
	}
}

var Cfg Config
var v *viper.Viper

func init() {
	_ = initConfig()
	parseCfg()
}

func initConfig() error {
	v = viper.New()
	v.SetConfigFile("/manifests/middleware/config.yaml")
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return err
	}
	v.WatchConfig()
	v.OnConfigChange(func(_ fsnotify.Event) {
		if err := v.ReadInConfig(); err != nil {
			return
		}
		parseCfg()
	})
	return nil
}

func parseCfg() {
	if v == nil {
		return
	}
	_ = v.Unmarshal(&Cfg)
	fmt.Println(Cfg)
}

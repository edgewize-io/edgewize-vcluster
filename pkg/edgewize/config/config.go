package config

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Config struct {
	AllowPodSyncDownRule []Rule `yaml:"allowPodSyncDownRule"`
	SkipPodSyncDownRule  []Rule `yaml:"skipPodSyncDownRule"`
}
type Rule struct {
	Selector string `yaml:"selector"`
	Name     string `yaml:"name"`
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

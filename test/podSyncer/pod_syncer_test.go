package podSyncer

import (
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/loft-sh/vcluster/pkg/edgewize/config"
	"github.com/loft-sh/vcluster/pkg/edgewize/utils"
	"github.com/spf13/viper"
	"testing"
)

var Cfg config.Config
var v *viper.Viper

func init() {
	_ = initConfig()
}

func initConfig() error {
	var err error
	v = viper.New()
	v.SetConfigFile("D:\\gopath\\src\\testdemo\\config\\test1_config.yaml")
	v.SetConfigType("yaml")
	err = v.ReadInConfig()
	if err != nil {
		fmt.Println(err)
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
	err := v.Unmarshal(&Cfg)
	fmt.Println(err)
}

func TestRunPodSyncerTests(t *testing.T) {
	parseCfg()
	data := map[string]interface{}{
		"namespace": "kube-system",
		"labels": map[string]string{
			"app.kubernetes.io/name": "router-manager-edge",
		},
	}
	bs, _ := json.Marshal(data)
	for _, r := range Cfg.Rules {
		match := utils.MatchObjectsByFieldSelector(bs, r.Selector)
		if match {
			fmt.Println(fmt.Sprintf("Matched object:%s with action:%s do %v", r.Name, r.Action, r.DoAction()))
		} else {
			fmt.Println(fmt.Sprintf("Not matched object:%s", r.Name))
		}
	}
}

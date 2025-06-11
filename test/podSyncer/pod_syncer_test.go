package podSyncer

import (
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/loft-sh/vcluster/pkg/edgewize/config"
	"github.com/loft-sh/vcluster/pkg/edgewize/utils"
	"github.com/spf13/viper"
	"gotest.tools/v3/assert"
	"testing"
)

var Cfg config.Config
var v *viper.Viper

type PodSyncTestData struct {
	Pod             map[string]interface{}
	MatchRuleName   string
	MatchRuleAction string
}

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

func podSyncTestData() []PodSyncTestData {
	return []PodSyncTestData{
		{
			Pod: map[string]interface{}{
				"namespace": "any-namespace",
				"labels": map[string]string{
					"app.kubernetes.io/instance": "vector-agent",
				},
			},
			MatchRuleName:   "deny vector-agent",
			MatchRuleAction: "deny",
		},
		{
			Pod: map[string]interface{}{
				"namespace": "any-namespace",
				"labels": map[string]string{
					"app.kubernetes.io/name": "router-manager-edge",
				},
			},
			MatchRuleName:   "deny router-manager-edge",
			MatchRuleAction: "deny",
		},
		{
			Pod: map[string]interface{}{
				"namespace": "any-namespace",
				"labels": map[string]string{
					"app": "eventbus",
				},
			},
			MatchRuleName:   "deny eventbus",
			MatchRuleAction: "deny",
		},
		{
			Pod: map[string]interface{}{
				"namespace": "any-namespace",
				"labels": map[string]string{
					"app.kubernetes.io/name": "vector",
				},
			},
			MatchRuleName:   "deny vector",
			MatchRuleAction: "deny",
		},
		{
			Pod: map[string]interface{}{
				"namespace": "any-namespace",
				"labels": map[string]string{
					"app": "edge-prom-agent",
				},
			},
			MatchRuleName:   "deny edge-prom-agent",
			MatchRuleAction: "deny",
		},
		{
			Pod: map[string]interface{}{
				"namespace": "default",
				"labels": map[string]string{
					"some.app": "edge-prom-agent",
				},
			},
			MatchRuleName:   "allow system namespace default",
			MatchRuleAction: "allow",
		},
		{
			Pod: map[string]interface{}{
				"namespace": "edgewize-system",
				"labels": map[string]string{
					"other.app": "edge-prom-agent",
				},
			},
			MatchRuleName:   "allow system namespace edgewize-system",
			MatchRuleAction: "allow",
		},
		{
			Pod: map[string]interface{}{
				"namespace": "extension-whizard-alerting",
				"labels": map[string]string{
					"also.app": "edge-prom-agent",
				},
			},
			MatchRuleName:   "allow system namespace extension-whizard-alerting",
			MatchRuleAction: "allow",
		},
		{
			Pod: map[string]interface{}{
				"namespace": "kube-node-lease",
				"labels": map[string]string{
					"and.app": "edge-prom-agent",
				},
			},
			MatchRuleName:   "allow system namespace kube-node-lease",
			MatchRuleAction: "allow",
		},
		{
			Pod: map[string]interface{}{
				"namespace": "kube-public",
				"labels": map[string]string{
					"always.app": "edge-prom-agent",
				},
			},
			MatchRuleName:   "allow system namespace kube-public",
			MatchRuleAction: "allow",
		},
		{
			Pod: map[string]interface{}{
				"namespace": "kube-system",
				"labels": map[string]string{
					"kube.system.app": "edge-prom-agent",
				},
			},
			MatchRuleName:   "allow system namespace kube-system",
			MatchRuleAction: "allow",
		},
		{
			Pod: map[string]interface{}{
				"namespace": "kubeedge",
				"labels": map[string]string{
					"kubeedge.app": "edge-prom-agent",
				},
			},
			MatchRuleName:   "allow system namespace kubeedge",
			MatchRuleAction: "allow",
		},
		{
			Pod: map[string]interface{}{
				"namespace": "kubesphere-controls-system",
				"labels": map[string]string{
					"kubesphere.app": "edge-prom-agent",
				},
			},
			MatchRuleName:   "allow system namespace kubesphere-controls-system",
			MatchRuleAction: "allow",
		},
		{
			Pod: map[string]interface{}{
				"namespace": "kubesphere-logging-system",
				"labels": map[string]string{
					"kubesphere.app": "edge-prom-agent",
				},
			},
			MatchRuleName:   "allow system namespace kubesphere-logging-system",
			MatchRuleAction: "allow",
		},
		{
			Pod: map[string]interface{}{
				"namespace": "kubesphere-monitoring-system",
				"labels": map[string]string{
					"kubesphere.app": "edge-prom-agent",
				},
			},
			MatchRuleName:   "allow system namespace kubesphere-monitoring-system",
			MatchRuleAction: "allow",
		},
		{
			Pod: map[string]interface{}{
				"namespace": "kubesphere-system",
				"labels": map[string]string{
					"kubesphere.app": "edge-prom-agent",
				},
			},
			MatchRuleName:   "allow system namespace kubesphere-system",
			MatchRuleAction: "allow",
		},
		{
			Pod: map[string]interface{}{
				"namespace": "default",
				"labels": map[string]string{
					"app":                        "edge-prom-agent1",
					"app.kubernetes.io/instance": "vector-agent2",
					"app.kubernetes.io/name":     "router-manager-edge3",
				},
			},
			MatchRuleName:   "allow system namespace default",
			MatchRuleAction: "allow",
		},
		{
			Pod: map[string]interface{}{
				"namespace": "default",
				"labels": map[string]string{
					"app":                        "edge-prom-agent1",
					"app.kubernetes.io/instance": "vector-agent",
					"app.kubernetes.io/name":     "router-manager-edge3",
				},
			},
			MatchRuleName:   "allow system namespace default",
			MatchRuleAction: "allow",
		},
	}
}

func TestRunPodSyncerTests(t *testing.T) {
	parseCfg()
	datas := podSyncTestData()
	for _, data := range datas {
		bs, _ := json.Marshal(data.Pod)
		for _, r := range Cfg.Rules {
			match := utils.MatchObjectsByFieldSelector(bs, r.Selector)
			if match {
				fmt.Println(fmt.Sprintf("Matched object:%s with action:%s do %v", r.Name, r.Action, r.DoAction()))
				assert.Equal(t, data.MatchRuleName, r.Name)
				assert.Equal(t, r.Action, data.MatchRuleAction)
			}
		}
	}

}

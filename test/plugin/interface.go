package plugin

import (
	"context"
	"k8s-leader-election/pkg/server/model"
	"k8s-leader-election/test/plugin/component_a"
	"k8s-leader-election/test/plugin/component_b"
	"k8s-leader-election/test/plugin/component_error"
)

type ComponentPlugin interface {
	IsAvailable() bool
	SetAvailable() error
	Start(ctx context.Context, request *model.Operation) (string, error)
}

var ComponentPluginMap map[string]ComponentPlugin

var (
	_ ComponentPlugin = &component_a.ComponentA{}
	_ ComponentPlugin = &component_b.ComponentB{}
)

func init() {
	// 注册插件
	ComponentPluginMap = make(map[string]ComponentPlugin)
	ComponentPluginMap["component_a"] = component_a.NewComponentA()
	ComponentPluginMap["component_b"] = component_b.NewComponentB()
	ComponentPluginMap["component_error"] = component_error.NewComponentError()
}

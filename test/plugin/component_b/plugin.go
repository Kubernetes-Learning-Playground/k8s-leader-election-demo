package component_b

import (
	"context"
	"k8s-leader-election/pkg/server/model"
	"k8s.io/klog"
	"time"
)

type ComponentB struct {
	Available bool
}

func (c *ComponentB) SetAvailable() error {
	c.Available = true
	return nil
}

func NewComponentB() *ComponentB {
	return &ComponentB{
		Available: false,
	}
}

func (c *ComponentB) Start(ctx context.Context, request *model.Operation) (string, error) {

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	time.Sleep(time.Second * 3)
	klog.Infof("component b do an operation: [Method]: %v, [Url]: %v, [Body]: %v", request.Method, request.Url, request.Body)

	return "successful", nil
}

func (c *ComponentB) IsAvailable() bool {
	return c.Available
}

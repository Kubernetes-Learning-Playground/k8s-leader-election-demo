package component_a

import (
	"context"
	"k8s-leader-election/pkg/server/model"
	"k8s.io/klog"
	"time"
)

type ComponentA struct {
	Available bool
}

func (c *ComponentA) SetAvailable() error {
	c.Available = true
	return nil
}

func NewComponentA() *ComponentA {
	return &ComponentA{
		Available: false,
	}
}

func (c *ComponentA) Start(ctx context.Context, request *model.Operation) (string, error) {

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	time.Sleep(time.Second * 2)
	klog.Infof("component a do an operation: [Method]: %v, [Url]: %v, [Body]: %v", request.Method, request.Url, request.Body)

	return "successful", nil
}

func (c *ComponentA) IsAvailable() bool {
	return c.Available
}

package component_error

import (
	"context"
	"k8s-leader-election/pkg/server/model"
)

// ComponentError 兜底插件
type ComponentError struct {
	// 本身就默认开启
	Available bool
}

func (c *ComponentError) SetAvailable() error {
	return nil
}

func NewComponentError() *ComponentError {
	return &ComponentError{
		Available: true,
	}
}

func (c *ComponentError) Start(_ context.Context, request *model.Operation) (string, error) {
	return "plugin not found", nil
}

func (c *ComponentError) IsAvailable() bool {
	return c.Available
}

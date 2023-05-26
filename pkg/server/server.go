package server

import (
	"context"
	"fmt"
	"k8s-leader-election/pkg/server/handler"
	"k8s.io/klog"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

type ServerOptions struct {
	Port       int
	HealthPort int
}

// Run 执行
func Run(ctx context.Context, options *ServerOptions) {

	// 心跳检测健康机制
	go func() {
		h := &healthz.Handler{
			Checks: map[string]healthz.Checker{
				"healthz": healthz.Ping,
			},
		}
		klog.Info(fmt.Sprintf("0.0.0.0:%d", options.HealthPort))
		if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", options.HealthPort), h); err != nil {
			klog.Fatalf("Failed to start healthz endpoint: %v", err)
		}
	}()

	http.HandleFunc("/test", handler.Test)
	http.HandleFunc("/ws/echo/", handler.Echo)
	http.HandleFunc("/send", handler.Send)
	klog.Info(fmt.Sprintf(":%v", options.Port))
	http.ListenAndServe(fmt.Sprintf(":%v", options.Port), nil)
}

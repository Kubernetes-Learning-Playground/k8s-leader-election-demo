package main

import (
	"context"
	"flag"
	"k8s-leader-election/pkg/cleanup"
	config2 "k8s-leader-election/pkg/configs"
	"k8s-leader-election/pkg/leaselock"
	"k8s-leader-election/pkg/server"
	"k8s-leader-election/pkg/signals"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/klog"
	"os"
	"time"
)

var (
	client *clientset.Clientset
)



/*
 使用 gin http server 启动test接口返回pod名称

 场景：模拟服务挂掉后还会有其他pod接受请求的场景，主要模拟有状态服务
*/

func main() {
	var (
		leaseLockName      string // 锁名
		leaseLockNamespace string // 获取锁的namespace
		leaseLockMode      bool   // 是否为选主模式
		debugMode          bool
		port 			   int
		healthPort         int
		podName            = os.Getenv("POD_NAME") // 需要取到pod name
	)
	flag.StringVar(&leaseLockName, "lease-name", "lease-default-name", "election lease leaselock name")
	flag.BoolVar(&leaseLockMode, "lease-mode", true, "Whether to use election mode")
	flag.BoolVar(&debugMode, "debug-mode", false, "Whether to use debug mode")
	flag.StringVar(&leaseLockNamespace, "lease-namespace", "default", "election lease leaselock namespace")
	flag.IntVar(&port, "server-port", 8888, "")
	flag.IntVar(&healthPort, "health-check-port", 9999, "")
	flag.Parse()

	opt := &server.ServerOptions{
		Port: port,
		HealthPort: healthPort,
	}

	// clientSet
	var config *rest.Config
	if debugMode {
		// 本地debug使用
		c := config2.K8sConfig{}
		config = c.K8sRestConfig()
	} else {
		config, _ = rest.InClusterConfig()
	}

	client = clientset.NewForConfigOrDie(config)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-signals.SetupSignalHandler()
		cancel()
	}()
	defer cancel()

	if leaseLockMode {
		lock := leaselock.GetNewLock(leaseLockName, podName, leaseLockNamespace, client)
		// 选主模式
		leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
			Lock:            lock,
			ReleaseOnCancel: true,
			LeaseDuration:   15 * time.Second, // 租约时长，follower用来判断集群锁是否过期
			RenewDeadline:   10 * time.Second, // leader更新锁的时长
			RetryPeriod:     2 * time.Second,  // 重试获取锁的间隔
			// 当发生不同选主事件时的回调方法
			Callbacks: leaderelection.LeaderCallbacks{
				// 成为leader时，需要执行的回调
				OnStartedLeading: func(c context.Context) {
					// 执行server逻辑
					klog.Info("leader election server running...")
					server.Run(c, opt)
				},
				// 不是leader时，需要执行的回调
				OnStoppedLeading: func() {
					klog.Info("no longer a leader...")
					klog.Info("clean up server...")
					// 如果有退出逻辑可以在此执行
					cleanup.CleanUp()
				},
				// 当产生新leader时，执行的回调
				OnNewLeader: func(currentId string) {
					if currentId == podName {
						klog.Info("still the leader!")
						return
					}
					klog.Infof("new leader is %v", currentId)
				},
			},
		})
	} else {
		// 一般模式
		klog.Info("server running...")
		server.Run(ctx, opt)
	}

}

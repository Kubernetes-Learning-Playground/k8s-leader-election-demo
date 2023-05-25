package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"k8s-leader-election/pkg/cleanup"
	"k8s-leader-election/pkg/signals"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

var (
	client *clientset.Clientset
)

// getNewLock 创建集群锁资源
func getNewLock(lockname, podname, namespace string) *resourcelock.LeaseLock {
	return &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      lockname,
			Namespace: namespace,
		},
		Client: client.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: podname,
		},
	}
}

/*
	 使用 gin http server 启动test接口返回pod名称

	 场景：模拟服务挂掉后还会有其他pod接受请求的场景，主要模拟有状态服务
*/

func test(w http.ResponseWriter, req *http.Request) {
	fmt.Println("test server")
	fmt.Println("pod name: ", os.Getenv("POD_NAME"))
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf("pod name: %v\n", os.Getenv("POD_NAME"))))
}

// Run 执行
func Run() {

	// 心跳检测健康机制
	go func() {
		handler := &healthz.Handler{
			Checks: map[string]healthz.Checker{
				"healthz": healthz.Ping,
			},
		}
		if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", 9999), handler); err != nil {
			klog.Fatalf("Failed to start healthz endpoint: %v", err)
		}
	}()

	http.HandleFunc("/test", test)
	http.ListenAndServe(fmt.Sprintf(":%v", "8080"), nil)
}


func main() {
	var (
		leaseLockName      string                  // 锁名
		leaseLockNamespace string                  // 获取锁的namespace
		leaseLockMode      bool					   // 是否为选主模式
		podName            = os.Getenv("POD_NAME") // 需要取到pod name
	)
	flag.StringVar(&leaseLockName, "lease-name", "d", "election lease lock name")
	flag.BoolVar(&leaseLockMode, "lease-mode", true, "Whether to use election mode")
	flag.StringVar(&leaseLockNamespace, "lease-namespace", "default", "election lease lock namespace")
	flag.Parse()

	// clientSet
	config, err := rest.InClusterConfig()
	client = clientset.NewForConfigOrDie(config)

	if err != nil {
		klog.Fatalf("failed to get kubeconfig")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if leaseLockMode {
		lock := getNewLock(leaseLockName, podName, leaseLockNamespace)
		// 选主模式
		leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
			Lock:            lock,
			ReleaseOnCancel: true,
			LeaseDuration:   15 * time.Second,
			RenewDeadline:   10 * time.Second,
			RetryPeriod:     2 * time.Second,
			Callbacks: leaderelection.LeaderCallbacks{
				OnStartedLeading: func(c context.Context) {

					_, cancel := context.WithCancel(context.Background())
					go func() {
						<-signals.SetupSignalHandler()
						cancel()
					}()
					// 执行server逻辑
					Run()
				},
				OnStoppedLeading: func() {
					klog.Info("no longer a leader.")
					// 如果有退出逻辑可以在此执行
					cleanup.CleanUp()
				},
				OnNewLeader: func(current_id string) {
					if current_id == podName {
						klog.Info("still the leader!")
						return
					}
					klog.Infof("new leader is %v", current_id)
				},
			},
		})
	} else {
		// 一般模式
		Run()
	}

}



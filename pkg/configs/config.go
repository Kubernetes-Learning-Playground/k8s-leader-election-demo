package configs

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
)

type K8sConfig struct{}

func getWd() string {
	wd := os.Getenv("WORK_DIR")
	if wd == "" {
		wd, _ = os.Getwd()
	}
	return wd
}

// K8sRestConfig 默认读取项目根目录的config文件
func (*K8sConfig) K8sRestConfig() *rest.Config {
	path := getWd()
	config, err := clientcmd.BuildConfigFromFlags("", path+"/configs")
	config.Insecure = true // 不使用认证的方式
	if err != nil {
		log.Fatal(err)
	}
	return config
}

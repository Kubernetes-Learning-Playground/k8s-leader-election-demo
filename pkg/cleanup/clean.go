package cleanup

import "k8s.io/klog"

func CleanUp() {
	klog.Infof("If there is any resource cleanup or exit operation, it can be executed here")
}

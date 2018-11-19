package hookexecutor

import (
	"k8s.io/client-go/kubernetes"
)

type Executor interface {
	Execute() error
}

type baseExcutor struct {
	kubeClient kubernetes.Interface
}

func PostInstallExecutor(kubeClient kubernetes.Interface) Executor {
	return newPostInstallExecutor(kubeClient)
}

func PreDeleteExecutor(kubeClient kubernetes.Interface) Executor {
	return newPreDeleteExecutor(kubeClient)
}

func SanityExecutor(kubeClient kubernetes.Interface) Executor {
	return newSanityExecutor(kubeClient)
}

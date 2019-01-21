package main

import (
	"context"

	"github.com/IBM/ubiquity-k8s/sidecars/flex"
	utilsk8s "github.com/IBM/ubiquity-k8s/utils/kubernetes"
	"k8s.io/client-go/kubernetes"
)

func main() {
	clientset := getClientset()
	ctx, _ := context.WithCancel(context.Background())
	s, err := flex.NewServiceSyncer(clientset, ctx)
	if err != nil {
		panic(err)
	}
	err = s.Sync()
	if err != nil {
		panic(err)
	}
}

func getClientset() kubernetes.Interface {
	return utilsk8s.GetClientset("Ubiquity flex sidecar")
}

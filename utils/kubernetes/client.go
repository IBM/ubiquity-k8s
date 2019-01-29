package kubernetes

import (
	"fmt"
	"runtime"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetClientset(baseName string) kubernetes.Interface {

	var config *rest.Config

	//config, err := rest.InClusterConfig()
	config, err := Config("", "", baseName)
	if err != nil {
		panic(fmt.Sprintf("Failed to create k8s InClusterConfig: %v", err))
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create client: %v", err))
	}
	return clientset
}

// config returns a *rest.Config, using either the kubeconfig (if specified)
// or an in-cluster configuration.
// baseName is used to build the user agent to tell the API Server who is calling it.
// set to "ubiquity" if we, the ubiquity hook executor, is calling the APIs.
// Note that we only need the in-cluster way in production env, others are for test perpose only.
func Config(kubeconfig, kubecontext, baseName string) (*rest.Config, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.ExplicitPath = kubeconfig
	configOverrides := &clientcmd.ConfigOverrides{CurrentContext: kubecontext}
	newConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, configOverrides)
	clientConfig, err := newConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	clientConfig.UserAgent = buildUserAgent(
		baseName,
		runtime.GOOS,
		runtime.GOARCH,
	)

	return clientConfig, nil
}

func buildUserAgent(command, os, arch string) string {
	return fmt.Sprintf(
		"%s (%s/%s)", command, os, arch)
}

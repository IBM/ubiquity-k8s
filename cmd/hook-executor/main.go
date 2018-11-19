package main

import (
	"fmt"
	"os"
	"runtime"

	_ "github.com/IBM/ubiquity-k8s/cmd/hook-executor/logger"
	"github.com/IBM/ubiquity-k8s/helm_chart/hookexecutor"

	flags "github.com/jessevdk/go-flags"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

//PostInstallCommand
type PostInstallCommand struct {
	PostInstall func() `short:"i" long:"postinstall" description:"post install"`
}

func (c *PostInstallCommand) Execute(args []string) error {
	client := getClientset()
	return hookexecutor.PostInstallExecutor(client).Execute()
}

//PreDeleteCommand
type PreDeleteCommand struct {
	PreDelete func() `short:"d" long:"predelete" description:"pre delete"`
}

func (c *PreDeleteCommand) Execute(args []string) error {
	client := getClientset()
	return hookexecutor.PreDeleteExecutor(client).Execute()
}

//SanityCommand
type SanityCommand struct {
	Sanity func() `short:"s" long:"sanity" description:"sanity"`
}

func (c *SanityCommand) Execute(args []string) error {
	client := getClientset()
	return hookexecutor.SanityExecutor(client).Execute()
}

type Options struct{}

func main() {
	var postInstallCommand PostInstallCommand
	var preDeleteCommand PreDeleteCommand
	var sanityCommand SanityCommand

	var options Options
	var parser = flags.NewParser(&options, flags.Default)

	parser.AddCommand("postinstall",
		"post install",
		"post install",
		&postInstallCommand)

	parser.AddCommand("predelete",
		"pre delete",
		"pre delete",
		&preDeleteCommand)

	parser.AddCommand("sanity",
		"sanity",
		"sanity",
		&sanityCommand)

	_, err := parser.Parse()
	if err != nil {
		panic(err)
		os.Exit(1)
	}
}

func getClientset() kubernetes.Interface {

	var config *rest.Config

	//config, err := rest.InClusterConfig()
	config, err := Config("", "", "ubiquity")
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

package main

import (
	"os"

	flags "github.com/jessevdk/go-flags"
	"k8s.io/client-go/kubernetes"

	"github.com/IBM/ubiquity-k8s/helm_chart/hookexecutor"
	utilsk8s "github.com/IBM/ubiquity-k8s/utils/kubernetes"
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
	return utilsk8s.GetClientset("Ubiquity helm utils")
}

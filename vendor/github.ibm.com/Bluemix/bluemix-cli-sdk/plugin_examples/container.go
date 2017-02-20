package main

import (
	"fmt"

	"github.ibm.com/Bluemix/bluemix-cli-sdk/plugin"
)

type BluemixContainerPluginDemo struct{}

func main() {
	plugin.Start(new(BluemixContainerPluginDemo))
}

func (docker *BluemixContainerPluginDemo) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "IBM-container",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 1,
			Build: 1,
		},
		Namespaces: []plugin.Namespace{
			{
				Name:        "ic",
				Description: "Manage Bluemix containers",
			},
		},
		Commands: []plugin.Command{
			{
				Namespace:   "ic",
				Name:        "init",
				Description: "Creates a Dockerfile matching the language and framework for the app.",
				Usage:       "bluemix ic init",
			},
			{
				Namespace:   "ic",
				Name:        "start",
				Description: "Start local Docker app container running Procfile-defined process.",
				Usage:       "bluemix ic start",
			},
			{
				Namespace:   "ic",
				Name:        "clean",
				Description: "Clean up and remove local Docker images",
				Usage:       "bluemix ic clean",
			},
		},
	}
}

func (demo *BluemixContainerPluginDemo) Run(context plugin.PluginContext, args []string) {
	switch args[0] {
	case "init":
		demo.Init()
	case "start":
		demo.Start()
	case "clean":
		demo.Clean()
	}
}

func (demo *BluemixContainerPluginDemo) Init() {
	fmt.Println("ic init ...")
}

func (demo *BluemixContainerPluginDemo) Start() {
	fmt.Println("ic start ...")
}

func (demo *BluemixContainerPluginDemo) Clean() {
	fmt.Println("ic clean ...")
}

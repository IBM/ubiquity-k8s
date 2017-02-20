package main

import (
	"fmt"
	"os"

	"github.ibm.com/Bluemix/bluemix-cli-sdk/bluemix/terminal"
	"github.ibm.com/Bluemix/bluemix-cli-sdk/plugin"
)

type BluemixCFPluginDemo struct {
	ui terminal.UI
}

func main() {
	plugin.Start(NewBluemixCFPluginDemo())
}

func NewBluemixCFPluginDemo() *BluemixCFPluginDemo {
	return &BluemixCFPluginDemo{
		ui: terminal.NewStdUI(),
	}
}

func (demo *BluemixCFPluginDemo) Run(context plugin.PluginContext, args []string) {
	switch args[0] {
	case "push":
		if len(args) != 2 {
			demo.ui.Failed("Invalid usage.")
			os.Exit(2)
		}

		if !context.IsLoggedIn() {
			demo.ui.Failed("Not logged in. Use 'bx login' to log in.")
			os.Exit(1)
		}

		fmt.Printf("Push application '%s' to Bluemix...\n", args[1])
	}
}

func (demo *BluemixCFPluginDemo) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "bluemix-cf-plugin",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 1,
			Build: 1,
		},
		Commands: []plugin.Command{
			{
				Namespace:   "cf",
				Name:        "push",
				Description: "Deploy an application to Bluemix. To obtain more information, use --help.",
				Usage:       "bluemix cf push APP_NAME",
			},
		},
	}
}

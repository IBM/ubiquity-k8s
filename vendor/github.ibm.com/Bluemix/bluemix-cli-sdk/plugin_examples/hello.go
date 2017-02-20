package main

import (
	"fmt"
	"github.ibm.com/Bluemix/bluemix-cli-sdk/plugin"
)

type HelloWorldPlugin struct{}

func main() {
	plugin.Start(new(HelloWorldPlugin))
}

func (pluginDemo *HelloWorldPlugin) Run(context plugin.PluginContext, args []string) {
	fmt.Println("Hi, this is my first plugin for Bluemix")
}

func (pluginDemo *HelloWorldPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "Hello",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 0,
			Build: 1,
		},
		Commands: []plugin.Command{
			{
				Name:        "hello",
				Alias:       "hi",
				Description: "say hello to Bluemix.",
				Usage:       "bluemix hello",
			},
		},
	}
}

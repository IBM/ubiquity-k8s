package plugin

import (
	"encoding/json"
	"fmt"
	"os"

	"github.ibm.com/Bluemix/bluemix-cli-sdk/bluemix/configuration/core_config"
)

func Start(plugin Plugin) {
	if isMetadataRequest() {
		json, err := json.Marshal(plugin.GetMetadata())
		if err != nil {
			panic(err)
		}
		os.Stdout.Write(json)
	} else {
		Run(plugin, os.Args[1:])
	}
}

func Run(plugin Plugin, args []string) {
	config := core_config.NewCoreConfig(func(err error) {
		panic(fmt.Sprintf("Configuration error: %v", err))
	})
	plugin.Run(NewPluginContext(plugin.GetMetadata().Name, config), args)
}

func isMetadataRequest() bool {
	args := os.Args
	return len(args) == 2 && args[1] == "SendMetadata"
}

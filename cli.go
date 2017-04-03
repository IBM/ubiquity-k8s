package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.com/BurntSushi/toml"
	"github.com/IBM/ubiquity-k8s/controller"
	flags "github.com/jessevdk/go-flags"

	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
)

var configFile = flag.String(
	"configFile",
	"/tmp/ubiquity-client.conf",
	"config file with ubiquity client configuration params",
)

type InitCommand struct {
	Init func() `short:"i" long:"init" description:"Initialize the plugin"`
}

func (i *InitCommand) Execute(args []string) error {
	controller, err := createController(*configFile)
	if err != nil {
		response := resources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed tocreate controller %#v", err),
			Device:  "",
		}
		return utils.PrintResponse(response)
	}
	response := controller.Init()
	return utils.PrintResponse(response)
}

type AttachCommand struct {
	Attach func() `short:"a" long:"attach" description:"Attach a volume"`
}

func (a *AttachCommand) Execute(args []string) error {
	attachRequest := make(map[string]string)
	err := json.Unmarshal([]byte(args[0]), &attachRequest)
	if err != nil {
		response := resources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to attach volume %#v", err),
			Device:  "",
		}
		return utils.PrintResponse(response)
	}
	controller, err := createController(*configFile)

	if err != nil {
		panic(fmt.Sprintf("backend %s not found", configFile))
	}
	attachResponse := controller.Attach(attachRequest)
	return utils.PrintResponse(attachResponse)
}

type DetachCommand struct {
	Detach func() `short:"d" long:"detach" description:"Detach a volume"`
}

func (d *DetachCommand) Execute(args []string) error {
	mountDevice := args[0]
	controller, err := createController(*configFile)

	if err != nil {
		panic("backend not found")
	}

	detachRequest := resources.FlexVolumeDetachRequest{Name: mountDevice}
	detachResponse := controller.Detach(detachRequest)
	return utils.PrintResponse(detachResponse)
}

type MountCommand struct {
	Mount func() `short:"m" long:"mount" description:"Mount a volume Id to a path"`
}

func (m *MountCommand) Execute(args []string) error {
	targetMountDir := args[0]
	mountDevice := args[1]
	var mountOpts map[string]interface{}

	err := json.Unmarshal([]byte(args[2]), &mountOpts)
	if err != nil {
		mountResponse := resources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to mount device %s to %s due to: %#v", mountDevice, targetMountDir, err),
			Device:  mountDevice,
		}
		return utils.PrintResponse(mountResponse)
	}

	mountRequest := resources.FlexVolumeMountRequest{
		MountPath:   targetMountDir,
		MountDevice: mountDevice,
		Opts:        mountOpts,
	}

	controller, err := createController(*configFile)

	if err != nil {
		panic("backend not found")
	}
	mountResponse := controller.Mount(mountRequest)
	return utils.PrintResponse(mountResponse)
}

type UnmountCommand struct {
	UnMount func() `short:"u" long:"unmount" description:"UnMount a volume Id to a path"`
}

func (u *UnmountCommand) Execute(args []string) error {
	mountDir := args[0]
	controller, err := createController(*configFile)

	if err != nil {
		panic("backend not found")
	}

	unmountRequest := resources.FlexVolumeUnmountRequest{
		MountPath: mountDir,
	}
	unmountResponse := controller.Unmount(unmountRequest)
	return utils.PrintResponse(unmountResponse)
}

type Options struct{}

func main() {
	var mountCommand MountCommand
	var unmountCommand UnmountCommand
	var attachCommand AttachCommand
	var detachCommand DetachCommand
	var initCommand InitCommand
	var options Options
	var parser = flags.NewParser(&options, flags.Default)

	parser.AddCommand("init",
		"Init the plugin",
		"The info command print the driver name and version.",
		&initCommand)
	parser.AddCommand("mount",
		"Mount Volume",
		"Mount a volume Id to a path - returning the path.",
		&mountCommand)
	parser.AddCommand("unmount",
		"Unmount Volume",
		"UnMount given a mount dir",
		&unmountCommand)
	parser.AddCommand("attach",
		"Attach Volume",
		"Attach Volume",
		&attachCommand)
	parser.AddCommand("detach",
		"Detach Volume",
		"Detach a Volume",
		&detachCommand)
	_, err := parser.Parse()
	if err != nil {
		panic(err)
		os.Exit(1)
	}
}

func createController(configFile string) (*controller.Controller, error) {
	var config resources.UbiquityPluginConfig
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		fmt.Printf("error decoding config file", err)
		return nil, err

	}

	logger, _ := setupLogger(config.LogPath)
	//defer closeLogs(logFile)

	storageApiURL := fmt.Sprintf("http://%s:%d/ubiquity_storage", config.UbiquityServer.Address, config.UbiquityServer.Port)
	controller, err := controller.NewController(logger, storageApiURL, config)
	return controller, err
}

func setupLogger(logPath string) (*log.Logger, *os.File) {
	logFile, err := os.OpenFile(path.Join(logPath, "ubiquity-flexvolume.log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)
	if err != nil {
		fmt.Printf("Failed to setup logger: %s\n", err.Error())
		return nil, nil
	}
	log.SetOutput(logFile)
	logger := log.New(io.MultiWriter(logFile), "ubiquity-flexvolume: ", log.Lshortfile|log.LstdFlags)
	return logger, logFile
}

func closeLogs(logFile *os.File) {
	logFile.Sync()
	logFile.Close()
}

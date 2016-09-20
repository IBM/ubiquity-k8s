package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	flags "github.com/jessevdk/go-flags"

	common "github.ibm.com/almaden-containers/spectrum-common.git/core"
	"github.ibm.com/almaden-containers/spectrum-common.git/models"
	"github.ibm.com/almaden-containers/spectrum-flexvolume-cli.git/core"
)

var filesystemName = flag.String(
	"filesystem",
	"gpfs1",
	"gpfs filesystem name for this plugin",
)
var defaultMountPath = flag.String(
	"mountpath",
	"/gpfs/gpfs1",
	"gpfs mount path",
)
var logPath = flag.String(
	"logPath",
	"/tmp/spectrum-flex/log",
	"log path",
)

type InitCommand struct {
	Init func() `short:"i" long:"init" description:"Initialize the plugin"`
}

func (i *InitCommand) Execute(args []string) error {
	response := models.FlexVolumeResponse{
		Status:  "Success",
		Message: "FlexVolume Init success",
		Device:  "",
	}
	return response.PrintResponse()
}

type AttachCommand struct {
	Attach func() `short:"a" long:"attach" description:"Attach a volume"`
}

func (a *AttachCommand) Execute(args []string) error {
	logger, logFile := setupLogger(*logPath)
	defer closeLogs(logFile)
	attachRequest := models.FlexVolumeAttachRequest{}
	err := json.Unmarshal([]byte(args[0]), &attachRequest)
	if err != nil {
		response := models.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to attach volume %#v", err),
			Device:  "",
		}
		return response.PrintResponse()
	}
	dbClient := common.NewDatabaseClient(logger, *filesystemName, *defaultMountPath)
	defer dbClient.Close()

	err = dbClient.Init()
	if err != nil {
		response := models.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to init databasee %#v", err),
			Device:  "",
		}
		return response.PrintResponse()
	}
	err = dbClient.CreateVolumeTable()
	if err != nil {
		response := models.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to create volume table %#v", err),
			Device:  "",
		}
		return response.PrintResponse()
	}

	controller := core.NewController(logger, *filesystemName, *defaultMountPath, dbClient)
	attachResponse := controller.Attach(&attachRequest)
	return attachResponse.PrintResponse()
}

type DetachCommand struct {
	Detach func() `short:"d" long:"detach" description:"Detach a volume"`
}

func (d *DetachCommand) Execute(args []string) error {
	mountDevice := args[0]

	logger, logFile := setupLogger(*logPath)
	defer closeLogs(logFile)
	dbClient := common.NewDatabaseClient(logger, *filesystemName, *defaultMountPath)
	defer dbClient.Close()

	err := dbClient.Init()
	if err != nil {
		response := models.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to init databasee %#v", err),
			Device:  "",
		}
		return response.PrintResponse()
	}

	controller := core.NewController(logger, *filesystemName, *defaultMountPath, dbClient)
	removeRequest := models.GenericRequest{Name: mountDevice}
	removeResponse := controller.Detach(&removeRequest)
	return removeResponse.PrintResponse()
}

type MountCommand struct {
	Mount func() `short:"m" long:"mount" description:"Mount a volume Id to a path"`
}

func (m *MountCommand) Execute(args []string) error {
	// type FlexVolumeMountRequest struct {
	// 	MountPath   string                 `json:"mountPath"`
	// 	MountDevice string                 `json:"name"`
	// 	Opts        map[string]interface{} `json:"opts"`
	// }
	targetMountDir := args[0]
	mountDevice := args[1]
	var mountOpts map[string]interface{}
	logger, logFile := setupLogger(*logPath)
	defer closeLogs(logFile)

	logger.Printf("mount-args %#v\n", args[2])
	err := json.Unmarshal([]byte(args[2]), &mountOpts)
	if err != nil {
		mountResponse := models.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to mount device %s to %s due to: %#v", mountDevice, targetMountDir, err),
			Device:  mountDevice,
		}
		return mountResponse.PrintResponse()
	}

	dbClient := common.NewDatabaseClient(logger, *filesystemName, *defaultMountPath)
	defer dbClient.Close()

	err = dbClient.Init()
	if err != nil {
		response := models.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to init databasee %#v", err),
			Device:  "",
		}
		return response.PrintResponse()
	}

	controller := core.NewController(logger, *filesystemName, *defaultMountPath, dbClient)

	mountRequest := models.FlexVolumeMountRequest{
		MountPath:   targetMountDir,
		MountDevice: mountDevice,
		Opts:        mountOpts,
	}
	mountResponse := controller.Mount(&mountRequest)
	return mountResponse.PrintResponse()
}

type UnmountCommand struct {
	UnMount func() `short:"u" long:"unmount" description:"UnMount a volume Id to a path"`
}

func (u *UnmountCommand) Execute(args []string) error {
	mountDir := args[0]

	logger, logFile := setupLogger(*logPath)
	defer closeLogs(logFile)
	//in this case the filesystem name will not be used
	// the spectrum client will get the right mapping from the mountDir
	logger.Printf("CLI: unmount arg0 (mountDir)", mountDir)
	dbClient := common.NewDatabaseClient(logger, *filesystemName, *defaultMountPath)
	defer dbClient.Close()

	err := dbClient.Init()
	if err != nil {
		response := models.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to init databasee %#v", err),
			Device:  "",
		}
		return response.PrintResponse()
	}

	controller := core.NewController(logger, *filesystemName, *defaultMountPath, dbClient)

	unmountRequest := models.GenericRequest{
		Name: mountDir,
	}
	unmountResponse := controller.Unmount(&unmountRequest)
	return unmountResponse.PrintResponse()
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

func setupLogger(logPath string) (*log.Logger, *os.File) {
	logFile, err := os.OpenFile(path.Join(logPath, "spectrum-scale-cli.log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)
	if err != nil {
		fmt.Printf("Failed to setup logger: %s\n", err.Error())
		return nil, nil
	}
	log.SetOutput(logFile)
	// logger := log.New(io.MultiWriter(logFile, os.Stdout), "spectrum-cli: ", log.Lshortfile|log.LstdFlags)
	logger := log.New(io.MultiWriter(logFile), "spectrum-cli: ", log.Lshortfile|log.LstdFlags)
	return logger, logFile
}

func closeLogs(logFile *os.File) {
	logFile.Sync()
	logFile.Close()
}

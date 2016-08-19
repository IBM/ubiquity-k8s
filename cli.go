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

	"github.ibm.com/almaden-containers/spectrum-common.git/models"
	"github.ibm.com/almaden-containers/spectrum-container-plugin.git/core"
)

var filesystemName = flag.String(
	"filesystem",
	"gpfs1",
	"gpfs filesystem name for this plugin",
)
var defaultMountPath = flag.String(
	"mountpath",
	"/gpfs/fs1",
	"gpfs mount path",
)
var logPath = flag.String(
	"logPath",
	"/tmp",
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
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return err
	}
	fmt.Printf("%s", string(responseBytes[:]))
	return nil
}

type AttachCommand struct {
	Attach func() `short:"a" long:"attach" description:"Attach a volume"`
}

func (a *AttachCommand) Execute(args []string) models.FlexVolumeResponse {
	logger, logFile := setupLogger(*logPath)
	defer closeLogs(logFile)

	attachOpt := models.FlexVolumeAttachOptions{}

	err := json.Unmarshal([]byte(args[0]), &attachOpt)
	if err != nil {
		return models.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to attach volume %#v", err),
			Device:  "",
		}
	}
	controller := core.NewController(logger, attachOpt.VolumeId, attachOpt.Path)
	activateResponse := controller.Activate()

	if len(activateResponse.Implements) == 0 {
		return models.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to attach the volume %#v", attachOpt),
			Device:  attachOpt.VolumeId,
		}
	}

	createrequest := models.CreateRequest{Name: attachOpt.VolumeId,
		Opts: map[string]interface{}{"VolumeId": attachOpt.VolumeId,
			"fileset": attachOpt.FileSet},
	}
	createResponse := controller.Create(&createrequest)

	if createResponse.Err != "" {
		return models.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to attach the volume %#v with error %s", attachOpt, createResponse.Err),
			Device:  attachOpt.VolumeId,
		}
	} else {
		return models.FlexVolumeResponse{
			Status:  "Success",
			Message: fmt.Sprintf("Device attached successfully %#v", attachOpt),
			Device:  attachOpt.VolumeId,
		}
	}
}

type DetachCommand struct {
	Detach func() `short:"d" long:"detach" description:"Detach a volume"`
}

func (d *DetachCommand) Execute(args []string) models.FlexVolumeResponse {
	mountDevice := args[0]
	logger, logFile := setupLogger(*logPath)
	defer closeLogs(logFile)

	controller := core.NewController(logger, "filesysten", "mountpath")
	removeRequest := models.GenericRequest{Name: mountDevice}
	removeResponse := controller.Remove(&removeRequest)
	if removeResponse.Err != "" {
		return models.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to detach device %s due to : %#v", mountDevice, removeResponse.Err),
			Device:  mountDevice,
		}

	} else {
		return models.FlexVolumeResponse{
			Status:  "Success",
			Message: fmt.Sprintf("Detached volume %s successfully", mountDevice),
			Device:  mountDevice,
		}
	}
}

type MountCommand struct {
	Mount func() `short:"m" long:"mount" description:"Mount a volume Id to a path"`
}

func (m *MountCommand) Execute(args []string) models.FlexVolumeResponse {
	targetMountDir := args[0]
	mountDevice := args[1]
	mountOpt := models.FlexVolumeMountOptions{}

	err := json.Unmarshal([]byte(args[2]), &mountOpt)
	if err != nil {
		return models.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to mount device %s to %s due to: %#v", mountDevice, targetMountDir, err),
			Device:  mountDevice,
		}
	}

	logger, logFile := setupLogger(*logPath)
	defer closeLogs(logFile)
	controller := core.NewController(logger, mountOpt.MountDevice, mountOpt.MountPath)

	mountRequest := models.GenericRequest{
		Name: mountDevice,
	}
	mountResponse := controller.Mount(&mountRequest)
	if mountResponse.Err != "" {
		return models.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to mount device %s to %s due to %s", mountDevice, targetMountDir, mountResponse.Err),
			Device:  mountDevice,
		}

	}
	return models.FlexVolumeResponse{
		Status:  "Success",
		Message: fmt.Sprintf("Device %s mounted to %s successfully", mountDevice, targetMountDir),
		Device:  mountResponse.Mountpoint,
	}
}

type UnmountCommand struct {
	UnMount func() `short:"u" long:"unmount" description:"UnMount a volume Id to a path"`
}

func (u *UnmountCommand) Execute(args []string) models.FlexVolumeResponse {
	mountDir := args[0]

	logger, logFile := setupLogger(*logPath)
	defer closeLogs(logFile)
	controller := core.NewController(logger, "filesystem", mountDir)

	unmountRequest := models.GenericRequest{
		Name: mountDir,
	}
	unmountResponse := controller.Unmount(&unmountRequest)
	if unmountResponse.Err != "" {
		return models.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to unmount %s due to %s", mountDir, unmountResponse.Err),
			Device:  "",
		}

	} else {
		return models.FlexVolumeResponse{
			Status:  "Success",
			Message: fmt.Sprintf("%s unmounted successfully", mountDir),
			Device:  "",
		}
	}
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
	logger := log.New(io.MultiWriter(logFile, os.Stdout), "spectrum-cli: ", log.Lshortfile|log.LstdFlags)
	return logger, logFile
}

func closeLogs(logFile *os.File) {
	logFile.Sync()
	logFile.Close()
}

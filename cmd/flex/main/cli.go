/**
 * Copyright 2016, 2017 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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

	k8sresources "github.com/IBM/ubiquity-k8s/resources"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils/logs"
)

var configFile = flag.String(
	"configFile",
	"/tmp/ubiquity-client.conf",
	"config file with ubiquity client configuration params",
)

//<driver executable> init
//<driver executable> getvolumename <json options>
//<driver executable> attach <json options> <node name>
//<driver executable> detach <mount device> <node name>
//<driver executable> waitforattach <mount device> <json options>
//<driver executable> isattached <json options> <node name>
//<driver executable> mountdevice <mount dir> <mount device> <json options>
//<driver executable> unmountdevice <mount device>
//<driver executable> mount <mount dir> <json options>
//<driver executable> unmount <mount dir>
//
//
//{
//"status": "<Success/Failure/Not supported>",
//"message": "<Reason for success/failure>",
//"device": "Path to the device attached. valid only for attach & waitforattach call-outs”
//"volumeName": "Cluster wide unique name of the volume”
//"attached": True/False}

type InitCommand struct {
	Init func() `short:"i" long:"init" description:"Initialize the plugin"`
}

func (i *InitCommand) Execute(args []string) error {
	config, err := readConfig(*configFile)
	if err != nil {
		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to read config in Init %#v", err),
		}
		return printResponse(response)
	}
	defer logs.InitFileLogger(logs.DEBUG, path.Join(config.LogPath, "ubiquity-flexvolume.log"))()
	controller, err := createController(config)
	if err != nil {
		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed tocreate controller %#v", err),
		}
		return printResponse(response)
	}
	response := controller.Init(config)
	return printResponse(response)
}

type GetVolumeNameCommand struct {
	GetVolumeName func() `short:"g" long:"getvolumename" description:"Get Volume Name"`
}

func (g *GetVolumeNameCommand) Execute(args []string) error {
	getVolumeNameRequestOpts := make(map[string]string)
	err := json.Unmarshal([]byte(args[0]), &getVolumeNameRequestOpts)
	if err != nil {
		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to read args in get volumeName %#v", err),
		}
		return printResponse(response)
	}
	config, err := readConfig(*configFile)
	if err != nil {
		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to read config in get volumeName %#v", err),
		}
		return printResponse(response)
	}
	defer logs.InitFileLogger(logs.DEBUG, path.Join(config.LogPath, "ubiquity-flexvolume.log"))()
	controller, err := createController(config)

	if err != nil {
		panic(fmt.Sprintf("backend %s not found", config))
	}
	getVolumeNameRequest := k8sresources.FlexVolumeGetVolumeNameRequest{Opts: getVolumeNameRequestOpts}
	getVolumeNameResponse := controller.GetVolumeName(getVolumeNameRequest)
	return printResponse(getVolumeNameResponse)
}

type AttachCommand struct {
	Attach func() `short:"a" long:"attach" description:"Attach a volume"`
}

func (a *AttachCommand) Execute(args []string) error {
	attachRequest := make(map[string]string)
	err := json.Unmarshal([]byte(args[0]), &attachRequest)
	if err != nil {
		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to attach volume %#v", err),
		}
		return printResponse(response)
	}
	config, err := readConfig(*configFile)
	if err != nil {
		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to read config in attach volume %#v", err),
		}
		return printResponse(response)
	}
	defer logs.InitFileLogger(logs.DEBUG, path.Join(config.LogPath, "ubiquity-flexvolume.log"))()
	controller, err := createController(config)

	if err != nil {
		panic(fmt.Sprintf("backend %s not found", config))
	}
	attachResponse := controller.Attach(attachRequest)
	return printResponse(attachResponse)
}

//WaitForAttach the volume to be attached on the node
type WaitForAttachCommand struct {
	WaitForAttach func() `short:"w" long:"waitfa" description:"Wait For Attach"`
}

func (wfa *WaitForAttachCommand) Execute(args []string) error {
	config, err := readConfig(*configFile)
	if err != nil {
		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to read config in waitForAttach volume %#v", err),
		}
		return printResponse(response)
	}
	defer logs.InitFileLogger(logs.DEBUG, path.Join(config.LogPath, "ubiquity-flexvolume.log"))()
	controller, err := createController(config)
	opts := make(map[string]string)
	err = json.Unmarshal([]byte(args[1]), &opts)
	if err != nil {
		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to marshall args in waitForAttach %#v", err),
		}
		return printResponse(response)
	}
	waitForAttachRequest := k8sresources.FlexVolumeWaitForAttachRequest{Name: args[0], Opts: opts}
	response := controller.WaitForAttach(waitForAttachRequest)
	return printResponse(response)
}

//	//Checks if the volume is attached to the node
//	IsAttached(request map[string]string, nodeName string) FlexVolumeResponse
//
type IsAttachedCommand struct {
	IsAttacheded func() `short:"z" long:"detach" description:"Detach a volume"`
}

func (d *IsAttachedCommand) Execute(args []string) error {
	config, err := readConfig(*configFile)
	if err != nil {
		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to read config in isAttached volume %#v", err),
		}
		return printResponse(response)
	}
	defer logs.InitFileLogger(logs.DEBUG, path.Join(config.LogPath, "ubiquity-flexvolume.log"))()
	controller, err := createController(config)
	opts := make(map[string]string)
	err = json.Unmarshal([]byte(args[0]), &opts)
	if err != nil {
		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to marshall args in isAttached %#v", err),
		}
		return printResponse(response)
	}
	isAttachedRequest := k8sresources.FlexVolumeIsAttachedRequest{Opts: opts, Host: args[1]}
	response := controller.IsAttached(isAttachedRequest)
	return printResponse(response)
}

type DetachCommand struct {
	Detach func() `short:"d" long:"detach" description:"Detach a volume"`
}

func (d *DetachCommand) Execute(args []string) error {
	mountDevice := args[0]

	config, err := readConfig(*configFile)
	if err != nil {
		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to read config in detach %#v", err),
		}
		return printResponse(response)
	}
	defer logs.InitFileLogger(logs.DEBUG, path.Join(config.LogPath, "ubiquity-flexvolume.log"))()
	controller, err := createController(config)

	if err != nil {
		panic("backend not found")
	}

	detachRequest := k8sresources.FlexVolumeDetachRequest{Name: mountDevice}
	detachResponse := controller.Detach(detachRequest)
	return printResponse(detachResponse)
}

type MountDeviceCommand struct {
	MountDevice func() `short:"x" long:"mountdevice" description:"Mounts a device"`
}

//MountDevice Mounts the device to a global path which individual pods can then bind mount
////<driver executable> mountdevice <mount dir> <mount device> <json options>
func (d *MountDeviceCommand) Execute(args []string) error {
	config, err := readConfig(*configFile)
	if err != nil {
		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to read config in MountDevice  %#v", err),
		}
		return printResponse(response)
	}
	defer logs.InitFileLogger(logs.DEBUG, path.Join(config.LogPath, "ubiquity-flexvolume.log"))()
	controller, err := createController(config)
	opts := make(map[string]string)
	err = json.Unmarshal([]byte(args[2]), &opts)
	if err != nil {
		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to marshall args in MountDevice %#v", err),
		}
		return printResponse(response)
	}
	mountDeviceRequest := k8sresources.FlexVolumeMountDeviceRequest{Path: args[0], Name: args[1], Opts: opts}
	response := controller.MountDevice(mountDeviceRequest)
	return printResponse(response)
}

//UnmountDevice	Unmounts the global mount for the device. This is called once all bind mounts have been unmounted
//<driver executable> unmountdevice <mount device>
type UnmountDeviceCommand struct {
	UnmountDevice func() `short:"ud" long:"umountdevice" description:"Unmounts a device"`
}

func (d *UnmountDeviceCommand) Execute(args []string) error {
	config, err := readConfig(*configFile)
	if err != nil {
		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to read config in UnmountDevice  %#v", err),
		}
		return printResponse(response)
	}
	defer logs.InitFileLogger(logs.DEBUG, path.Join(config.LogPath, "ubiquity-flexvolume.log"))()
	controller, err := createController(config)

	unmountDeviceRequest := k8sresources.FlexVolumeUnmountDeviceRequest{Name: args[0]}
	response := controller.UnmountDevice(unmountDeviceRequest)
	return printResponse(response)
}

type MountCommand struct {
	Mount func() `short:"m" long:"mount" description:"Mount a volume Id to a path"`
}

func (m *MountCommand) Execute(args []string) error {
	targetMountDir := args[0]

	var mountOpts map[string]interface{}

	err := json.Unmarshal([]byte(args[1]), &mountOpts)
	if err != nil {
		mountResponse := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to mount device %s to %s due to: %#v", targetMountDir, err),
		}
		return printResponse(mountResponse)
	}
	volumeName, ok := mountOpts["volumeName"]
	if !ok {
		mountResponse := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to get volumeName in opts: %#v", mountOpts),
		}
		return printResponse(mountResponse)

	}
	mountRequest := k8sresources.FlexVolumeMountRequest{
		MountPath:   targetMountDir,
		MountDevice: volumeName.(string),
		Opts:        mountOpts,
	}

	config, err := readConfig(*configFile)
	if err != nil {
		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to read config in mount %#v", err),
			Device:  "",
		}
		return printResponse(response)
	}
	defer logs.InitFileLogger(logs.DEBUG, path.Join(config.LogPath, "ubiquity-flexvolume.log"))()
	controller, err := createController(config)

	if err != nil {
		panic("backend not found")
	}
	mountResponse := controller.Mount(mountRequest)
	return printResponse(mountResponse)
}

type UnmountCommand struct {
	UnMount func() `short:"u" long:"unmount" description:"UnMount a volume Id to a path"`
}

func (u *UnmountCommand) Execute(args []string) error {
	mountDir := args[0]
	config, err := readConfig(*configFile)
	if err != nil {
		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to read config in Unmount %#v", err),
			Device:  "",
		}
		return printResponse(response)
	}
	defer logs.InitFileLogger(logs.DEBUG, path.Join(config.LogPath, "ubiquity-flexvolume.log"))()
	controller, err := createController(config)

	if err != nil {
		panic("backend not found")
	}

	unmountRequest := k8sresources.FlexVolumeUnmountRequest{
		MountPath: mountDir,
	}
	unmountResponse := controller.Unmount(unmountRequest)
	return printResponse(unmountResponse)
}

type Options struct{}

func main() {
	var mountCommand MountCommand
	var unmountCommand UnmountCommand
	var attachCommand AttachCommand
	var detachCommand DetachCommand
	var initCommand InitCommand
	var getVolumeNameCommand GetVolumeNameCommand
	var isAttachedCommand IsAttachedCommand
	var waitForAttachCommand WaitForAttachCommand
	var mountDeviceCommand MountDeviceCommand
	var unmountDeviceCommand UnmountDeviceCommand

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
	parser.AddCommand("isattached",
		"Is Volume Attached",
		"Is Volume Attached",
		&isAttachedCommand)
	parser.AddCommand("waitforattach",
		"Wait for volume to get attached",
		"Wait for volume to get attached",
		&waitForAttachCommand)
	parser.AddCommand("getvolumename",
		"Get Volume Name",
		"Get Volume Name",
		&getVolumeNameCommand)
	parser.AddCommand("mountdevice",
		"Mount Device",
		"Mount Device",
		&mountDeviceCommand)
	parser.AddCommand("unmountdevice",
		"Unmount Device",
		"Unmount Device",
		&unmountDeviceCommand)

	_, err := parser.Parse()
	if err != nil {
		panic(err)
		os.Exit(1)
	}
}

func createController(config resources.UbiquityPluginConfig) (*controller.Controller, error) {

	logger, _ := setupLogger(config.LogPath)
	//defer closeLogs(logFile)

	storageApiURL := fmt.Sprintf("http://%s:%d/ubiquity_storage", config.UbiquityServer.Address, config.UbiquityServer.Port)
	controller, err := controller.NewController(logger, storageApiURL, config)
	return controller, err
}

func readConfig(configFile string) (resources.UbiquityPluginConfig, error) {
	var config resources.UbiquityPluginConfig
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		fmt.Printf("error decoding config file", err)
		return resources.UbiquityPluginConfig{}, err

	}
	return config, nil
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

func printResponse(f k8sresources.FlexVolumeResponse) error {
	responseBytes, err := json.Marshal(f)
	if err != nil {
		return err
	}
	fmt.Printf("%s", string(responseBytes[:]))
	return nil
}

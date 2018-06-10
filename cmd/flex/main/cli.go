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
	"os"
	"path"

	"github.com/BurntSushi/toml"
	"github.com/IBM/ubiquity-k8s/controller"
	flags "github.com/jessevdk/go-flags"

	"strconv"

	k8sresources "github.com/IBM/ubiquity-k8s/resources"
	k8sutils "github.com/IBM/ubiquity-k8s/utils"
	"github.com/IBM/ubiquity/remote"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
	"github.com/IBM/ubiquity/utils/logs"
)

var configFile = flag.String(
	"configFile",
	k8sresources.FlexConfPath,
	"Flex Volume configuration file",
)

var logger_params = logs.LoggerParams{ShowGoid: false, ShowPid : true}

// All the method should printout as response:
//{
//"status": "<Success/Failure/Not supported>",
//"message": "<Reason for success/failure>",
//"device": "Path to the device attached. valid only for attach & waitforattach call-outs”
//"volumeName": "Cluster wide unique name of the volume”
//"attached": True/False}

//InitCommand initializes the plugin
//<driver executable> init (v>=1.5)
type InitCommand struct {
	Init func() `short:"i" long:"init" description:"Initialize the plugin"`
}

func (i *InitCommand) Execute(args []string) error {
	response := k8sresources.FlexVolumeResponse{
		Status:  "Success",
		Message: "Plugin init successfully",
	}

	return printResponse(response)
}

//GetVolumeNameCommand gets a unique volume name
//<driver executable> getvolumename <json options> (v>=1.6)
type GetVolumeNameCommand struct {
	GetVolumeName func() `short:"g" long:"getvolumename" description:"Get Volume Name"`
}

func (g *GetVolumeNameCommand) Execute(args []string) error {
	// This GetVolumeName action in FlexVolume CLI is not relevant, we can just return not supported without logging anything.
	response := k8sresources.FlexVolumeResponse{
		Status: "Not supported",
	}
	return printResponse(response)
}

//AttachCommand attaches a volume to a node
//<driver executable> attach <json options> <node name> (v=1.5 with json options, v >= 1.6 json options and node name)

type AttachCommand struct {
	Attach func() `short:"a" long:"attach" description:"Attach a volume"`
}

func (a *AttachCommand) Execute(args []string) error {
	var version string
	var hostname string

	requestContext := k8sutils.GetNewRequestContext()
	fmt.Sprintf("Starting attach command [%s]", k8sutils.GetContextRequestString(requestContext))

	if len(args) < 1 {

		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Not enough arguments to attach call out"),
		}
		return printResponse(response)
	}
	if len(args) == 1 {
		version = k8sresources.KubernetesVersion_1_5
	} else {
		hostname = args[1]
		version = k8sresources.KubernetesVersion_1_6OrLater
	}
	attachRequestOpts := make(map[string]string)
	err := json.Unmarshal([]byte(args[0]), &attachRequestOpts)
	if err != nil {
		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to unmarshall request in attach volume %#v", err),
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
	defer logs.InitFileLogger(logs.GetLogLevelFromString(config.LogLevel), path.Join(config.LogPath, k8sresources.UbiquityFlexLogFileName), logger_params)()
	controller, err := createController(config)

	if err != nil {
		panic(fmt.Sprintf("backend %s not found", config))
	}

	volumeName, ok := attachRequestOpts["volumeName"]
	if !ok {
		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("volumeName is mandatory for attach %#v", attachRequestOpts),
		}
		return printResponse(response)
	}

	attachRequest := k8sresources.FlexVolumeAttachRequest{Name: volumeName, Host: hostname, Opts: attachRequestOpts, Version: version, Context: requestContext}

	attachResponse := controller.Attach(attachRequest)
	fmt.Sprintf("Finishing attach command [%s]", k8sutils.GetContextRequestString(requestContext))
	return printResponse(attachResponse)
}

//WaitForAttach the volume to be attached on the node
//<driver executable> waitforattach <mount device> <json options> (v >= 1.6)
type WaitForAttachCommand struct {
	WaitForAttach func() `short:"w" long:"waitfa" description:"Wait For Attach"`
}

func (wfa *WaitForAttachCommand) Execute(args []string) error {
	requestContext := k8sutils.GetNewRequestContext()
	fmt.Sprintf("Starting WaitForAttachCommand command [%s]", k8sutils.GetContextRequestString(requestContext))

	if len(args) < 2 {

		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Not enough arguments to waitForAttach call out"),
		}
		return printResponse(response)
	}
	config, err := readConfig(*configFile)
	if err != nil {
		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to read config in waitForAttach volume %#v ", err),
		}
		return printResponse(response)
	}
	defer logs.InitFileLogger(logs.GetLogLevelFromString(config.LogLevel), path.Join(config.LogPath, k8sresources.UbiquityFlexLogFileName), logger_params)()
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
	waitForAttachRequest := k8sresources.FlexVolumeWaitForAttachRequest{Name: args[0], Opts: opts, Context: requestContext}
	response := controller.WaitForAttach(waitForAttachRequest)
	fmt.Sprintf("Finished WaitForAttachCommand command [%s]", k8sutils.GetContextRequestString(requestContext))
	return printResponse(response)
}

//IsAttachedCommand Checks if the volume is attached to the node
//<driver executable> isattached <json options> <node name> (v >= 1.6)
type IsAttachedCommand struct {
	IsAttacheded func() `short:"z" long:"detach" description:"Detach a volume"`
}

func (d *IsAttachedCommand) Execute(args []string) error {
	requestContext := k8sutils.GetNewRequestContext()
	fmt.Sprintf("Starting IsAttachedCommand command [%s]", k8sutils.GetContextRequestString(requestContext))
	if len(args) < 2 {

		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Not enough arguments to isAttached call out"),
		}
		return printResponse(response)
	}
	config, err := readConfig(*configFile)
	if err != nil {
		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to read config in isAttached volume %#v", err),
		}
		return printResponse(response)
	}
	defer logs.InitFileLogger(logs.GetLogLevelFromString(config.LogLevel), path.Join(config.LogPath, k8sresources.UbiquityFlexLogFileName), logger_params)()
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
	isAttachedRequest := k8sresources.FlexVolumeIsAttachedRequest{Opts: opts, Host: args[1], Context: requestContext}
	response := controller.IsAttached(isAttachedRequest)
	fmt.Sprintf("Finished IsAttachedCommand command [%s]", k8sutils.GetContextRequestString(requestContext))
	return printResponse(response)
}

//DetachCommand detaches a volume from a given node
//<driver executable> detach <mount device> <node name> (v=1.5 with mount device, v >= 1.6 mount device and node name)

type DetachCommand struct {
	Detach func() `short:"d" long:"detach" description:"Detach a volume"`
}

func (d *DetachCommand) Execute(args []string) error {
	requestContext := k8sutils.GetNewRequestContext()
	fmt.Sprintf("Starting DetachCommand command [%s]", k8sutils.GetContextRequestString(requestContext))
	var hostname string
	var version string
	if len(args) < 1 {

		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Not enough arguments to getVolumeName call out"),
		}
		return printResponse(response)
	}
	mountDevice := args[0]
	if len(args) == 1 {
		version = k8sresources.KubernetesVersion_1_5
	} else {
		hostname = args[1]
		version = k8sresources.KubernetesVersion_1_6OrLater
	}

	config, err := readConfig(*configFile)
	if err != nil {
		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to read config in detach %#v", err),
		}
		return printResponse(response)
	}
	defer logs.InitFileLogger(logs.GetLogLevelFromString(config.LogLevel), path.Join(config.LogPath, k8sresources.UbiquityFlexLogFileName), logger_params)()
	controller, err := createController(config)

	if err != nil {
		panic("backend not found")
	}

	detachRequest := k8sresources.FlexVolumeDetachRequest{Name: mountDevice, Host: hostname, Version: version, Context: requestContext}
	detachResponse := controller.Detach(detachRequest)
	fmt.Sprintf("Finished DetachCommand command [%s]", k8sutils.GetContextRequestString(requestContext))
	return printResponse(detachResponse)
}

//MountDevice Mounts the device to a global path which individual pods can then bind mount
//<driver executable> mountdevice <mount dir> <mount device> <json options> (v >= 1.6)
type MountDeviceCommand struct {
	MountDevice func() `short:"x" long:"mountdevice" description:"Mounts a device"`
}

func (d *MountDeviceCommand) Execute(args []string) error {
	requestContext := k8sutils.GetNewRequestContext()
	fmt.Sprintf("Starting MountDeviceCommand command [%s]", k8sutils.GetContextRequestString(requestContext))
	if len(args) < 3 {

		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Not enough arguments to mountDevice call out"),
		}
		return printResponse(response)
	}
	config, err := readConfig(*configFile)
	if err != nil {
		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to read config in MountDevice  %#v", err),
		}
		return printResponse(response)
	}
	defer logs.InitFileLogger(logs.GetLogLevelFromString(config.LogLevel), path.Join(config.LogPath, k8sresources.UbiquityFlexLogFileName), logger_params)()
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
	mountDeviceRequest := k8sresources.FlexVolumeMountDeviceRequest{Path: args[0], Name: args[1], Opts: opts, Context: requestContext}
	response := controller.MountDevice(mountDeviceRequest)
	fmt.Sprintf("Finished MountDeviceCommand command [%s]", k8sutils.GetContextRequestString(requestContext))
	return printResponse(response)
}

//UnmountDevice	Unmounts the global mount for the device. This is called once all bind mounts have been unmounted
//<driver executable> unmountdevice <mount device> (v >= 1.6)
type UnmountDeviceCommand struct {
	UnmountDevice func() `short:"y" long:"umountdevice" description:"Unmounts a device"`
}

func (d *UnmountDeviceCommand) Execute(args []string) error {
	requestContext := k8sutils.GetNewRequestContext()
	fmt.Sprintf("Starting UnmountDeviceCommand command [%s]", k8sutils.GetContextRequestString(requestContext))
	if len(args) < 1 {

		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Not enough arguments to unmountDevice call out"),
		}
		return printResponse(response)
	}
	config, err := readConfig(*configFile)
	if err != nil {
		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to read config in UnmountDevice  %#v", err),
		}
		return printResponse(response)
	}
	defer logs.InitFileLogger(logs.GetLogLevelFromString(config.LogLevel), path.Join(config.LogPath, k8sresources.UbiquityFlexLogFileName), logger_params)()
	controller, err := createController(config)

	unmountDeviceRequest := k8sresources.FlexVolumeUnmountDeviceRequest{Name: args[0], Context: requestContext}
	response := controller.UnmountDevice(unmountDeviceRequest)
	fmt.Sprintf("Finished UnmountDeviceCommand command [%s]", k8sutils.GetContextRequestString(requestContext))
	return printResponse(response)
}

//MountCommand mounts a given volume to a given mountpoint
//<driver executable> mount <mount dir> <mountDevice> <json options> (v>=1.5)
//<driver executable> mount <mount dir> <json options> (v>=1.6)
type MountCommand struct {
	Mount func() `short:"m" long:"mount" description:"Mount a volume Id to a path"`
}

func (m *MountCommand) Execute(args []string) error {
	var volumeName string
	var mountOpts map[string]string
	var mountOptsIndex int
	var ok bool
	var version string

	requestContext := k8sutils.GetNewRequestContext()
	fmt.Sprintf("Starting MountCommand command [%s]", k8sutils.GetContextRequestString(requestContext))

	//should error out when not enough args
	if len(args) < 2 {

		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Not enough arguments to mount call out"),
		}
		return printResponse(response)
	}
	targetMountDir := args[0]
	// kubernetes version 1.5
	if len(args) == 3 {
		volumeName = args[1]
		mountOptsIndex = 2
		version = k8sresources.KubernetesVersion_1_5

	} else /*kubernetes version 1.6*/ {
		mountOptsIndex = 1
		version = k8sresources.KubernetesVersion_1_6OrLater
	}

	err := json.Unmarshal([]byte(args[mountOptsIndex]), &mountOpts)

	if err != nil {
		mountResponse := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to mount device %s to %s due to: %#v", targetMountDir, err),
		}
		return printResponse(mountResponse)
	}

	if volumeName == "" {
		volumeName, ok = mountOpts["volumeName"]
		if !ok {
			mountResponse := k8sresources.FlexVolumeResponse{
				Status:  "Failure",
				Message: fmt.Sprintf("Failed to get volumeName in opts: %#v", mountOpts),
			}
			return printResponse(mountResponse)

		}
	}

	mountRequest := k8sresources.FlexVolumeMountRequest{
		MountPath:   targetMountDir,
		MountDevice: volumeName, // The PV name
		Opts:        mountOpts,
		Version:     version,
		Context:     requestContext,
	}

	config, err := readConfig(*configFile)
	if err != nil {
		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to read config in mount %#v", err),
		}
		return printResponse(response)
	}

	defer logs.InitFileLogger(logs.GetLogLevelFromString(config.LogLevel), path.Join(config.LogPath, k8sresources.UbiquityFlexLogFileName), logger_params)()
	controller, err := createController(config)

	if err != nil {
		panic("backend not found")
	}
	mountResponse := controller.Mount(mountRequest)

	fmt.Sprintf("Finished MountCommand command [%s]", k8sutils.GetContextRequestString(requestContext))
	return printResponse(mountResponse)
}

//UnmountCommand unmounts a given mountedDirectory
//<driver executable> unmount <mount dir> (v>=1.5)
type UnmountCommand struct {
	UnMount func() `short:"u" long:"unmount" description:"UnMount a volume Id to a path"`
}

func (u *UnmountCommand) Execute(args []string) error {
	requestContext := k8sutils.GetNewRequestContext()
	fmt.Sprintf("Starting UnmountCommand command [%s]", k8sutils.GetContextRequestString(requestContext))
	if len(args) < 1 {

		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Not enough arguments to unmount call out"),
		}
		return printResponse(response)
	}
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
	defer logs.InitFileLogger(logs.GetLogLevelFromString(config.LogLevel), path.Join(config.LogPath, k8sresources.UbiquityFlexLogFileName), logger_params)()
	controller, err := createController(config)

	if err != nil {
		panic("backend not found")
	}

	unmountRequest := k8sresources.FlexVolumeUnmountRequest{
		MountPath: mountDir,
		Context:   requestContext,
	}
	unmountResponse := controller.Unmount(unmountRequest)
	fmt.Sprintf("Finished UnmountCommand command [%s]", k8sutils.GetContextRequestString(requestContext))
	return printResponse(unmountResponse)
}

type TestUbiquityCommand struct {
	Test func() `short:"i" long:"init" description:"Initialize the plugin"`
}

func (i *TestUbiquityCommand) Execute(args []string) error {
	config, err := readConfig(*configFile)
	if err != nil {
		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to read config in Test Ubiquity %#v", err),
		}
		return printResponse(response)
	}
	defer logs.InitFileLogger(logs.GetLogLevelFromString(config.LogLevel), path.Join(config.LogPath, k8sresources.UbiquityFlexLogFileName), logger_params)()
	controller, err := createController(config)
	if err != nil {
		response := k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to create controller %#v", err),
		}
		return printResponse(response)
	}
	response := controller.TestUbiquity(config)
	return printResponse(response)
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
	var testUbiquityCommand TestUbiquityCommand

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
	parser.AddCommand("testubiquity",
		"Tests connectivity to ubiquity",
		"Tests connectivity to ubiquity",
		&testUbiquityCommand)

	_, err := parser.Parse()
	if err != nil {
		panic(err)
		os.Exit(1)
	}
}

func createController(config resources.UbiquityPluginConfig) (*controller.Controller, error) {
	logger := utils.SetupOldLogger(k8sresources.UbiquityFlexLogFileName)
	controller, err := controller.NewController(logger, config)
	return controller, err
}

func readConfig(configFile string) (resources.UbiquityPluginConfig, error) {
	var config resources.UbiquityPluginConfig
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		fmt.Printf("error decoding config file", err)
		return resources.UbiquityPluginConfig{}, err

	}
	// Create environment variables for some of the config params
	os.Setenv(remote.KeyUseSsl, strconv.FormatBool(config.SslConfig.UseSsl))
	os.Setenv(resources.KeySslMode, config.SslConfig.SslMode)
	os.Setenv(remote.KeyVerifyCA, config.SslConfig.VerifyCa)
	return config, nil
}

func printResponse(f k8sresources.FlexVolumeResponse) error {
	responseBytes, err := json.Marshal(f)
	if err != nil {
		return err
	}
	fmt.Printf("%s", string(responseBytes[:]))
	return nil
}

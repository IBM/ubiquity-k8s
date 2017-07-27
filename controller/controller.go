/**
 * Copyright 2017 IBM Corp.
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

package controller

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	k8sresources "github.com/IBM/ubiquity-k8s/resources"
	"github.com/IBM/ubiquity/remote"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
)

//Controller this is a structure that controls volume management
type Controller struct {
	Client resources.StorageClient
	logger *log.Logger
	exec   utils.Executor
}

//NewController allows to instantiate a controller
func NewController(logger *log.Logger, storageApiURL string, config resources.UbiquityPluginConfig) (*Controller, error) {

	remoteClient, err := remote.NewRemoteClient(logger, storageApiURL, config)
	if err != nil {
		return nil, err
	}
	return &Controller{logger: logger, Client: remoteClient, exec: utils.NewExecutor()}, nil
}

//NewControllerWithClient is made for unit testing purposes where we can pass a fake client
func NewControllerWithClient(logger *log.Logger, client resources.StorageClient, exec utils.Executor) *Controller {
	utils.NewExecutor()
	return &Controller{logger: logger, Client: client, exec: exec}
}

//Init method is to initialize the k8sresourcesvolume
func (c *Controller) Init(config resources.UbiquityPluginConfig) k8sresources.FlexVolumeResponse {
	c.logger.Println("controller-activate-start")
	defer c.logger.Println("controller-activate-end")

	activateRequest := resources.ActivateRequest{Backends: config.Backends}
	err := c.Client.Activate(activateRequest)
	if err != nil {
		return k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Plugin init failed %#v ", err),
			Device:  "",
		}

	}

	return k8sresources.FlexVolumeResponse{
		Status:  "Success",
		Message: "Plugin init successfully",
		Device:  "",
	}
}

//Attach method attaches a volume to a host
func (c *Controller) Attach(attachRequest map[string]string) k8sresources.FlexVolumeResponse {
	c.logger.Println("controller-attach-start")
	defer c.logger.Println("controller-attach-end")
	c.logger.Printf("attach-details %#v\n", attachRequest)
	var attachResponse k8sresources.FlexVolumeResponse
	volumeName, exists := attachRequest["volumeName"]
	if !exists {

		attachResponse = k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to attach volume: VolumeName not found : #%v", attachRequest),
			Device:  volumeName,
		}
		c.logger.Printf("Failed-to-attach-volume, VolumeName found %#v ", attachRequest)
		return attachResponse

	}

	getVolumeRequest := resources.GetVolumeRequest{Name: volumeName}
	_, err := c.Client.GetVolume(getVolumeRequest)

	if err != nil {
		return k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: "Failed checking volume, call create before attach",
			Device:  volumeName}

	}

	return k8sresources.FlexVolumeResponse{
		Status:  "Success",
		Message: "Volume already attached",
		Device:  volumeName,
	}

}

//WaitForAttach Waits for a volume to get attached to the node
func (c *Controller) WaitForAttach(waitForAttachRequest k8sresources.FlexVolumeWaitForAttachRequest) k8sresources.FlexVolumeResponse {
	c.logger.Println("controller-waitForAttach-start")
	defer c.logger.Println("controller-waitForAttach-end")
	return k8sresources.FlexVolumeResponse{
		Status:     "Success",
		Message:    "Volume attached",
		VolumeName: waitForAttachRequest.Name,
		Device:     waitForAttachRequest.Name,
		Attached:   true,
	}
}

//IsAttached checks if volume is attached
func (c *Controller) IsAttached(isAttachRequest k8sresources.FlexVolumeIsAttachedRequest) k8sresources.FlexVolumeResponse {
	c.logger.Println("controller-isAttached-start")
	defer c.logger.Println("controller-isAttached-end")
	return k8sresources.FlexVolumeResponse{
		Status:     "Success",
		Message:    "Volume is attached",
		VolumeName: isAttachRequest.Name,
		Device:     isAttachRequest.Name,
		Attached:   true,
	}
}

//Detach detaches the volume/ fileset from the pod
func (c *Controller) Detach(detachRequest k8sresources.FlexVolumeDetachRequest) k8sresources.FlexVolumeResponse {
	c.logger.Println("controller-detach-start")
	defer c.logger.Println("controller-detach-end")

	c.logger.Printf("detach-details %#v. (no operation done in the detach action)\n", detachRequest)

	// TODO : no-op for now, will change with latest k8sresources api update

	return k8sresources.FlexVolumeResponse{
		Status:  "Success",
		Message: "Volume detached successfully",
		Device:  detachRequest.Name,
	}
}

//MountDevice mounts a device in a given location
func (c *Controller) MountDevice(mountDeviceRequest k8sresources.FlexVolumeMountDeviceRequest) k8sresources.FlexVolumeResponse {
	c.logger.Println("controller-MountDevice-start")
	defer c.logger.Println("controller-MountDevice-end")
	return k8sresources.FlexVolumeResponse{
		Status:  "Success",
		Message: "Volume is mounted",
	}
}

//UnmountDevice checks if volume is unmounted
func (c *Controller) UnmountDevice(unmountDeviceRequest k8sresources.FlexVolumeUnmountDeviceRequest) k8sresources.FlexVolumeResponse {
	c.logger.Println("controller-UnmountDevice-start")
	defer c.logger.Println("controller-UnmountDevice-end")
	return k8sresources.FlexVolumeResponse{
		Status:  "Success",
		Message: "Volume is unmounted",
	}
}

//Mount method allows to mount the volume/fileset to a given location for a pod
func (c *Controller) Mount(mountRequest k8sresources.FlexVolumeMountRequest) k8sresources.FlexVolumeResponse {
	c.logger.Println("controller-mount-start")
	defer c.logger.Println("controller-mount-end")
	c.logger.Println(fmt.Sprintf("mountRequest [%#v]", mountRequest))

	attachRequest := resources.AttachRequest{Name: mountRequest.MountDevice, Host: getHost()}
	mountedPath, err := c.Client.Attach(attachRequest)

	if err != nil {
		msg := fmt.Sprintf("Failed to mount volume [%s], Error: %#v", mountRequest.MountDevice, err)
		c.logger.Println(msg)
		return k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: msg,
			Device:  "",
		}
	}
	dir := filepath.Dir(mountRequest.MountPath)

	c.logger.Printf("volume mounted at %s", mountedPath)
	k8sRequiredMountPoint := path.Join(mountRequest.MountPath, mountRequest.MountDevice)
	if _, err = os.Stat(k8sRequiredMountPoint); err != nil {
		if os.IsNotExist(err) {

			c.logger.Printf("creating volume directory %s", dir)
			err = os.MkdirAll(dir, 0777)
			if err != nil && !os.IsExist(err) {
				msg := fmt.Sprintf("Failed creating volume directory %#v", err)
				c.logger.Println(msg)

				return k8sresources.FlexVolumeResponse{
					Status:  "Failure",
					Message: msg,
					Device:  "",
				}

			}

			symLinkCommand := "/bin/ln"
			args := []string{"-s", mountedPath, mountRequest.MountPath}
			c.logger.Printf(fmt.Sprintf("creating slink from %s -> %s", mountedPath, mountRequest.MountPath))

			cmd := exec.Command(symLinkCommand, args...)
			_, err = cmd.Output()
			if err != nil {
				msg := fmt.Sprintf("Controller: mount failed to symlink %#v", err)
				c.logger.Println(msg)
				return k8sresources.FlexVolumeResponse{
					Status:  "Failure",
					Message: msg,
					Device:  "",
				}

			}
			msg := fmt.Sprintf("Volume mounted successfully to %s", mountedPath)
			c.logger.Println(msg)

			return k8sresources.FlexVolumeResponse{
				Status:  "Success",
				Message: msg,
				Device:  "",
			}
		}
		msg := fmt.Sprintf("Fail to mount because cannot state [%s]. Error : %#v ",
			k8sRequiredMountPoint,
			err,
		)
		c.logger.Println(msg)
		return k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: msg,
			Device:  "",
		}
	} else {
		c.logger.Println(fmt.Sprintf("Mount was done, but no need to slink [%s] because its already there",
			k8sRequiredMountPoint))

	}

	return k8sresources.FlexVolumeResponse{
		Status:  "Success",
		Message: fmt.Sprintf("Volume mounted successfully to %s", mountedPath),
		Device:  "",
	}

}

//Unmount methods unmounts the volume from the pod
func (c *Controller) Unmount(unmountRequest k8sresources.FlexVolumeUnmountRequest) k8sresources.FlexVolumeResponse {
	c.logger.Println("Controller: unmount start")
	defer c.logger.Println("Controller: unmount end")
	var detachRequest resources.DetachRequest
	var pvName string

	// Validate that the mountpoint is a symlink as ubiquity expect it to be
	realMountPoint, err := c.exec.EvalSymlinks(unmountRequest.MountPath)
	if err != nil {
		msg := fmt.Sprintf("Cannot execute umount because the mountPath [%s] is not a symlink as expected. Error: %#v", unmountRequest.MountPath, err)
		c.logger.Println(msg)
		return k8sresources.FlexVolumeResponse{Status: "Failure", Message: msg, Device: ""}
	}
	ubiquityMountPrefix := fmt.Sprintf(resources.PathToMountUbiquityBlockDevices, "")
	if strings.HasPrefix(realMountPoint, ubiquityMountPrefix) {
		// SCBE backend flow
		pvName = path.Base(unmountRequest.MountPath)

		detachRequest = resources.DetachRequest{Name: pvName, Host: getHost()}
		err = c.Client.Detach(detachRequest)
		if err != nil {
			msg := fmt.Sprintf(
				"Failed to unmount volume [%s] on mountpoint [%s]. Error: %#v",
				pvName,
				unmountRequest.MountPath,
				err)
			c.logger.Println(msg)
			return k8sresources.FlexVolumeResponse{Status: "Failure", Message: msg, Device: ""}
		}

		c.logger.Println(fmt.Sprintf("Removing the slink [%s] to the real mountpoint [%s]", unmountRequest.MountPath, realMountPoint))
		err := c.exec.Remove(unmountRequest.MountPath)
		if err != nil {
			msg := fmt.Sprintf("fail to remove slink %s. Error %#v", unmountRequest.MountPath, err)
			c.logger.Println(msg)
			return k8sresources.FlexVolumeResponse{Status: "Failure", Message: msg, Device: ""}
		}

	} else {

		listVolumeRequest := resources.ListVolumesRequest{}
		volumes, err := c.Client.ListVolumes(listVolumeRequest)
		if err != nil {
			msg := fmt.Sprintf("Error getting the volume list from ubiquity server %#v", err)
			c.logger.Println(msg)
			return k8sresources.FlexVolumeResponse{
				Status:  "Failure",
				Message: msg,
				Device:  "",
			}
		}

		volume, err := getVolumeForMountpoint(unmountRequest.MountPath, volumes)
		if err != nil {
			msg := fmt.Sprintf(
				"Error finding the volume with mountpoint [%s] from the list of ubiquity volumes %#v. Error is : %#v",
				unmountRequest.MountPath,
				volumes,
				err)
			c.logger.Println(msg)
			return k8sresources.FlexVolumeResponse{
				Status:  "Failure",
				Message: msg,
				Device:  "",
			}
		}

		detachRequest = resources.DetachRequest{Name: volume.Name}
		err = c.Client.Detach(detachRequest)
		if err != nil && err.Error() != "fileset not linked" {
			msg := fmt.Sprintf(
				"Failed to unmount volume [%s] on mountpoint [%s]. Error: %#v",
				volume.Name,
				unmountRequest.MountPath,
				err)
			c.logger.Println(msg)

			return k8sresources.FlexVolumeResponse{
				Status:  "Failure",
				Message: msg,
				Device:  "",
			}
		}

		pvName = volume.Name
	}

	msg := fmt.Sprintf(
		"Succeeded to umount volume [%s] on mountpoint [%s]",
		pvName,
		unmountRequest.MountPath,
	)
	c.logger.Println(msg)

	return k8sresources.FlexVolumeResponse{
		Status:  "Success",
		Message: "Volume unmounted successfully",
		Device:  "",
	}
}

func getVolumeForMountpoint(mountpoint string, volumes []resources.Volume) (resources.Volume, error) {
	for _, volume := range volumes {
		if volume.Mountpoint == mountpoint {
			return volume, nil
		}
	}
	return resources.Volume{}, fmt.Errorf("Volume not found")
}

//TODO check os.Host
func getHost() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hostname
}

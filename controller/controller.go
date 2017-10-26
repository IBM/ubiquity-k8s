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
	"strings"

	"bytes"
	k8sresources "github.com/IBM/ubiquity-k8s/resources"
	"github.com/IBM/ubiquity/remote"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
	"path/filepath"
)

//Controller this is a structure that controls volume management
type Controller struct {
	Client resources.StorageClient
	logger *log.Logger
	exec   utils.Executor
}

//NewController allows to instantiate a controller
func NewController(logger *log.Logger, config resources.UbiquityPluginConfig) (*Controller, error) {

	remoteClient, err := remote.NewRemoteClientSecure(logger, config)
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

	return k8sresources.FlexVolumeResponse{
		Status:  "Success",
		Message: "Plugin init successfully",
	}
}

//TestUbiquity method is to test connectivity to ubiquity
func (c *Controller) TestUbiquity(config resources.UbiquityPluginConfig) k8sresources.FlexVolumeResponse {
	c.logger.Println("controller-testubiquity-start")
	defer c.logger.Println("controller-testubiquity-end")

	activateRequest := resources.ActivateRequest{Backends: config.Backends}
	err := c.Client.Activate(activateRequest)
	if err != nil {
		return k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Test ubiquity failed %#v ", err),
		}

	}

	return k8sresources.FlexVolumeResponse{
		Status:  "Success",
		Message: "Test ubiquity successfully",
	}
}

//Attach method attaches a volume to a host
func (c *Controller) Attach(attachRequest k8sresources.FlexVolumeAttachRequest) k8sresources.FlexVolumeResponse {
	c.logger.Println("controller-attach-start")
	defer c.logger.Println("controller-attach-end")

	if attachRequest.Version == k8sresources.KubernetesVersion_1_5 {
		c.logger.Printf("k8s 1.5 attach just returning Success")
		return k8sresources.FlexVolumeResponse{
			Status: "Success",
		}
	}
	c.logger.Printf("For k8s version 1.6 and higher, Ubiquity just returns NOT supported for Attach API. This might change in the future")
	return k8sresources.FlexVolumeResponse{
		Status: "Not supported",
	}
}

//GetVolumeName checks if volume is attached
func (c *Controller) GetVolumeName(getVolumeNameRequest k8sresources.FlexVolumeGetVolumeNameRequest) k8sresources.FlexVolumeResponse {
	c.logger.Println("controller-isAttached-start")
	defer c.logger.Println("controller-isAttached-end")

	return k8sresources.FlexVolumeResponse{
		Status: "Not supported",
	}
}

//WaitForAttach Waits for a volume to get attached to the node
func (c *Controller) WaitForAttach(waitForAttachRequest k8sresources.FlexVolumeWaitForAttachRequest) k8sresources.FlexVolumeResponse {
	c.logger.Println("controller-waitForAttach-start")
	return k8sresources.FlexVolumeResponse{
		Status: "Not supported",
	}
}

//IsAttached checks if volume is attached
func (c *Controller) IsAttached(isAttachedRequest k8sresources.FlexVolumeIsAttachedRequest) k8sresources.FlexVolumeResponse {
	c.logger.Println("controller-isAttached-start")
	return k8sresources.FlexVolumeResponse{
		Status: "Not supported",
	}
}

//Detach detaches the volume/ fileset from the pod
func (c *Controller) Detach(detachRequest k8sresources.FlexVolumeDetachRequest) k8sresources.FlexVolumeResponse {
	c.logger.Println("controller-detach-start")
	defer c.logger.Println("controller-detach-end")
	if detachRequest.Version == k8sresources.KubernetesVersion_1_5 {
		return k8sresources.FlexVolumeResponse{
			Status: "Success",
		}
	}
	return k8sresources.FlexVolumeResponse{
		Status: "Not supported",
	}
}

//MountDevice mounts a device in a given location
func (c *Controller) MountDevice(mountDeviceRequest k8sresources.FlexVolumeMountDeviceRequest) k8sresources.FlexVolumeResponse {
	c.logger.Println("controller-MountDevice-start")
	defer c.logger.Println("controller-MountDevice-end")
	return k8sresources.FlexVolumeResponse{
		Status: "Not supported",
	}
}

//UnmountDevice checks if volume is unmounted
func (c *Controller) UnmountDevice(unmountDeviceRequest k8sresources.FlexVolumeUnmountDeviceRequest) k8sresources.FlexVolumeResponse {
	c.logger.Println("controller-UnmountDevice-start")
	defer c.logger.Println("controller-UnmountDevice-end")
	return k8sresources.FlexVolumeResponse{
		Status: "Not supported",
	}
}

//Mount method allows to mount the volume/fileset to a given location for a pod
func (c *Controller) Mount(mountRequest k8sresources.FlexVolumeMountRequest) k8sresources.FlexVolumeResponse {
	c.logger.Println("controller-mount-start")
	defer c.logger.Println("controller-mount-end")
	c.logger.Println(fmt.Sprintf("mountRequest [%#v]", mountRequest))
	var lnPath string
	attachRequest := resources.AttachRequest{Name: mountRequest.MountDevice, Host: getHost()}
	mountedPath, err := c.Client.Attach(attachRequest)

	if err != nil {
		msg := fmt.Sprintf("Failed to mount volume [%s], Error: %#v", mountRequest.MountDevice, err)
		c.logger.Println(msg)
		return k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: msg,
		}
	}
	if mountRequest.Version == k8sresources.KubernetesVersion_1_5 {
		//For k8s 1.5, by the time we do the attach/mount, the mountDir (MountPath) is not created trying to do mount and ln will fail because the dir is not found, so we need to create the directory before continuing
		dir := filepath.Dir(mountRequest.MountPath)
		c.logger.Printf("mountrequest.MountPath %s", mountRequest.MountPath)
		lnPath = mountRequest.MountPath
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
					}

				}
			}
		}
		// For k8s 1.6 and later kubelet creates a folder as the MountPath, including the volume name, whenwe try to create the symlink this will fail because the same name exists. This is why we need to remove it before continuing.
	} else {
		ubiquityMountPrefix := fmt.Sprintf(resources.PathToMountUbiquityBlockDevices, "")
		if strings.HasPrefix(mountedPath, ubiquityMountPrefix) {
			lnPath = mountRequest.MountPath
		} else {
			lnPath, _ = path.Split(mountRequest.MountPath)
		}
		c.logger.Printf("removing folder %s", mountRequest.MountPath)

		err = os.Remove(mountRequest.MountPath)
		if err != nil && !os.IsExist(err) {
			msg := fmt.Sprintf("Failed removing existing volume directory %#v", err)
			c.logger.Println(msg)

			return k8sresources.FlexVolumeResponse{
				Status:  "Failure",
				Message: msg,
			}

		}

	}
	symLinkCommand := "/bin/ln"
	args := []string{"-s", mountedPath, lnPath}
	c.logger.Printf(fmt.Sprintf("creating slink from %s -> %s", mountedPath, lnPath))

	var stderr bytes.Buffer
	cmd := exec.Command(symLinkCommand, args...)
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		msg := fmt.Sprintf("Controller: mount failed to symlink %#v", stderr.String())
		c.logger.Println(msg)
		return k8sresources.FlexVolumeResponse{
			Status:  "Failure",
			Message: msg,
		}

	}
	msg := fmt.Sprintf("Volume mounted successfully to %s", mountedPath)
	c.logger.Println(msg)

	return k8sresources.FlexVolumeResponse{
		Status:  "Success",
		Message: msg,
	}
}

//Unmount methods unmounts the volume from the pod
func (c *Controller) Unmount(unmountRequest k8sresources.FlexVolumeUnmountRequest) k8sresources.FlexVolumeResponse {
	c.logger.Println("Controller: unmount start")
	defer c.logger.Println("Controller: unmount end")
	c.logger.Printf("unmountRequest %#v", unmountRequest)
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

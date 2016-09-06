package core

import (
	"fmt"
	"log"

	common "github.ibm.com/almaden-containers/spectrum-common.git/core"
	"github.ibm.com/almaden-containers/spectrum-common.git/models"
)

type Controller struct {
	Client common.SpectrumClient
	log    *log.Logger
}

func NewController(logger *log.Logger, filesystem, mountpath string, dbClient *common.DatabaseClient) *Controller {
	return &Controller{log: logger, Client: common.NewSpectrumClient(logger, filesystem, mountpath, dbClient)}
}

func NewControllerWithClient(logger *log.Logger, client common.SpectrumClient) *Controller {
	return &Controller{log: logger, Client: client}
}

func (c *Controller) Init() *models.FlexVolumeResponse {
	c.log.Println("controller-activate-start")
	defer c.log.Println("controller-activate-end")

	return &models.FlexVolumeResponse{
		Status:  "Success",
		Message: "Plugin init successfully",
		Device:  "",
	}
}

func (c *Controller) Attach(attachRequest *models.FlexVolumeAttachRequest) *models.FlexVolumeResponse {
	c.log.Println("controller-attach-start")
	defer c.log.Println("controller-attach-end")
	c.log.Printf("attach-details %#v\n", attachRequest)
	var opts map[string]interface{}
	opts = map[string]interface{}{"fileset": attachRequest.VolumeId}

	var attachResponse *models.FlexVolumeResponse
	err := c.Client.CreateWithoutProvisioning(attachRequest.VolumeId, opts)
	if err != nil && err.Error() != "Volume already exists" {
		attachResponse = &models.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to attach volume: %#v", err),
			Device:  attachRequest.VolumeId,
		}
		c.log.Printf("Failed-to-attach-volume %#v ", err)
	} else if err != nil && err.Error() == "Volume already exists" {
		attachResponse = &models.FlexVolumeResponse{
			Status:  "Success",
			Message: "Volume already attached",
			Device:  attachRequest.VolumeId,
		}

	} else {
		attachResponse = &models.FlexVolumeResponse{
			Status:  "Success",
			Message: "Volume attached successfully",
			Device:  attachRequest.VolumeId,
		}
	}
	return attachResponse
}

func (c *Controller) Detach(detachRequest *models.GenericRequest) *models.FlexVolumeResponse {
	c.log.Println("controller-detach-start")
	defer c.log.Println("controller-detach-end")

	c.log.Printf("detach-details %#v\n", detachRequest)

	existingVolume, _, err := c.Client.Get(detachRequest.Name)

	if err != nil && err.Error() != "Cannot find info" {
		return &models.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to detach volume %#v", err),
			Device:  detachRequest.Name,
		}
	}

	if existingVolume != nil {
		err = c.Client.RemoveWithoutDeletingVolume(detachRequest.Name)
		if err != nil {
			return &models.FlexVolumeResponse{
				Status:  "Failure",
				Message: fmt.Sprintf("Failed to detach volume %#v", err),
				Device:  detachRequest.Name,
			}
		}

		return &models.FlexVolumeResponse{
			Status:  "Success",
			Message: "Volume detached successfully",
			Device:  detachRequest.Name,
		}
	}

	return &models.FlexVolumeResponse{
		Status:  "Failure",
		Message: "Volume not found",
		Device:  detachRequest.Name,
	}
}

func (c *Controller) Mount(mountRequest *models.FlexVolumeMountRequest) *models.FlexVolumeResponse {
	c.log.Println("controller-mount-start")
	defer c.log.Println("controller-mount-end")

	existingVolume, _, err := c.Client.Get(mountRequest.MountDevice)
	if err != nil && err.Error() != "Cannot find info" {
		return &models.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to mount volume %#v", err),
			Device:  "",
		}
		c.log.Printf("Error getting volume info %#v", err)
	}

	if existingVolume == nil {
		return &models.FlexVolumeResponse{
			Status:  "Failure",
			Message: "Failed to mount volume: volume not found",
			Device:  "",
		}
		c.log.Printf("Volume %s could not be found", mountRequest.MountDevice)
	}

	mountedPath, err := c.Client.Attach(mountRequest.MountDevice)
	if err != nil {
		return &models.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to mount volume %#v", err),
			Device:  "",
		}
		c.log.Printf("Failed to mount volume %#v", err)
	}

	return &models.FlexVolumeResponse{
		Status:  "Success",
		Message: fmt.Sprintf("Volume mounted successfully to %s", mountedPath),
		Device:  "",
	}
}

func (c *Controller) Unmount(unmountRequest *models.GenericRequest) *models.FlexVolumeResponse {
	c.log.Println("Controller: unmount start")
	defer c.log.Println("Controller: unmount end")

	filesetName, err := c.Client.GetFileSetForMountPoint(unmountRequest.Name)
	if err != nil {
		return &models.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Error finding the volume %#v", err),
			Device:  "",
		}
	}
	if filesetName == "" {
		return &models.FlexVolumeResponse{
			Status:  "Failure",
			Message: "Volume not found",
			Device:  "",
		}
	}
	c.log.Printf("Controller: unmount trying to unlink volume %s .", filesetName)
	err = c.Client.Detach(filesetName)
	if err != nil && err.Error() != "fileset not linked" {
		return &models.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to unmount volume %#v", err),
			Device:  "",
		}
	}

	return &models.FlexVolumeResponse{
		Status:  "Success",
		Message: "Volume unmounted successfully",
		Device:  "",
	}
}

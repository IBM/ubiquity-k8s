package core

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.ibm.com/almaden-containers/ubiquity.git/model"
	"github.ibm.com/almaden-containers/ubiquity.git/remote"
)

//Controller this is a structure that controls volume management
type Controller struct {
	Client model.StorageClient
	logger *log.Logger
}

//NewController allows to instantiate a controller
func NewController(logger *log.Logger, storageApiURL, backendName string) (*Controller, error) {
	remoteClient, err := remote.NewRemoteClient(logger, storageApiURL, backendName)
	if err != nil {
		return nil, err
	}
	return &Controller{logger: logger, Client: remoteClient}, nil
}

//NewControllerWithClient is made for unit testing purposes where we can pass a fake client
func NewControllerWithClient(logger *log.Logger, client model.StorageClient) *Controller {
	return &Controller{logger: logger, Client: client}
}

//Init method is to initialize the flexvolume, it is a no op right now
func (c *Controller) Init() model.FlexVolumeResponse {
	c.logger.Println("controller-activate-start")
	defer c.logger.Println("controller-activate-end")

	return model.FlexVolumeResponse{
		Status:  "Success",
		Message: "Plugin init successfully",
		Device:  "",
	}
}

//Attach method attaches a volume/ fileset to a pod
func (c *Controller) Attach(attachRequest model.FlexVolumeAttachRequest) model.FlexVolumeResponse {
	c.logger.Println("controller-attach-start")
	defer c.logger.Println("controller-attach-end")
	c.logger.Printf("attach-details %#v\n", attachRequest)
	var opts map[string]interface{}
	opts = map[string]interface{}{"fileset": attachRequest.VolumeId, "filesystem": attachRequest.Filesystem}

	var attachResponse model.FlexVolumeResponse
	err := c.Client.CreateVolume(attachRequest.VolumeId, opts)
	if err != nil && err.Error() != "Volume already exists" {
		attachResponse = model.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to attach volume: %#v", err),
			Device:  attachRequest.VolumeId,
		}
		c.logger.Printf("Failed-to-attach-volume %#v ", err)
	} else if err != nil && err.Error() == "Volume already exists" {
		attachResponse = model.FlexVolumeResponse{
			Status:  "Success",
			Message: "Volume already attached",
			Device:  attachRequest.VolumeId,
		}

	} else {
		attachResponse = model.FlexVolumeResponse{
			Status:  "Success",
			Message: "Volume attached successfully",
			Device:  attachRequest.VolumeId,
		}
	}
	return attachResponse
}

//Detach detaches the volume/ fileset from the pod
func (c *Controller) Detach(detachRequest model.FlexVolumeDetachRequest) model.FlexVolumeResponse {
	c.logger.Println("controller-detach-start")
	defer c.logger.Println("controller-detach-end")

	c.logger.Printf("detach-details %#v\n", detachRequest)

	err := c.Client.RemoveVolume(detachRequest.Name, false)
	if err != nil {
		return model.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to detach volume %#v", err),
			Device:  detachRequest.Name,
		}
	}

	return model.FlexVolumeResponse{
		Status:  "Success",
		Message: "Volume detached successfully",
		Device:  detachRequest.Name,
	}
}

//Mount method allows to mount the volume/fileset to a given location for a pod
func (c *Controller) Mount(mountRequest model.FlexVolumeMountRequest) model.FlexVolumeResponse {
	c.logger.Println("controller-mount-start")
	defer c.logger.Println("controller-mount-end")

	mountedPath, err := c.Client.Attach(mountRequest.MountDevice)

	if err != nil {
		c.logger.Printf("Failed to mount volume %#v", err)
		return model.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to mount volume %#v", err),
			Device:  "",
		}
	}
	dir := filepath.Dir(mountRequest.MountPath)

	c.logger.Printf("volume/ fileset mounted at %s", mountedPath)

	c.logger.Printf("creating volume directory %s", dir)
	err = os.MkdirAll(dir, 0777)
	if err != nil && !os.IsExist(err) {
		return model.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed creating volume directory %#v", err),
			Device:  "",
		}

	}

	symLinkCommand := "/bin/ln"
	args := []string{"-s", mountedPath, mountRequest.MountPath}
	cmd := exec.Command(symLinkCommand, args...)
	_, err = cmd.Output()
	if err != nil {
		c.logger.Printf("Controller: mount failed to symlink %#v", err)
		return model.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed running ln command %#v", err),
			Device:  "",
		}

	}

	return model.FlexVolumeResponse{
		Status:  "Success",
		Message: fmt.Sprintf("Volume mounted successfully to %s", mountedPath),
		Device:  "",
	}
}

//Unmount methods unmounts the volume/ fileset from the pod
func (c *Controller) Unmount(unmountRequest model.FlexVolumeUnmountRequest) model.FlexVolumeResponse {
	c.logger.Println("Controller: unmount start")
	defer c.logger.Println("Controller: unmount end")

	volumes, err := c.Client.ListVolumes()
	if err != nil {
		return model.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Error finding the volume %#v", err),
			Device:  "",
		}
	}

	volume, err := getVolumeForMountpoint(unmountRequest.MountPath, volumes)
	if err != nil {
		return model.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Error finding the volume %#v", err),
			Device:  "",
		}
	}

	err = c.Client.Detach(volume.Name)
	if err != nil && err.Error() != "fileset not linked" {
		return model.FlexVolumeResponse{
			Status:  "Failure",
			Message: fmt.Sprintf("Failed to unmount volume %#v", err),
			Device:  "",
		}
	}

	return model.FlexVolumeResponse{
		Status:  "Success",
		Message: "Volume unmounted successfully",
		Device:  "",
	}
}

func getVolumeForMountpoint(mountpoint string, volumes []model.VolumeMetadata) (model.VolumeMetadata, error) {

	for _, volume := range volumes {
		if volume.Mountpoint == mountpoint {
			return volume, nil
		}
	}
	return model.VolumeMetadata{}, fmt.Errorf("Volume not found")
}

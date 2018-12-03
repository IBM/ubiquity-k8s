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
	"path"
	"path/filepath"
	"strings"
	"time"
	"syscall"

	k8sresources "github.com/IBM/ubiquity-k8s/resources"
	"github.com/IBM/ubiquity/remote"
	"github.com/IBM/ubiquity/remote/mounter"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils"
	"github.com/IBM/ubiquity/utils/logs"
	"github.com/nightlyone/lockfile"
)
 
const(
	K8sPodsDirecotryName = "pods"
	k8sMountPointvolumeDirectoryName = "ibm~ubiquity-k8s-flex"
	FlexSuccessStr = "Success"
	FlexFailureStr = "Failure"
	FlexNotSupportedStr = "Not supported"
)

//Controller this is a structure that controls volume management
type Controller struct {
	Client            resources.StorageClient
	exec              utils.Executor
	logger            logs.Logger
	legacyLogger      *log.Logger
	config            resources.UbiquityPluginConfig
	mounterPerBackend map[string]resources.Mounter
	unmountFlock      lockfile.Lockfile
	mounterFactory    mounter.MounterFactory
}

func newController(logger *log.Logger, config resources.UbiquityPluginConfig, client resources.StorageClient, exec utils.Executor, mFactory mounter.MounterFactory) (*Controller, error) {
	unmountFlock, err := lockfile.New(filepath.Join(os.TempDir(), "ubiquity.unmount.lock"))
	if err != nil {
		panic(err)
	}

	return &Controller{
		logger:            logs.GetLogger(),
		legacyLogger:      logger,
		Client:            client,
		exec:              exec,
		config:            config,
		mounterPerBackend: make(map[string]resources.Mounter),
		unmountFlock:      unmountFlock,
		mounterFactory:    mFactory,
	}, nil
}

//NewController allows to instantiate a controller
func NewController(logger *log.Logger, config resources.UbiquityPluginConfig) (*Controller, error) {
	remoteClient, err := remote.NewRemoteClientSecure(logger, config)
	if err != nil {
		return nil, err
	}
	
	return newController(logger, config, remoteClient, utils.NewExecutor(), mounter.NewMounterFactory())
}

//NewControllerWithClient is made for unit testing purposes where we can pass a fake client
func NewControllerWithClient(logger *log.Logger, config resources.UbiquityPluginConfig, client resources.StorageClient, exec utils.Executor, mFactory mounter.MounterFactory) *Controller {
	controller, _ :=  newController(logger, config, client, exec, mFactory)
	return controller
}

//Init method is to initialize the k8sresourcesvolume
func (c *Controller) Init(config resources.UbiquityPluginConfig) k8sresources.FlexVolumeResponse {
	defer c.logger.Trace(logs.DEBUG)()

	response := k8sresources.FlexVolumeResponse{
		Status:  "Success",
		Message: "Plugin init successfully",
	}

	c.logger.Debug("", logs.Args{{"response", response}})
	return response
}

//TestUbiquity method is to test connectivity to ubiquity
func (c *Controller) TestUbiquity(config resources.UbiquityPluginConfig) k8sresources.FlexVolumeResponse {
	defer c.logger.Trace(logs.DEBUG)()
	var response k8sresources.FlexVolumeResponse

	activateRequest := resources.ActivateRequest{Backends: config.Backends}
	c.logger.Debug("", logs.Args{{"request", activateRequest}})

	err := c.doActivate(activateRequest)
	if err != nil {
		response = c.failureFlexVolumeResponse(err, "Test ubiquity failed")
	} else {
		response = c.successFlexVolumeResponse("Test ubiquity successfully")
	}

	c.logger.Debug("", logs.Args{{"response", response}})
	return response
}

//Attach method attaches a volume to a host
func (c *Controller) Attach(attachRequest k8sresources.FlexVolumeAttachRequest) k8sresources.FlexVolumeResponse {
	go_id := logs.GetGoID()
	logs.GoIdToRequestIdMap.Store(go_id, attachRequest.Context)
	defer logs.GetDeleteFromMapFunc(go_id)
	defer c.logger.Trace(logs.DEBUG)()
	var response k8sresources.FlexVolumeResponse
	var err error
	c.logger.Debug("", logs.Args{{"request", attachRequest}})

	getVolumeRequest := resources.GetVolumeRequest{Name:attachRequest.Name}
	volume, err :=  c.Client.GetVolume(getVolumeRequest)

	if err != nil {
		errormsg := fmt.Sprintf("Failed to get Volume details [%s]", attachRequest.Name)
		response = c.failureFlexVolumeResponse(err, errormsg)
		return response
	}

	//TODO: later we should consider to get the volume backend by using the attachRequest.opts instead of calling to ubiquity server
	if (volume.Backend == resources.SpectrumScale) {
		response = c.notSupportedFlexVolumeResponse("Flex Attach API is not supported for spectrum-scale backend, because the backend assume that volume is already attached to all nodes")
		return response
	}

	err = c.doAttach(attachRequest)
	if err != nil {
		msg := fmt.Sprintf("Failed to attach volume [%s]", attachRequest.Name)
		response = c.failureFlexVolumeResponse(err, msg)
	} else {
		response = c.successFlexVolumeResponse("")
	}

	c.logger.Debug("", logs.Args{{"response", response}})
	return response
}

//GetVolumeName checks if volume is attached
func (c *Controller) GetVolumeName(getVolumeNameRequest k8sresources.FlexVolumeGetVolumeNameRequest) k8sresources.FlexVolumeResponse {
	defer c.logger.Trace(logs.DEBUG)()
	var response k8sresources.FlexVolumeResponse
	c.logger.Debug("", logs.Args{{"request", getVolumeNameRequest}})

	response = k8sresources.FlexVolumeResponse{
		Status: "Not supported",
	}

	c.logger.Debug("", logs.Args{{"response", response}})
	return response
}

//WaitForAttach Waits for a volume to get attached to the node
func (c *Controller) WaitForAttach(waitForAttachRequest k8sresources.FlexVolumeWaitForAttachRequest) k8sresources.FlexVolumeResponse {
	go_id := logs.GetGoID()
	logs.GoIdToRequestIdMap.Store(go_id, waitForAttachRequest.Context)
	defer logs.GetDeleteFromMapFunc(go_id)
	defer c.logger.Trace(logs.DEBUG)()
	var response k8sresources.FlexVolumeResponse
	c.logger.Debug("", logs.Args{{"request", waitForAttachRequest}})

	response = k8sresources.FlexVolumeResponse{
		Status: "Not supported",
	}

	c.logger.Debug("", logs.Args{{"response", response}})
	return response
}

//IsAttached checks if volume is attached
func (c *Controller) IsAttached(isAttachedRequest k8sresources.FlexVolumeIsAttachedRequest) k8sresources.FlexVolumeResponse {
	go_id := logs.GetGoID()
	logs.GoIdToRequestIdMap.Store(go_id, isAttachedRequest.Context)
	defer logs.GetDeleteFromMapFunc(go_id)
	defer c.logger.Trace(logs.DEBUG)()
	var response k8sresources.FlexVolumeResponse
	var volume resources.Volume
	var err error
	c.logger.Debug("", logs.Args{{"request", isAttachedRequest}})

	volName, ok := isAttachedRequest.Opts["volumeName"]

	if !ok {
		err := fmt.Errorf("volumeName not found in isAttachedRequest")
		response = c.failureFlexVolumeResponse(err, "")
		return response
	}

	getVolumeRequest := resources.GetVolumeRequest{Name: volName}
	if volume, err = c.Client.GetVolume(getVolumeRequest); err != nil {
		if strings.Contains(err.Error(), resources.VolumeNotFoundErrorMsg) {
			warningMsg := fmt.Sprintf("%s (backend error=%v)", IdempotentIsAttachedSkipOnVolumeNotExistWarnigMsg, err)
			c.logger.Warning(warningMsg)
			response = k8sresources.FlexVolumeResponse{
            	Status:   "Success",
            	Attached: false,
				Message: warningMsg,
        	}
			return response
		}
		return c.failureFlexVolumeResponse(err, "")
	}

	if (volume.Backend == resources.SpectrumScale) {
		response = c.notSupportedFlexVolumeResponse("Flex IsAttached API is not supported for spectrum-scale backend, because the backend assume that volume is always attached")
		return response
	}

	isAttached, err := c.doIsAttached(isAttachedRequest)
	if err != nil {
		msg := fmt.Sprintf("Failed to check IsAttached volume [%s]", isAttachedRequest.Name)
		response = c.failureFlexVolumeResponse(err, msg)
	} else {
		response = k8sresources.FlexVolumeResponse{
			Status:   "Success",
			Attached: isAttached,
		}
	}

	c.logger.Debug("", logs.Args{{"response", response}})
	return response
}

//Detach detaches the volume/ fileset from the pod
func (c *Controller) Detach(detachRequest k8sresources.FlexVolumeDetachRequest) k8sresources.FlexVolumeResponse {
	go_id := logs.GetGoID()
	logs.GoIdToRequestIdMap.Store(go_id, detachRequest.Context)
	defer logs.GetDeleteFromMapFunc(go_id)
	defer c.logger.Trace(logs.DEBUG)()
	var response k8sresources.FlexVolumeResponse
	var volume resources.Volume
	var err error
	c.logger.Debug("", logs.Args{{"request", detachRequest}})

	if detachRequest.Version == k8sresources.KubernetesVersion_1_5 {
		c.logger.Debug("legacy detach (skipping)")
		response = k8sresources.FlexVolumeResponse{
			Status: "Success",
		}
	} else {
		getVolumeRequest := resources.GetVolumeRequest{Name: detachRequest.Name}
		if volume, err = c.Client.GetVolume(getVolumeRequest); err != nil {
	        if strings.Contains(err.Error(), resources.VolumeNotFoundErrorMsg) {
				warningMsg := fmt.Sprintf("%s (backend error=%v)", IdempotentDetachSkipOnVolumeNotExistWarnigMsg, err)
				c.logger.Warning(warningMsg)
				return c.successFlexVolumeResponse(warningMsg)
	        }
	        return c.failureFlexVolumeResponse(err, "")
	    }

		if (volume.Backend == resources.SpectrumScale) {
			response = c.notSupportedFlexVolumeResponse("Flex Detach API is not supported for spectrum-scale backend, because the backend assume that volume is always remain attached")
			return response
		}

		err = c.doDetach(detachRequest, true)
		if err != nil {
			msg := fmt.Sprintf(
				"Failed to detach volume [%s] from host [%s]. Error: %#v",
				detachRequest.Name,
				detachRequest.Host,
				err)
			response = k8sresources.FlexVolumeResponse{Status: "Failure", Message: msg, Device: ""}
		} else {
			response = k8sresources.FlexVolumeResponse{
				Status: "Success",
			}
		}
	}

	c.logger.Debug("", logs.Args{{"response", response}})
	return response
}

//MountDevice mounts a device in a given location
func (c *Controller) MountDevice(mountDeviceRequest k8sresources.FlexVolumeMountDeviceRequest) k8sresources.FlexVolumeResponse {
	go_id := logs.GetGoID()
	logs.GoIdToRequestIdMap.Store(go_id, mountDeviceRequest.Context)
	defer logs.GetDeleteFromMapFunc(go_id)
	defer c.logger.Trace(logs.DEBUG)()
	var response k8sresources.FlexVolumeResponse
	c.logger.Debug("", logs.Args{{"request", mountDeviceRequest}})

	response = k8sresources.FlexVolumeResponse{
		Status: "Not supported",
	}

	c.logger.Debug("", logs.Args{{"response", response}})
	return response
}

//UnmountDevice checks if volume is unmounted
func (c *Controller) UnmountDevice(unmountDeviceRequest k8sresources.FlexVolumeUnmountDeviceRequest) k8sresources.FlexVolumeResponse {
	go_id := logs.GetGoID()
	logs.GoIdToRequestIdMap.Store(go_id, unmountDeviceRequest.Context)
	defer logs.GetDeleteFromMapFunc(go_id)
	defer c.logger.Trace(logs.DEBUG)()
	var response k8sresources.FlexVolumeResponse
	c.logger.Debug("", logs.Args{{"request", unmountDeviceRequest}})

	response = k8sresources.FlexVolumeResponse{
		Status: "Not supported",
	}

	c.logger.Debug("", logs.Args{{"response", response}})
	return response
}


//Mount method allows to mount the volume/fileset to a given location for a pod
func (c *Controller) Mount(mountRequest k8sresources.FlexVolumeMountRequest) k8sresources.FlexVolumeResponse {
	go_id := logs.GetGoID()
	logs.GoIdToRequestIdMap.Store(go_id, mountRequest.Context)
	defer logs.GetDeleteFromMapFunc(go_id)
	defer c.logger.Trace(logs.DEBUG)()
	var response k8sresources.FlexVolumeResponse
	c.logger.Debug("", logs.Args{{"request", mountRequest}})
	
	flockName := fmt.Sprintf("ubiquity.mount.%s.lock", mountRequest.MountDevice)
	mountFlock, err := lockfile.New(filepath.Join(os.TempDir(), flockName ))
	if err != nil {
		panic(err)
	}

	for {
		err := mountFlock.TryLock()
		if err == nil {
			break
		}
		c.logger.Debug("mountFlock.TryLock failed", logs.Args{{"error", err}})
		time.Sleep(time.Duration(500 * time.Millisecond))
	}
	c.logger.Debug("Got mountFlock for volume.", logs.Args{{"volume", mountRequest.MountDevice}})
	defer mountFlock.Unlock()
	
	// TODO check if volume exist first and what its backend type
	mountedPath, err := c.doMount(mountRequest)
	if err != nil {
		response = c.failureFlexVolumeResponse(err, "")
	} else {
		err = c.doAfterMount(mountRequest, mountedPath)
		if err != nil {
			response = c.failureFlexVolumeResponse(err, "")
		} else {
			response = k8sresources.FlexVolumeResponse{
				Status: "Success",
			}
		}
	}

	c.logger.Debug("", logs.Args{{"response", response}})
	return response
}

func (c *Controller) successFlexVolumeResponse(msg string) k8sresources.FlexVolumeResponse {
	defer c.logger.Trace(logs.DEBUG)()
	response := k8sresources.FlexVolumeResponse{
		Status:  FlexSuccessStr,
		Message: msg,
	}
	c.logger.Info(fmt.Sprintf("%#v", response))
	return response
}

func (c *Controller) failureFlexVolumeResponse(err error, additionalMsg string) k8sresources.FlexVolumeResponse {
	defer c.logger.Trace(logs.DEBUG)()

	response := k8sresources.FlexVolumeResponse{
		Status:  FlexFailureStr,
		Message: additionalMsg + err.Error(),
	}
	c.logger.Error(fmt.Sprintf("%#v", response))
	return response
}

func (c *Controller) notSupportedFlexVolumeResponse(msg string) k8sresources.FlexVolumeResponse {
	defer c.logger.Trace(logs.DEBUG)()
	response := k8sresources.FlexVolumeResponse{
		Status:  FlexNotSupportedStr,
		Message: msg,
	}
	c.logger.Info(fmt.Sprintf("%#v", response))
	return response
}

func (c *Controller) checkSlinkBeforeUmount(k8sPVDirectoryPath string, realMountedPath string) (bool, error) {
	/*
	   Return <isSlink, error>
	   true,  nil   : slink exist as expected
	   false, nil   : slink not exist, which indicate for idempotent issue
	   true,  error : its slink but error during evaluate it or its slink but not to the right place
	   false, error : its something different then slink
	*/
	defer c.logger.Trace(logs.DEBUG)()

	// Identify the PV directory by using Lstat and then handle all idempotent cases (use Lstat to get the dir or slink detail and not the evaluation of it)
	fileInfo, err := c.exec.Lstat(k8sPVDirectoryPath)
	if err != nil {
		if c.exec.IsNotExist(err) {
			// The k8s PV directory not exist (its a rare case and indicate on idempotent flow)
			c.logger.Warning("PV directory(k8s-mountpoint) does not exist.", logs.Args{{"k8s-mountpoint", k8sPVDirectoryPath}, {"should-point-to-mountpoint", realMountedPath}}) 
			return false, nil                                                                                                                                                 // Idempotent flow
		} else {
			// Maybe some permissions issue
			return false, c.logger.ErrorRet(err, "Controller: failed to identify PV directory(k8s-mountpoint)", logs.Args{{"k8s-mountpoint", k8sPVDirectoryPath}})
		}
	} else if c.exec.IsSlink(fileInfo) {
		// Its already slink so check if slink is ok and skip else raise error
		evalSlink, err := c.exec.EvalSymlinks(k8sPVDirectoryPath)
		if err != nil {
			message := "Controller: Idempotent - failed to eval the slink of the PV directory(k8s-mountpoint)"
			if os.IsNotExist(err){
				c.logger.Warning(message, logs.Args{{"k8s-mountpoint", k8sPVDirectoryPath},{"error" , err}})
				return true, nil
			}
			return true, c.logger.ErrorRet(err, message, logs.Args{{"k8s-mountpoint", k8sPVDirectoryPath}})
		}
		if evalSlink == realMountedPath {
			c.logger.Info("PV directory(k8s-mountpoint) is slink that point to the right mountpoint.", logs.Args{{"k8s-mountpoint", k8sPVDirectoryPath}, {"mountpoint", realMountedPath}})
			return true, nil // Good path
		} else {
			// Very edge case, where the slink is point to wrong mountpoint
			return true, c.logger.ErrorRet(
				&wrongSlinkError{slink: k8sPVDirectoryPath, wrongPointTo: evalSlink, expectedPointTo: realMountedPath},
				"failed")
		}
	} else {
		return false, c.logger.ErrorRet(&k8sPVDirectoryIsNotSlinkError{k8sPVDirectoryPath, fileInfo}, "failed")
	}

}

func (c *Controller) getRealMountpointForPvByBackend(volumeBackend string, volumeConfig map[string]interface{}) (string, error) {
	// TODO we should create agnostic function based on backend that returns the real mountpoint
	defer c.logger.Trace(logs.DEBUG)()
	if volumeBackend == resources.SCBE {
		return fmt.Sprintf(resources.PathToMountUbiquityBlockDevices, volumeConfig["Wwn"].(string)), nil
	} else if volumeBackend == resources.SpectrumScale {
                return volumeConfig["mountpoint"].(string), nil
		//TODO: scale should remove the mountpoint
	} else {
		return "", &PvBackendNotSupportedError{Backend: volumeBackend}
	}
}
func (c *Controller) doUnmount(k8sPVDirectoryPath string, volumeBackend string, volumeConfig map[string]interface{}, mounter resources.Mounter) error {
	/*
		Call to unmount mounter if slink exist and then delete the slink
		if slink not exist skip unmount (indicate for idempotent flow).
	*/
	defer c.logger.Trace(logs.DEBUG)()
	var realMountedPath string
	var slinkExist bool
	var err error
	if realMountedPath, err = c.getRealMountpointForPvByBackend(volumeBackend, volumeConfig); err != nil {
		return c.logger.ErrorRet(err, "getRealMountpointForPvByBackend failed")
	}

	// ------------------------------------
	if slinkExist, err = c.checkSlinkBeforeUmount(k8sPVDirectoryPath, realMountedPath); err != nil {
		return c.logger.ErrorRet(err, "checkSlinkBeforeUmount failed")
	}
	if slinkExist {

		ubUnmountRequest := resources.UnmountRequest{VolumeConfig: volumeConfig} // TODO need to add to the request the real mountpoint to umount
		if err := mounter.Unmount(ubUnmountRequest); err != nil {
			return c.logger.ErrorRet(err, "mounter.Unmount failed")
		}

		c.logger.Debug(fmt.Sprintf("Removing the slink [%s] to the real mountpoint [%s]", k8sPVDirectoryPath, realMountedPath))
		if err := c.exec.Remove(k8sPVDirectoryPath); err != nil {
			// the slink must exist at this point
			return c.logger.ErrorRet(err, "fail to remove slink "+k8sPVDirectoryPath)
		}
	} else {
		c.logger.Warning("PV directory(k8s-mountpoint) does not exist. Idempotent - skip unmount flow",
			logs.Args{{"k8s-mountpoint", k8sPVDirectoryPath}, {"should-point-to-mountpoint", realMountedPath}})
	}
	return nil // Finish successfully to umount
}

//Unmount methods unmounts the volume from the pod
func (c *Controller) Unmount(unmountRequest k8sresources.FlexVolumeUnmountRequest) k8sresources.FlexVolumeResponse {
	go_id := logs.GetGoID()
	logs.GoIdToRequestIdMap.Store(go_id, unmountRequest.Context)
	defer logs.GetDeleteFromMapFunc(go_id)
	defer c.logger.Trace(logs.DEBUG, logs.Args{{"unmountRequest", unmountRequest}})()
	k8sPVDirectoryPath := unmountRequest.MountPath

	// locking for concurrent rescans and reduce rescans if no need
	c.logger.Debug("Ask for unmountFlock for mountpath", logs.Args{{"mountpath", k8sPVDirectoryPath}})
	for {
		err := c.unmountFlock.TryLock()
		if err == nil {
			break
		}
		c.logger.Debug("unmountFlock.TryLock failed", logs.Args{{"error", err}})
		time.Sleep(time.Duration(500 * time.Millisecond))
	}
	c.logger.Debug("Got unmountFlock for mountpath", logs.Args{{"mountpath", k8sPVDirectoryPath}})
	defer c.unmountFlock.Unlock()
	defer c.logger.Debug("Released unmountFlock for mountpath", logs.Args{{"mountpath", k8sPVDirectoryPath}})

	var mounter resources.Mounter
	var volumeConfig map[string]interface{}
	var volume resources.Volume
	var err error

	// GetVolume by pv name to identify if it exist in ubiquity DB and to receive the backend
	pvName := path.Base(k8sPVDirectoryPath) // Assumption that the k8s mountpoint directory contains(basename) the pv name it self.
	getVolumeRequest := resources.GetVolumeRequest{Name: pvName}
	if volume, err = c.Client.GetVolume(getVolumeRequest); err != nil {
		if strings.Contains(err.Error(), resources.VolumeNotFoundErrorMsg) {
			warningMsg := fmt.Sprintf("%s (backend error=%v)", IdempotentUnmountSkipOnVolumeNotExistWarnigMsg, err)
			c.logger.Warning(warningMsg)
			return c.successFlexVolumeResponse(warningMsg)
		}
		return c.failureFlexVolumeResponse(err, "")
	}

	if mounter, err = c.getMounterForBackend(volume.Backend, unmountRequest.Context); err != nil {
		return c.failureFlexVolumeResponse(err, "Error determining mounter for volume. ")
	}

	getVolumeConfigRequest := resources.GetVolumeConfigRequest{Name: pvName}
	if volumeConfig, err = c.Client.GetVolumeConfig(getVolumeConfigRequest); err != nil {
		return c.failureFlexVolumeResponse(err, "Error unmount for volume. ")
	}

	if err := c.doUnmount(k8sPVDirectoryPath, volume.Backend, volumeConfig, mounter); err != nil {
		return c.failureFlexVolumeResponse(err, "")
	}

	if (volume.Backend == resources.SpectrumScale) {
		return c.successFlexVolumeResponse("")
	} else if (volume.Backend == resources.SCBE) {
	// Do legacy detach (means trigger detach as part of the umount from the k8s node)
		if err := c.doLegacyDetach(unmountRequest); err != nil {
			return c.failureFlexVolumeResponse(err, "")
		}
	} else {
		err = c.logger.ErrorRet(&PvBackendNotSupportedError{Backend: volume.Backend}, "failed")
		return c.failureFlexVolumeResponse(err, "")
	}

	return c.successFlexVolumeResponse("")
}

func (c *Controller) doLegacyDetach(unmountRequest k8sresources.FlexVolumeUnmountRequest) error {
	defer c.logger.Trace(logs.DEBUG)()
	var err error

	pvName := path.Base(unmountRequest.MountPath)
	detachRequest := k8sresources.FlexVolumeDetachRequest{Name: pvName, Context: unmountRequest.Context}
	err = c.doDetach(detachRequest, false)
	if err != nil {
		return c.logger.ErrorRet(err, "failed")
	} else {
		err = c.doAfterDetach(detachRequest)
		if err != nil {
			return c.logger.ErrorRet(err, "failed")
		}
	}

	return nil
}

func (c *Controller) getMounterForBackend(backend string, requestContext resources.RequestContext) (resources.Mounter, error) {
	defer c.logger.Trace(logs.DEBUG)()
	var err error
	mounterInst, ok := c.mounterPerBackend[backend]
	if ok {
		// mounter already exist in the controller backend list
		return mounterInst, nil
	} else {
		// mounter not exist in the controller backend list, so get it now
		c.mounterPerBackend[backend], err = c.mounterFactory.GetMounterPerBackend(backend, c.legacyLogger, c.config, requestContext)
		if err != nil {
			return nil, err
		}
	}
	return c.mounterPerBackend[backend], nil
}

func (c *Controller) getMountpointForVolume(mountRequest k8sresources.FlexVolumeMountRequest, volumeConfig map[string]interface{}, volumeBackend string) (string, error) {

	defer c.logger.Trace(logs.DEBUG)()

	var volumeMountPoint string
	var ok bool

	if (volumeBackend == resources.SpectrumScale) {
		volumeMountPoint, ok = volumeConfig["mountpoint"].(string)
		if !ok {
			return "", c.logger.ErrorRet(&SpectrumScaleMissingMntPtVolumeError{VolumeName: mountRequest.MountDevice}, "failed")
		}
	} else if (volumeBackend == resources.SCBE) {
		wwn, ok := mountRequest.Opts["Wwn"]
		if !ok {
			err := fmt.Errorf(MissingWwnMountRequestErrorStr)
			return "", c.logger.ErrorRet(err, "failed")
		}
		volumeMountPoint = fmt.Sprintf(resources.PathToMountUbiquityBlockDevices, wwn)
	} else {
		return "", c.logger.ErrorRet(&PvBackendNotSupportedError{Backend: volumeBackend}, "failed")
	}

	return volumeMountPoint, nil
}

func (c *Controller) prepareUbiquityMountRequest(mountRequest k8sresources.FlexVolumeMountRequest, volumeBackend string) (resources.MountRequest, error) {
	/*
		Prepare the mounter.Mount request
	*/
	defer c.logger.Trace(logs.DEBUG)()

	// Prepare request for mounter - step1 get volume's config from ubiquity
	getVolumeConfigRequest := resources.GetVolumeConfigRequest{Name: mountRequest.MountDevice, Context: mountRequest.Context}
	volumeConfig, err := c.Client.GetVolumeConfig(getVolumeConfigRequest)
	if err != nil {
		return resources.MountRequest{}, c.logger.ErrorRet(err, "Client.GetVolumeConfig failed")
	}

	volumeMountpoint, err := c.getMountpointForVolume(mountRequest, volumeConfig, volumeBackend)
	if err != nil {
		return resources.MountRequest{}, err
	}

	ubMountRequest := resources.MountRequest{Mountpoint: volumeMountpoint, VolumeConfig: volumeConfig, Context: mountRequest.Context}
	return ubMountRequest, nil
}

func (c *Controller) getMounterByPV(mountRequest k8sresources.FlexVolumeMountRequest) (resources.Mounter, string, error) {
	defer c.logger.Trace(logs.DEBUG)()

	getVolumeRequest := resources.GetVolumeRequest{Name: mountRequest.MountDevice, Context: mountRequest.Context}
	volume, err := c.Client.GetVolume(getVolumeRequest)
	if err != nil {
		return nil, "", c.logger.ErrorRet(err, "GetVolume failed")
	}
	mounter, err := c.getMounterForBackend(volume.Backend, mountRequest.Context)
	if err != nil {
		return nil, "", c.logger.ErrorRet(err, "getMounterForBackend failed")
	}

	return mounter, volume.Backend, nil
}

func getK8sPodsBaseDir(k8sMountPoint string) (string, error ){
	/*
		function is going to get from this kind of dir:
			/var/lib/kubelet/pods/1f94f1d9-8f36-11e8-b227-005056a4d4cb/volumes/ibm~ubiquity-k8s-flex/...
		to /var/lib/kubelet/pods/
		note: /var/lib can be changed in configuration so we need to get the base dir at runtime
	*/ 
	tempMountPoint := k8sMountPoint
	for i := 0 ; i < 4 ; i++ {
		tempMountPoint = filepath.Dir(tempMountPoint)
	}
		
	if filepath.Base(tempMountPoint) != K8sPodsDirecotryName{
		return "", &WrongK8sDirectoryPathError{k8sMountPoint}
	}
	
	return tempMountPoint, nil
}

func checkSlinkIsActuallyInUse(k8sMountPoint string, logger logs.Logger, executer utils.Executor) (bool, error){
	slinkStat, err := executer.Stat(k8sMountPoint)  // TODO: Lstat?
	if err != nil{
		return false, logger.ErrorRet(err, "Failed to get stat from k8s slink file.", logs.Args{{"k8sMountPoint", k8sMountPoint}})
	}
	
	rootDirStat, err := executer.Lstat(k8sMountPoint + "/..")
	if err != nil{
		return false, logger.ErrorRet(err, "Failed to get stat from root directory.", logs.Args{{"directory", k8sMountPoint}})
	}
	
	if slinkStat.Sys().(*syscall.Stat_t).Dev != rootDirStat.Sys().(*syscall.Stat_t).Dev {
		return true, nil
	}	 
	
	return false, nil
}


func checkSlinkAlreadyExistsOnMountPoint(mountPoint string, k8sMountPoint string, logger logs.Logger, executer utils.Executor) (error){
	/*
		the mountpoint parameter is the actual mountpoint we are pointing to: /ubiquity/WWN
		the k8sMountPoint is the path to the link we want to create from the /var/lib/kubelet directory to the mountpoint
	*/
	defer logger.Trace(logs.INFO, logs.Args{{"mountPoint", mountPoint}, {"k8sMountPoint", k8sMountPoint}})()
	
	k8sBaseDir, err := getK8sPodsBaseDir(k8sMountPoint)	
	if err != nil{
		return logger.ErrorRet(err, "Failed to get k8s pods base dir.", logs.Args{{"k8sMountPoint", k8sMountPoint}})
	}
		
	file_pattern := filepath.Join(k8sBaseDir, "*","volumes", k8sMountPointvolumeDirectoryName,"*")
	files, err := executer.GetGlobFiles(file_pattern)
	if err != nil{
		return logger.ErrorRet(err, "Failed to get files that match the pattern.", logs.Args{{"file_pattern", file_pattern}})
	}
	
	slinks := []string{}
	if len(files) == 0 {
		logger.Debug("There is no Pod that uses ibm flex PV on this node (No files matched the given pattern were found).", logs.Args{{"pattern", file_pattern}})
		return nil
	}
	
	mountStat, err := executer.Stat(mountPoint)
	if err != nil {
		if os.IsNotExist(err){
			logger.Debug("Mount point path does not exist yet.", logs.Args{{"mountpoint",mountPoint}})
			return nil
		}
		return logger.ErrorRet(err, "Failed to get stat for mount point file.", logs.Args{{"file", mountPoint}})
	}
	
	// go over the files and check if any of them is the same as our destinated mountpoint
	// this checks if any of the solinks are pointing to the PV we want to mount.
	for _, file := range files {
		fileStat, err := executer.Stat(file)
		if err != nil{
			logger.Warning("Failed to get stat for file.", logs.Args{{"file", file}})
			continue
		}
		
		isSameFile := executer.IsSameFile(fileStat, mountStat)
		if isSameFile{
			isInUse, err := checkSlinkIsActuallyInUse(k8sMountPoint, logger, executer)
			if err != nil {
				logger.Warning("Failed to check whether slink is in use.", logs.Args{{"file", file}})
				continue
			}
			if isInUse{
				slinks = append(slinks, file)
			} else {
				logger.Warning("found un mounted slink to current mount point", logs.Args{{"file", file}})
			}
			
		}
	}
		
	if len(slinks) == 0 {
		logger.Debug("There were no soft links pointing to given mountpoint.", logs.Args{{"mountpoint", mountPoint}})
		return nil
	}
	
	if len(slinks) > 1 {
		return logger.ErrorRet(&PVIsAlreadyUsedByAnotherPod{mountPoint, slinks}, "More then 1 soft link was pointing to mountpoint.") 
	}
	
	slink := slinks[0]
	if slink != k8sMountPoint{
		return logger.ErrorRet(&PVIsAlreadyUsedByAnotherPod{mountPoint, slinks}, "PV is used by another pod.") 
	}
	
	return nil
}


func (c *Controller) doMount(mountRequest k8sresources.FlexVolumeMountRequest) (string, error) {
	defer c.logger.Trace(logs.DEBUG)()

	// Support only >=1.6
	if mountRequest.Version == k8sresources.KubernetesVersion_1_5 {
		return "", c.logger.ErrorRet(&k8sVersionNotSupported{mountRequest.Version}, "failed")
	}

	mounter, volumeBackend, err := c.getMounterByPV(mountRequest)
	if err != nil {
		return "", c.logger.ErrorRet(err, "getMounterByPV failed")
	}

	ubMountRequest, err := c.prepareUbiquityMountRequest(mountRequest, volumeBackend)
	if err != nil {
		return "", c.logger.ErrorRet(err, "prepareUbiquityMountRequest failed")
	}

	err = checkSlinkAlreadyExistsOnMountPoint(ubMountRequest.Mountpoint, mountRequest.MountPath, c.logger, c.exec)
	if err != nil {
		return "", c.logger.ErrorRet(err, "Failed to check if other links point to mountpoint")
	}

	mountpoint, err := mounter.Mount(ubMountRequest)
	if err != nil {
		return "", c.logger.ErrorRet(err, "mounter.Mount failed")
	}

	return mountpoint, nil
}

func (c *Controller) doAfterMount(mountRequest k8sresources.FlexVolumeMountRequest, mountedPath string) error {
	/*
		 Create symbolic link instead of the k8s PV directory that will point to the ubiquity mountpoint.
		 For example(SCBE backend):
		 	k8s PV directory : /var/lib/kubelet/pods/a9671a20-0fd6-11e8-b968-005056a41609/volumes/ibm~ubiquity-k8s-flex/pvc-6811c716-0f43-11e8-b968-005056a41609
		 	symbolic link should be : /ubiquity/<WWN>
		 Idempotent:
		 	1. if k8s PV dir not exist, then just create the slink.
		 	2. if k8s PV dir exist, then delete the dir and create slink instead.
		 	3. if k8s PV dir is already slink to the right location, skip.
		 	4. if k8s PV dir is already slink to wrong location, raise error.
		 	5. else raise error.
		 Params:
		 	mountedPath : the real mountpoint (e.g: scbe backend its /ubiqutiy/<WWN>)
		 	mountRequest.MountPath : the PV k8s directory
		 k8s version support:
		  	k8s version < 1.6 not supported. Note in version <1.6 the attach/mount,
		  		the mountDir (MountPath) is not created trying to do mount and ln will fail because the dir is not found,
		  		so we need to create the directory before continuing)
			k8s version >= 1.6 supported. Note in version >=1.6 the kubelet creates a folder as the MountPath,
				including the volume name, whenwe try to create the symlink this will fail because the same name exists.
				This is why we need to remove it before continuing.

	*/

	defer c.logger.Trace(logs.DEBUG)()
	var k8sPVDirectoryPath string
	var err error

	k8sPVDirectoryPath = mountRequest.MountPath

	// Identify the PV directory by using Lstat and then handle all idempotent cases (use Lstat to get the dir or slink detail and not the evaluation of it)
	fileInfo, err := c.exec.Lstat(k8sPVDirectoryPath)
	if err != nil {
		if c.exec.IsNotExist(err) {
			// The k8s PV directory not exist (its a rare case and indicate on idempotent flow)
			c.logger.Info("PV directory(k8s-mountpoint) nor slink do not exist. Idempotent - skip delete PV directory(k8s-mountpoint).", logs.Args{{"k8s-mountpoint", k8sPVDirectoryPath}, {"should-point-to-mountpoint", mountedPath}})
			c.logger.Info("Creating slink(k8s-mountpoint) that point to mountpoint", logs.Args{{"k8s-mountpoint", k8sPVDirectoryPath}, {"mountpoint", mountedPath}})
			err = c.exec.Symlink(mountedPath, k8sPVDirectoryPath)
			if err != nil {
				return c.logger.ErrorRet(err, "Controller: failed to create symlink")
			}
		} else {
			// Maybe some permissions issue
			return c.logger.ErrorRet(err, "Controller: failed to identify PV directory(k8s-mountpoint)", logs.Args{{"k8s-mountpoint", k8sPVDirectoryPath}})
		}
	} else if c.exec.IsDir(fileInfo) {
		// Positive flow - the k8s-mountpoint should exist in advance and we should delete it in order to create slink instead
		c.logger.Debug("As expected the PV directory(k8s-mountpoint) is a directory, so remove it to prepare slink to mountpoint instead", logs.Args{{"k8s-mountpath", k8sPVDirectoryPath}, {"mountpoint", mountedPath}})
		err = c.exec.Remove(k8sPVDirectoryPath)
		if err != nil {
			return c.logger.ErrorRet(
				&FailRemovePVorigDirError{k8sPVDirectoryPath, err},
				"failed")
		}
		c.logger.Info("Creating slink(k8s-mountpoint) that point to mountpoint", logs.Args{{"k8s-mountpoint", k8sPVDirectoryPath}, {"mountpoint", mountedPath}})
		err = c.exec.Symlink(mountedPath, k8sPVDirectoryPath)
		if err != nil {
			return c.logger.ErrorRet(err, "Controller: failed to create symlink")
		}
	} else if c.exec.IsSlink(fileInfo) {
		// Its already slink so check if slink is ok and skip else raise error
		evalSlink, err := c.exec.EvalSymlinks(k8sPVDirectoryPath)
		if err != nil {
			return c.logger.ErrorRet(err, "Controller: Idempotent - failed eval the slink of PV directory(k8s-mountpoint)", logs.Args{{"k8s-mountpoint", k8sPVDirectoryPath}})
		}
		if evalSlink == mountedPath {
			c.logger.Info("PV directory(k8s-mountpoint) is already slink and point to the right mountpoint. Idempotent - skip slink creation.", logs.Args{{"k8s-mountpoint", k8sPVDirectoryPath}, {"mountpoint", mountedPath}})
		} else {
			return c.logger.ErrorRet(
				&wrongSlinkError{slink: k8sPVDirectoryPath, wrongPointTo: evalSlink, expectedPointTo: mountedPath},
				"failed")
		}
	} else {
		return c.logger.ErrorRet(&k8sPVDirectoryIsNotDirNorSlinkError{k8sPVDirectoryPath, fileInfo}, "failed")
	}

	c.logger.Debug("Volume mounted successfully", logs.Args{{"mountedPath", mountedPath}})
	return nil
}

func (c *Controller) doAfterDetach(detachRequest k8sresources.FlexVolumeDetachRequest) error {
	defer c.logger.Trace(logs.DEBUG)()

	getVolumeRequest := resources.GetVolumeRequest{Name: detachRequest.Name, Context: detachRequest.Context}
	volume, err := c.Client.GetVolume(getVolumeRequest)
	mounter, err := c.getMounterForBackend(volume.Backend, detachRequest.Context)
	if err != nil {
		err = fmt.Errorf("Error determining mounter for volume: %s", err.Error())
		return c.logger.ErrorRet(err, "failed")
	}

	getVolumeConfigRequest := resources.GetVolumeConfigRequest{Name: detachRequest.Name, Context: detachRequest.Context}
	volumeConfig, err := c.Client.GetVolumeConfig(getVolumeConfigRequest)
	if err != nil {
		err = fmt.Errorf("Error for volume: %s", err.Error())
		return c.logger.ErrorRet(err, "Client.GetVolumeConfig failed")
	}

	afterDetachRequest := resources.AfterDetachRequest{VolumeConfig: volumeConfig, Context: detachRequest.Context}
	if err := mounter.ActionAfterDetach(afterDetachRequest); err != nil {
		err = fmt.Errorf("Error execute action after detaching the volume : %#v", err)
		return c.logger.ErrorRet(err, "mounter.ActionAfterDetach failed")
	}

	return nil
}

func (c *Controller) doUnmountSsc(unmountRequest k8sresources.FlexVolumeUnmountRequest, realMountPoint string) error {
	defer c.logger.Trace(logs.DEBUG)()
	// TODO : double check why for SScale the function trigger detach instead of umount? in addition its bad practice to get all vols.
	//        Consider to delete this function since there is no need for special flow for UnMount for SSc.
	listVolumeRequest := resources.ListVolumesRequest{}
	volumes, err := c.Client.ListVolumes(listVolumeRequest)
	if err != nil {
		err = fmt.Errorf("Error getting the volume list from ubiquity server %#v", err)
		return c.logger.ErrorRet(err, "failed")
	}

	volume, err := getVolumeForMountpoint(unmountRequest.MountPath, volumes)
	if err != nil {
		err = fmt.Errorf(
			"Error finding the volume with mountpoint [%s] from the list of ubiquity volumes %#v. Error is : %#v",
			unmountRequest.MountPath,
			volumes,
			err)
		return c.logger.ErrorRet(err, "failed")
	}

	detachRequest := resources.DetachRequest{Name: volume.Name}
	err = c.Client.Detach(detachRequest)
	if err != nil && err.Error() != "fileset not linked" {
		err = fmt.Errorf(
			"Failed to unmount volume [%s] on mountpoint [%s]. Error: %#v",
			volume.Name,
			unmountRequest.MountPath,
			err)
		return c.logger.ErrorRet(err, "failed")
	}

	return nil
}

func (c *Controller) doActivate(activateRequest resources.ActivateRequest) error {
	defer c.logger.Trace(logs.DEBUG)()

	err := c.Client.Activate(activateRequest)
	if err != nil {
		return c.logger.ErrorRet(err, "Client.Activate failed")
	}

	return nil
}

func (c *Controller) doAttach(attachRequest k8sresources.FlexVolumeAttachRequest) error {
	defer c.logger.Trace(logs.DEBUG)()

	// Support only >=1.6
	if attachRequest.Version == k8sresources.KubernetesVersion_1_5 {
		return c.logger.ErrorRet(&k8sVersionNotSupported{attachRequest.Version}, "failed")
	}

	ubAttachRequest := resources.AttachRequest{Name: attachRequest.Name, Host: getHost(attachRequest.Host), Context: attachRequest.Context}
	_, err := c.Client.Attach(ubAttachRequest)
	if err != nil {
		return c.logger.ErrorRet(err, "Client.Attach failed")
	}

	return nil
}

func (c *Controller) doDetach(detachRequest k8sresources.FlexVolumeDetachRequest, checkIfAttached bool) error {
	defer c.logger.Trace(logs.DEBUG)()

	if checkIfAttached {
		opts := make(map[string]string)
		opts["volumeName"] = detachRequest.Name
		isAttachedRequest := k8sresources.FlexVolumeIsAttachedRequest{Name: "", Host: detachRequest.Host, Opts: opts, Context: detachRequest.Context}
		isAttached, err := c.doIsAttached(isAttachedRequest)
		if err != nil {
			return c.logger.ErrorRet(err, "failed")
		}
		if !isAttached {
			c.logger.Debug(fmt.Sprintf("Skip detach [%s] to host [%s] because its already detached.", detachRequest.Name, detachRequest.Host))
			return nil
		}
	}

	host := detachRequest.Host

	if host == "" {
		// only when triggered during unmount
		var err error
		
		// TODO: if the automation will stop using loggin authenticatin and the WWN here is no longer needed we can
		// we can go back to using regulat getHostAttached and remove getHostAttachUsingConfig.
		getVolumeConfigRequest := resources.GetVolumeConfigRequest{Name: detachRequest.Name, Context: detachRequest.Context}
		volumeConfig, err := c.Client.GetVolumeConfig(getVolumeConfigRequest)
		if err != nil {
			return c.logger.ErrorRet(err, "getVolumeConfig failed.", logs.Args{{"vol", detachRequest.Name}})
		}
		
		host, err = c.getHostAttachUsingConfig(volumeConfig)
		if err != nil {
			return c.logger.ErrorRet(err, "getHostAttached failed.")
		}
		
		if host == "" {
			// this means that the host is not attached to anything so no reason to call detach
			c.logger.Warning(fmt.Sprintf("Idempotent issue encountered - vol is not attached to any host. so no detach action is called"), logs.Args{{"vol wwn", volumeConfig["Wwn"]}})
			return nil
		}
	}

	ubDetachRequest := resources.DetachRequest{Name: detachRequest.Name, Host: host, Context: detachRequest.Context}
	err := c.Client.Detach(ubDetachRequest)
	if err != nil {
		return c.logger.ErrorRet(err, "failed")
	}

	return nil
}

func (c *Controller) doIsAttached(isAttachedRequest k8sresources.FlexVolumeIsAttachedRequest) (bool, error) {
	defer c.logger.Trace(logs.DEBUG)()

	volName, ok := isAttachedRequest.Opts["volumeName"]
	if !ok {
		err := fmt.Errorf("volumeName not found in isAttachedRequest")
		return false, c.logger.ErrorRet(err, "failed")
	}

	attachTo, err := c.getHostAttached(volName, isAttachedRequest.Context)
	if err != nil {
		return false, c.logger.ErrorRet(err, "getHostAttached failed")
	}

	isAttached := isAttachedRequest.Host == attachTo
	c.logger.Debug("", logs.Args{{"host", isAttachedRequest.Host}, {"attachTo", attachTo}, {"isAttached", isAttached}})
	return isAttached, nil
}

func (c *Controller) getHostAttachUsingConfig(volumeConfig map[string]interface{}) (string, error){
	attachTo, ok := volumeConfig[resources.ScbeKeyVolAttachToHost].(string)
	if !ok {
		return "", c.logger.ErrorRet(fmt.Errorf("GetVolumeConfig is missing info"), "GetVolumeConfig missing info.", logs.Args{{"arg", resources.ScbeKeyVolAttachToHost}})
	}
	c.logger.Debug("", logs.Args{{"volumeConfig", volumeConfig}, {"attachTo", attachTo}})

	return attachTo, nil
}

func (c *Controller) getHostAttached(volName string, requestContext resources.RequestContext) (string, error) {
	defer c.logger.Trace(logs.DEBUG)()

	getVolumeConfigRequest := resources.GetVolumeConfigRequest{Name: volName, Context: requestContext}
	volumeConfig, err := c.Client.GetVolumeConfig(getVolumeConfigRequest)
	if err != nil {
		return "", c.logger.ErrorRet(err, "Client.GetVolumeConfig failed")
	}

	return c.getHostAttachUsingConfig(volumeConfig)
}

func getVolumeForMountpoint(mountpoint string, volumes []resources.Volume) (resources.Volume, error) {
	for _, volume := range volumes {
		if volume.Mountpoint == mountpoint {
			return volume, nil
		}
	}
	return resources.Volume{}, fmt.Errorf("Volume not found")
}

func getHost(hostRequest string) string {
	if hostRequest != "" {
		return hostRequest
	}
	// Only in k8s 1.5 this os.Hostname will happened,
	// because in k8s 1.5 the flex CLI doesn't get the host to attach with. TODO consider to refactor to remove support for 1.5
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hostname
}

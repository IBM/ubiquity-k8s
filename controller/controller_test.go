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

package controller_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	ctl "github.com/IBM/ubiquity-k8s/controller"
	k8sresources "github.com/IBM/ubiquity-k8s/resources"
	"github.com/IBM/ubiquity/fakes"
	"github.com/IBM/ubiquity/resources"
)

var _ = Describe("Controller", func() {
	Context(".Init", func() {
		var (
			fakeClient     *fakes.FakeStorageClient
			controller     *ctl.Controller
			fakeExec       *fakes.FakeExecutor
			ubiquityConfig resources.UbiquityPluginConfig
		)
		BeforeEach(func() {
			fakeExec = new(fakes.FakeExecutor)
			ubiquityConfig = resources.UbiquityPluginConfig{}
			fakeClient = new(fakes.FakeStorageClient)
			controller = ctl.NewControllerWithClient(testLogger, fakeClient, fakeExec)
		})
		It("does not error when init is successful", func() {
			initResponse := controller.Init(ubiquityConfig)
			Expect(initResponse.Status).To(Equal("Success"))
			Expect(initResponse.Message).To(Equal("Plugin init successfully"))
			Expect(initResponse.Device).To(Equal(""))
		})

		Context(".Attach", func() {

			It("fails when attachRequest does not have volumeName", func() {
				fakeClient.GetVolumeReturns(resources.Volume{}, fmt.Errorf("GetVolume error"))
				attachRequest := map[string]string{"Filesystem": "gpfs1", "Size": "200m", "Fileset": "fs1", "Path": "myPath"}
				attachResponse := controller.Attach(attachRequest)
				Expect(attachResponse.Status).To(Equal("Failure"))
				Expect(fakeClient.GetVolumeCallCount()).To(Equal(0))
			})

			It("fails when client fails to fetch volume", func() {
				fakeClient.GetVolumeReturns(resources.Volume{}, fmt.Errorf("GetVolume error"))
				attachRequest := map[string]string{"volumeName": "vol1", "Filesystem": "gpfs1", "Size": "200m", "Fileset": "fs1", "Path": "myPath"}
				attachResponse := controller.Attach(attachRequest)
				Expect(attachResponse.Status).To(Equal("Failure"))
				Expect(attachResponse.Message).To(Equal("Failed checking volume, call create before attach"))
				Expect(attachResponse.Device).To(Equal("vol1"))
			})

			It("Succeeds when volume exists", func() {
				fakeClient.GetVolumeReturns(resources.Volume{}, nil)
				attachRequest := map[string]string{"volumeName": "vol1", "Filesystem": "gpfs1", "Size": "200m", "Fileset": "fs1", "Path": "myPath"}
				attachResponse := controller.Attach(attachRequest)
				Expect(attachResponse.Status).To(Equal("Success"))
				Expect(attachResponse.Message).To(Equal("Volume already attached"))
				Expect(attachResponse.Device).To(Equal("vol1"))
				Expect(fakeClient.CreateVolumeCallCount()).To(Equal(0))
			})
		})
		//
		//Context(".Detach", func() {
		//	It("does not error when existing volume name is given", func() {
		//		fakeClient.RemoveVolumeReturns(nil)
		//		detachRequest := resources.FlexVolumeDetachRequest{Name: "vol1"}
		//		detachResponse := controller.Detach(detachRequest)
		//		Expect(detachResponse.Status).To(Equal("Success"))
		//		Expect(detachResponse.Message).To(Equal("Volume detached successfully"))
		//		Expect(detachResponse.Device).To(Equal("vol1"))
		//		Expect(fakeClient.RemoveVolumeCallCount()).To(Equal(1))
		//	})
		//
		//	It("error when client fails to detach volume", func() {
		//		err := fmt.Errorf("error detaching volume")
		//		fakeClient.RemoveVolumeReturns(err)
		//		detachRequest := resources.FlexVolumeDetachRequest{Name: "vol1"}
		//		detachResponse := controller.Detach(detachRequest)
		//		Expect(detachResponse.Status).To(Equal("Failure"))
		//		Expect(detachResponse.Message).To(Equal(fmt.Sprintf("Failed to detach volume %#v", err)))
		//		Expect(detachResponse.Device).To(Equal("vol1"))
		//		Expect(fakeClient.RemoveVolumeCallCount()).To(Equal(1))
		//	})
		//})
		Context(".Mount", func() {
			It("does not error when volume exists and is not currently mounted", func() {
				fakeClient.AttachReturns("/tmp/mnt1", nil)
				mountRequest := k8sresources.FlexVolumeMountRequest{MountPath: "/tmp/mnt2", MountDevice: "vol1", Opts: map[string]string{}}
				mountResponse := controller.Mount(mountRequest)
				Expect(mountResponse.Status).To(Equal("Success"))
				Expect(mountResponse.Message).To(Equal("Volume mounted successfully to /tmp/mnt1"))
				Expect(mountResponse.Device).To(Equal(""))
				Expect(fakeClient.AttachCallCount()).To(Equal(1))
			})
			AfterEach(func() {
				err := os.RemoveAll("/tmp/mnt2")
				Expect(err).ToNot(HaveOccurred())
			})
			It("errors when volume exists and client fails to mount it", func() {
				err := fmt.Errorf("failed to mount volume")
				fakeClient.AttachReturns("", err)
				mountRequest := k8sresources.FlexVolumeMountRequest{MountPath: "some-mountpath", MountDevice: "vol1", Opts: map[string]string{}}
				mountResponse := controller.Mount(mountRequest)
				Expect(mountResponse.Status).To(Equal("Failure"))
				Expect(mountResponse.Message).To(MatchRegexp(err.Error()))
				Expect(mountResponse.Device).To(Equal(""))
				Expect(fakeClient.AttachCallCount()).To(Equal(1))
			})
		})
		Context(".Unmount", func() {
			var volumes []resources.Volume
			It("succeeds when volume exists and is currently mounted", func() {
				fakeExec.EvalSymlinksReturns("/path/gpfs/fs/mountpoint", nil)
				fakeClient.DetachReturns(nil)
				volume := resources.Volume{Name: "vol1", Mountpoint: "some-mountpoint"}
				volumes = []resources.Volume{volume}
				fakeClient.ListVolumesReturns(volumes, nil)
				unmountRequest := k8sresources.FlexVolumeUnmountRequest{MountPath: "some-mountpoint"}
				unmountResponse := controller.Unmount(unmountRequest)
				Expect(unmountResponse.Status).To(Equal("Success"))
				Expect(unmountResponse.Message).To(Equal("Volume unmounted successfully"))
				Expect(unmountResponse.Device).To(Equal(""))
				Expect(fakeClient.DetachCallCount()).To(Equal(1))
				Expect(fakeClient.ListVolumesCallCount()).To(Equal(1))
			})
			It("errors when client fails to get volume related to the mountpoint", func() {
				err := fmt.Errorf("failed to get fileset")
				fakeClient.ListVolumesReturns(volumes, err)
				unmountRequest := k8sresources.FlexVolumeUnmountRequest{MountPath: "some-mountpoint"}
				unmountResponse := controller.Unmount(unmountRequest)

				Expect(unmountResponse.Status).To(Equal("Failure"))
				Expect(unmountResponse.Message).To(MatchRegexp(err.Error()))
				Expect(unmountResponse.Device).To(Equal(""))
				Expect(fakeClient.ListVolumesCallCount()).To(Equal(1))
				Expect(fakeClient.DetachCallCount()).To(Equal(0))
			})
			It("errors when volume does not exist", func() {
				volumes = []resources.Volume{}
				fakeClient.ListVolumesReturns(volumes, nil)
				unmountRequest := k8sresources.FlexVolumeUnmountRequest{MountPath: "some-mountpoint"}
				unmountResponse := controller.Unmount(unmountRequest)

				Expect(unmountResponse.Status).To(Equal("Failure"))
				Expect(unmountResponse.Message).To(MatchRegexp("Volume not found"))
				Expect(unmountResponse.Device).To(Equal(""))
				Expect(fakeClient.ListVolumesCallCount()).To(Equal(1))
				Expect(fakeClient.DetachCallCount()).To(Equal(0))
			})
			It("errors when volume exists and client fails to unmount it", func() {
				err := fmt.Errorf("error detaching the volume")
				volume := resources.Volume{Name: "vol1", Mountpoint: "some-mountpoint"}
				volumes = []resources.Volume{volume}
				fakeClient.ListVolumesReturns(volumes, nil)
				fakeClient.DetachReturns(err)
				unmountRequest := k8sresources.FlexVolumeUnmountRequest{MountPath: "some-mountpoint"}
				unmountResponse := controller.Unmount(unmountRequest)

				Expect(unmountResponse.Status).To(Equal("Failure"))
				Expect(unmountResponse.Message).To(MatchRegexp(err.Error()))
				Expect(unmountResponse.Device).To(Equal(""))
				Expect(fakeClient.ListVolumesCallCount()).To(Equal(1))
				Expect(fakeClient.DetachCallCount()).To(Equal(1))

			})
			It("should fail to umount if mountpoint is not slink", func() {
				errMsg := fmt.Errorf("not a link")
				fakeExec.EvalSymlinksReturns("", errMsg)

				unmountRequest := k8sresources.FlexVolumeUnmountRequest{MountPath: "some-mountpoint"}
				unmountResponse := controller.Unmount(unmountRequest)

				Expect(unmountResponse.Status).To(Equal("Failure"))
				Expect(unmountResponse.Message).To(MatchRegexp(errMsg.Error()))
				Expect(unmountResponse.Device).To(Equal(""))
			})
			It("should fail to umount if detach failed", func() {
				errMsg := fmt.Errorf("error")
				realMountPoint := fmt.Sprintf(resources.PathToMountUbiquityBlockDevices, "fakeWWN")
				fakeExec.EvalSymlinksReturns(realMountPoint, nil)
				fakeClient.DetachReturns(errMsg)

				unmountRequest := k8sresources.FlexVolumeUnmountRequest{MountPath: "/k8s/podid/some/pvname"}
				unmountResponse := controller.Unmount(unmountRequest)
				Expect(unmountResponse.Status).To(Equal("Failure"))
				Expect(unmountResponse.Message).To(MatchRegexp(errMsg.Error()))
				Expect(unmountResponse.Device).To(Equal(""))
				detachRequest := fakeClient.DetachArgsForCall(0)
				Expect(detachRequest.Name).To(Equal("pvname"))
			})
			It("should fail to umount if detach failed", func() {
				errMsg := fmt.Errorf("error")
				realMountPoint := fmt.Sprintf(resources.PathToMountUbiquityBlockDevices, "fakeWWN")
				fakeExec.EvalSymlinksReturns(realMountPoint, nil)
				fakeClient.DetachReturns(errMsg)

				unmountRequest := k8sresources.FlexVolumeUnmountRequest{MountPath: "/k8s/podid/some/pvname"}
				unmountResponse := controller.Unmount(unmountRequest)
				Expect(unmountResponse.Status).To(Equal("Failure"))
				Expect(unmountResponse.Message).To(MatchRegexp(errMsg.Error()))
				Expect(unmountResponse.Device).To(Equal(""))
				detachRequest := fakeClient.DetachArgsForCall(0)
				Expect(detachRequest.Name).To(Equal("pvname"))
			})

			It("should fail to umount if fail to remove the slink", func() {
				errMsg := fmt.Errorf("error")
				realMountPoint := fmt.Sprintf(resources.PathToMountUbiquityBlockDevices, "fakeWWN")
				fakeExec.EvalSymlinksReturns(realMountPoint, nil)
				fakeClient.DetachReturns(nil)
				fakeExec.RemoveReturns(errMsg)

				unmountRequest := k8sresources.FlexVolumeUnmountRequest{MountPath: "/k8s/podid/some/pvname"}
				unmountResponse := controller.Unmount(unmountRequest)
				Expect(unmountResponse.Status).To(Equal("Failure"))
				Expect(unmountResponse.Message).To(MatchRegexp(errMsg.Error()))
				Expect(unmountResponse.Device).To(Equal(""))
				detachRequest := fakeClient.DetachArgsForCall(0)
				Expect(detachRequest.Name).To(Equal("pvname"))
				Expect(fakeExec.RemoveCallCount()).To(Equal(1))
			})
			It("should succeed to umount if the scbe umount flow finished ok", func() {
				realMountPoint := fmt.Sprintf(resources.PathToMountUbiquityBlockDevices, "fakeWWN")
				fakeExec.EvalSymlinksReturns(realMountPoint, nil)
				fakeClient.DetachReturns(nil)
				fakeExec.RemoveReturns(nil)

				unmountRequest := k8sresources.FlexVolumeUnmountRequest{MountPath: "/k8s/podid/some/pvname"}
				unmountResponse := controller.Unmount(unmountRequest)
				Expect(unmountResponse.Status).To(Equal("Success"))
				Expect(unmountResponse.Message).To(Equal("Volume unmounted successfully"))
				Expect(unmountResponse.Device).To(Equal(""))
				detachRequest := fakeClient.DetachArgsForCall(0)
				Expect(detachRequest.Name).To(Equal("pvname"))
				Expect(fakeExec.RemoveCallCount()).To(Equal(1))
			})

		})
	})
})

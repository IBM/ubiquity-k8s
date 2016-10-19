package core_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.ibm.com/almaden-containers/ubiquity-flexvolume.git/core"
	"github.ibm.com/almaden-containers/ubiquity.git/fakes"
	"github.ibm.com/almaden-containers/ubiquity.git/model"
)

var _ = Describe("Controller", func() {
	Context(".Init", func() {
		var (
			fakeClient *fakes.FakeStorageClient
			controller *core.Controller
		)
		BeforeEach(func() {
			fakeClient = new(fakes.FakeStorageClient)
			controller = core.NewControllerWithClient(testLogger, fakeClient)
		})
		It("does not error when init is successful", func() {
			initResponse := controller.Init()
			Expect(initResponse.Status).To(Equal("Success"))
			Expect(initResponse.Message).To(Equal("Plugin init successfully"))
			Expect(initResponse.Device).To(Equal(""))
		})

		Context(".Attach", func() {
			It("does not error on create with valid opts", func() {
				fakeClient.CreateVolumeReturns(nil)
				attachRequest := model.FlexVolumeAttachRequest{VolumeId: "vol1", Filesystem: "gpfs1", Size: "200m", Fileset: "fs1", Path: "myPath"}
				attachResponse := controller.Attach(attachRequest)
				Expect(attachResponse.Status).To(Equal("Success"))
				Expect(attachResponse.Message).To(Equal("Volume attached successfully"))
				Expect(attachResponse.Device).To(Equal("vol1"))
				Expect(fakeClient.CreateVolumeCallCount()).To(Equal(1))
			})
			It("does error on create when client fails to attach", func() {
				err := fmt.Errorf("Spectrum internal error on attach")
				fakeClient.CreateVolumeReturns(err)
				attachRequest := model.FlexVolumeAttachRequest{VolumeId: "vol", Size: "200m", Filesystem: "gpfs1", Fileset: "fs1", Path: "myPath"}
				attachResponse := controller.Attach(attachRequest)
				Expect(attachResponse.Status).To(Equal("Failure"))
				Expect(attachResponse.Message).To(Equal(fmt.Sprintf("Failed to attach volume: %#v", err)))
				Expect(attachResponse.Device).To(Equal("vol"))
				Expect(fakeClient.CreateVolumeCallCount()).To(Equal(1))
			})

		})

		Context(".Detach", func() {
			It("does not error when existing volume name is given", func() {
				fakeClient.RemoveVolumeReturns(nil)
				detachRequest := model.FlexVolumeDetachRequest{Name: "vol1"}
				detachResponse := controller.Detach(detachRequest)
				Expect(detachResponse.Status).To(Equal("Success"))
				Expect(detachResponse.Message).To(Equal("Volume detached successfully"))
				Expect(detachResponse.Device).To(Equal("vol1"))
				Expect(fakeClient.RemoveVolumeCallCount()).To(Equal(1))
			})

			It("error when client fails to detach volume", func() {
				err := fmt.Errorf("error detaching volume")
				fakeClient.RemoveVolumeReturns(err)
				detachRequest := model.FlexVolumeDetachRequest{Name: "vol1"}
				detachResponse := controller.Detach(detachRequest)
				Expect(detachResponse.Status).To(Equal("Failure"))
				Expect(detachResponse.Message).To(Equal(fmt.Sprintf("Failed to detach volume %#v", err)))
				Expect(detachResponse.Device).To(Equal("vol1"))
				Expect(fakeClient.RemoveVolumeCallCount()).To(Equal(1))
			})
		})
		Context(".Mount", func() {
			It("does not error when volume exists and is not currently mounted", func() {
				fakeClient.AttachReturns("/tmp/mnt1", nil)
				mountRequest := model.FlexVolumeMountRequest{MountPath: "/tmp/mnt2", MountDevice: "vol1", Opts: map[string]interface{}{}}
				mountResponse := controller.Mount(mountRequest)
				Expect(mountResponse.Status).To(Equal("Success"))
				Expect(mountResponse.Message).To(Equal("Volume mounted successfully to /tmp/mnt1"))
				Expect(mountResponse.Device).To(Equal(""))
				Expect(fakeClient.AttachCallCount()).To(Equal(1))
			})
			AfterEach(func() {
				os.RemoveAll("/tmp/mnt2")
			})
			It("errors when volume exists and client fails to mount it", func() {
				err := fmt.Errorf("failed to mount volume")
				fakeClient.AttachReturns("", err)
				mountRequest := model.FlexVolumeMountRequest{MountPath: "some-mountpath", MountDevice: "vol1", Opts: map[string]interface{}{}}
				mountResponse := controller.Mount(mountRequest)
				Expect(mountResponse.Status).To(Equal("Failure"))
				Expect(mountResponse.Message).To(Equal(fmt.Sprintf("Failed to mount volume %#v", err)))
				Expect(mountResponse.Device).To(Equal(""))
				Expect(fakeClient.AttachCallCount()).To(Equal(1))
			})
		})
		Context(".Unmount", func() {
			var volumes []model.VolumeMetadata
			It("succeeds when volume exists and is currently mounted", func() {
				fakeClient.DetachReturns(nil)
				volume := model.VolumeMetadata{Name: "vol1", Mountpoint: "some-mountpoint"}
				volumes = []model.VolumeMetadata{volume}
				fakeClient.ListVolumesReturns(volumes, nil)
				unmountRequest := model.FlexVolumeUnmountRequest{MountPath: "some-mountpoint"}
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
				unmountRequest := model.FlexVolumeUnmountRequest{MountPath: "some-mountpoint"}
				unmountResponse := controller.Unmount(unmountRequest)

				Expect(unmountResponse.Status).To(Equal("Failure"))
				Expect(unmountResponse.Message).To(Equal(fmt.Sprintf("Error finding the volume %#v", err)))
				Expect(unmountResponse.Device).To(Equal(""))
				Expect(fakeClient.ListVolumesCallCount()).To(Equal(1))
				Expect(fakeClient.DetachCallCount()).To(Equal(0))
			})
			It("errors when volume does not exist", func() {
				volumes = []model.VolumeMetadata{}
				fakeClient.ListVolumesReturns(volumes, nil)
				unmountRequest := model.FlexVolumeUnmountRequest{MountPath: "some-mountpoint"}
				unmountResponse := controller.Unmount(unmountRequest)

				Expect(unmountResponse.Status).To(Equal("Failure"))
				Expect(unmountResponse.Message).To(Equal("Error finding the volume &errors.errorString{s:\"Volume not found\"}"))
				Expect(unmountResponse.Device).To(Equal(""))
				Expect(fakeClient.ListVolumesCallCount()).To(Equal(1))
				Expect(fakeClient.DetachCallCount()).To(Equal(0))
			})
			It("errors when volume exists and client fails to unmount it", func() {
				err := fmt.Errorf("error detaching the volume")
				volume := model.VolumeMetadata{Name: "vol1", Mountpoint: "some-mountpoint"}
				volumes = []model.VolumeMetadata{volume}
				fakeClient.ListVolumesReturns(volumes, nil)
				fakeClient.DetachReturns(err)
				unmountRequest := model.FlexVolumeUnmountRequest{MountPath: "some-mountpoint"}
				unmountResponse := controller.Unmount(unmountRequest)

				Expect(unmountResponse.Status).To(Equal("Failure"))
				Expect(unmountResponse.Message).To(Equal(fmt.Sprintf("Failed to unmount volume %#v", err)))
				Expect(unmountResponse.Device).To(Equal(""))
				Expect(fakeClient.ListVolumesCallCount()).To(Equal(1))
				Expect(fakeClient.DetachCallCount()).To(Equal(1))

			})
		})
	})
})

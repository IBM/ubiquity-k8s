package core_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.ibm.com/almaden-containers/spectrum-common.git/fakes"
	"github.ibm.com/almaden-containers/spectrum-common.git/models"
	"github.ibm.com/almaden-containers/spectrum-flexvolume-cli.git/core"
)

var _ = Describe("Controller", func() {
	Context(".Init", func() {
		var (
			fakeClient *fakes.FakeSpectrumClient
			controller *core.Controller
		)
		BeforeEach(func() {
			fakeClient = new(fakes.FakeSpectrumClient)
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
				fakeClient.CreateWithoutProvisioningReturns(nil)
				attachRequest := &models.FlexVolumeAttachRequest{VolumeId: "vol1", Filesystem: "gpfs1", Size: "200m", Fileset: "fs1", Path: "myPath"}
				attachResponse := controller.Attach(attachRequest)
				Expect(attachResponse.Status).To(Equal("Success"))
				Expect(attachResponse.Message).To(Equal("Volume attached successfully"))
				Expect(attachResponse.Device).To(Equal("vol1"))
				Expect(fakeClient.CreateWithoutProvisioningCallCount()).To(Equal(1))
			})
			It("does error on create when client fails to attach", func() {
				err := fmt.Errorf("Spectrum internal error on attach")
				fakeClient.CreateWithoutProvisioningReturns(err)
				attachRequest := &models.FlexVolumeAttachRequest{VolumeId: "vol", Size: "200m", Filesystem: "gpfs1", Fileset: "fs1", Path: "myPath"}
				attachResponse := controller.Attach(attachRequest)
				Expect(attachResponse.Status).To(Equal("Failure"))
				Expect(attachResponse.Message).To(Equal(fmt.Sprintf("Failed to attach volume: %#v", err)))
				Expect(attachResponse.Device).To(Equal("vol"))
				Expect(fakeClient.CreateWithoutProvisioningCallCount()).To(Equal(1))
			})

		})

		Context(".Detach", func() {
			It("does not error when existing volume name is given", func() {
				volume := &models.VolumeMetadata{Name: "vol1"}
				fakeClient.GetReturns(volume, nil, nil)
				fakeClient.RemoveWithoutDeletingVolumeReturns(nil)
				detachRequest := &models.GenericRequest{Name: "vol1"}
				detachResponse := controller.Detach(detachRequest)
				Expect(detachResponse.Status).To(Equal("Success"))
				Expect(detachResponse.Message).To(Equal("Volume detached successfully"))
				Expect(detachResponse.Device).To(Equal("vol1"))
				Expect(fakeClient.GetCallCount()).To(Equal(1))
				Expect(fakeClient.RemoveWithoutDeletingVolumeCallCount()).To(Equal(1))
			})
			It("error when volume not found", func() {
				fakeClient.GetReturns(nil, nil, nil)
				detachRequest := &models.GenericRequest{Name: "vol1"}
				detachResponse := controller.Detach(detachRequest)
				Expect(detachResponse.Status).To(Equal("Failure"))
				Expect(detachResponse.Message).To(Equal("Volume not found"))
				Expect(detachResponse.Device).To(Equal("vol1"))
				Expect(fakeClient.GetCallCount()).To(Equal(1))
				Expect(fakeClient.RemoveWithoutDeletingVolumeCallCount()).To(Equal(0))
			})

			It("error when client fails to retrieve volume info", func() {
				err := fmt.Errorf("Client error")
				fakeClient.GetReturns(nil, nil, err)
				detachRequest := &models.GenericRequest{Name: "vol1"}
				detachResponse := controller.Detach(detachRequest)
				Expect(detachResponse.Status).To(Equal("Failure"))
				Expect(detachResponse.Message).To(Equal(fmt.Sprintf("Failed to detach volume %#v", err)))
				Expect(detachResponse.Device).To(Equal("vol1"))
				Expect(fakeClient.GetCallCount()).To(Equal(1))
				Expect(fakeClient.RemoveCallCount()).To(Equal(0))
			})

			It("error when client fails to detach volume", func() {
				err := fmt.Errorf("error detaching volume")
				volume := &models.VolumeMetadata{Name: "vol1"}
				fakeClient.GetReturns(volume, nil, nil)
				fakeClient.RemoveWithoutDeletingVolumeReturns(err)
				detachRequest := &models.GenericRequest{Name: "vol1"}
				detachResponse := controller.Detach(detachRequest)
				Expect(detachResponse.Status).To(Equal("Failure"))
				Expect(detachResponse.Message).To(Equal(fmt.Sprintf("Failed to detach volume %#v", err)))
				Expect(detachResponse.Device).To(Equal("vol1"))
				Expect(fakeClient.GetCallCount()).To(Equal(1))
				Expect(fakeClient.RemoveWithoutDeletingVolumeCallCount()).To(Equal(1))
			})
		})
		Context(".Mount", func() {
			It("does not error when volume exists and is not currently mounted", func() {
				volume := &models.VolumeMetadata{Name: "vol1"}
				fakeClient.GetReturns(volume, nil, nil)
				fakeClient.AttachReturns("/tmp/mnt1", nil)
				mountRequest := &models.FlexVolumeMountRequest{MountPath: "/tmp/mnt2", MountDevice: "vol1", Opts: map[string]interface{}{}}
				mountResponse := controller.Mount(mountRequest)
				Expect(mountResponse.Status).To(Equal("Success"))
				Expect(mountResponse.Message).To(Equal("Volume mounted successfully to /tmp/mnt1"))
				Expect(mountResponse.Device).To(Equal(""))
				Expect(fakeClient.GetCallCount()).To(Equal(1))
				Expect(fakeClient.AttachCallCount()).To(Equal(1))
			})
			AfterEach(func() {
				os.RemoveAll("/tmp/mnt2")
			})
			It("errors when volume get returns error", func() {
				err := fmt.Errorf("error listing volume")
				fakeClient.GetReturns(nil, nil, err)
				mountRequest := &models.FlexVolumeMountRequest{MountPath: "some-mountpath", MountDevice: "vol1", Opts: map[string]interface{}{}}
				mountResponse := controller.Mount(mountRequest)
				Expect(mountResponse.Status).To(Equal("Failure"))
				Expect(mountResponse.Message).To(Equal(fmt.Sprintf("Failed to mount volume %#v", err)))
				Expect(mountResponse.Device).To(Equal(""))
				Expect(fakeClient.GetCallCount()).To(Equal(1))
				Expect(fakeClient.AttachCallCount()).To(Equal(0))
			})
			It("errors when volume does not exist", func() {
				fakeClient.GetReturns(nil, nil, nil)
				mountRequest := &models.FlexVolumeMountRequest{MountPath: "some-mountpath", MountDevice: "vol1", Opts: map[string]interface{}{}}
				mountResponse := controller.Mount(mountRequest)
				Expect(mountResponse.Status).To(Equal("Failure"))
				Expect(mountResponse.Message).To(Equal("Failed to mount volume: volume not found"))
				Expect(mountResponse.Device).To(Equal(""))
				Expect(fakeClient.GetCallCount()).To(Equal(1))
				Expect(fakeClient.AttachCallCount()).To(Equal(0))
			})
			It("errors when volume exists and client fails to mount it", func() {
				err := fmt.Errorf("failed to mount volume")
				volume := &models.VolumeMetadata{Name: "vol1"}
				fakeClient.GetReturns(volume, nil, nil)
				fakeClient.AttachReturns("", err)
				mountRequest := &models.FlexVolumeMountRequest{MountPath: "some-mountpath", MountDevice: "vol1", Opts: map[string]interface{}{}}
				mountResponse := controller.Mount(mountRequest)
				Expect(mountResponse.Status).To(Equal("Failure"))
				Expect(mountResponse.Message).To(Equal(fmt.Sprintf("Failed to mount volume %#v", err)))
				Expect(mountResponse.Device).To(Equal(""))
				Expect(fakeClient.GetCallCount()).To(Equal(1))
				Expect(fakeClient.AttachCallCount()).To(Equal(1))
			})
		})
		Context(".Unmount", func() {
			It("succeeds when volume exists and is currently mounted", func() {
				fakeClient.GetFileSetForMountPointReturns("vol1", nil)
				fakeClient.DetachReturns(nil)
				unmountRequest := &models.GenericRequest{Name: "some-mountpoint"}
				unmountResponse := controller.Unmount(unmountRequest)
				Expect(unmountResponse.Status).To(Equal("Success"))
				Expect(unmountResponse.Message).To(Equal("Volume unmounted successfully"))
				Expect(unmountResponse.Device).To(Equal(""))
				Expect(fakeClient.GetFileSetForMountPointCallCount()).To(Equal(1))
				Expect(fakeClient.DetachCallCount()).To(Equal(1))
			})
			It("errors when client fails to get volume related to the mountpoint", func() {
				err := fmt.Errorf("failed to get fileset")
				fakeClient.GetFileSetForMountPointReturns("", err)
				unmountRequest := &models.GenericRequest{Name: "some-mountpoint"}
				unmountResponse := controller.Unmount(unmountRequest)

				Expect(unmountResponse.Status).To(Equal("Failure"))
				Expect(unmountResponse.Message).To(Equal(fmt.Sprintf("Error finding the volume %#v", err)))
				Expect(unmountResponse.Device).To(Equal(""))
				Expect(fakeClient.GetFileSetForMountPointCallCount()).To(Equal(1))
				Expect(fakeClient.DetachCallCount()).To(Equal(0))
			})
			It("errors when volume does not exist", func() {
				fakeClient.GetFileSetForMountPointReturns("", nil)
				unmountRequest := &models.GenericRequest{Name: "some-mountpoint"}
				unmountResponse := controller.Unmount(unmountRequest)

				Expect(unmountResponse.Status).To(Equal("Failure"))
				Expect(unmountResponse.Message).To(Equal("Volume not found"))
				Expect(unmountResponse.Device).To(Equal(""))
				Expect(fakeClient.GetFileSetForMountPointCallCount()).To(Equal(1))
				Expect(fakeClient.DetachCallCount()).To(Equal(0))
			})
			It("errors when volume exists and client fails to unmount it", func() {
				err := fmt.Errorf("error detaching the volume")
				fakeClient.GetFileSetForMountPointReturns("vol1", nil)
				fakeClient.DetachReturns(err)
				unmountRequest := &models.GenericRequest{Name: "some-mountpoint"}
				unmountResponse := controller.Unmount(unmountRequest)

				Expect(unmountResponse.Status).To(Equal("Failure"))
				Expect(unmountResponse.Message).To(Equal(fmt.Sprintf("Failed to unmount volume %#v", err)))
				Expect(unmountResponse.Device).To(Equal(""))
				Expect(fakeClient.GetFileSetForMountPointCallCount()).To(Equal(1))
				Expect(fakeClient.DetachCallCount()).To(Equal(1))

			})
		})
	})
})

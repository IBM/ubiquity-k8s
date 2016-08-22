package core_test

// import (
// . "github.com/onsi/ginkgo"
// . "github.com/onsi/gomega"
// )

// var _ = Describe("Controller", func() {
// 	Context("on activate", func() {
// 		var (
// 			fakeClient *fakes.FakeSpectrumClient
// 			controller *core.Controller
// 		)
// 		BeforeEach(func() {
// 			fakeClient = new(fakes.FakeSpectrumClient)
// 			controller = core.NewControllerWithClient(testLogger, fakeClient)
// 		})
// 		It("does not error when mount is successful", func() {
// 			activateResponse := controller.Activate()
// 			Expect(activateResponse.Implements).ToNot(Equal(nil))
// 			Expect(len(activateResponse.Implements)).To(Equal(1))
// 			Expect(activateResponse.Implements[0]).To(Equal("VolumeDriver"))
// 			Expect(fakeClient.MountCallCount()).To(Equal(1))
// 		})
// 		It("does not error when previously mounted", func() {
// 			fakeClient.IsMountedReturns(true, nil)
// 			activateResponse := controller.Activate()
// 			Expect(activateResponse.Implements).ToNot(Equal(nil))
// 			Expect(len(activateResponse.Implements)).To(Equal(1))
// 			Expect(activateResponse.Implements[0]).To(Equal("VolumeDriver"))
// 			Expect(fakeClient.MountCallCount()).To(Equal(0))
// 		})
// 		It("errors when mount fails", func() {
// 			fakeClient.MountReturns(fmt.Errorf("Failed to mount"))
// 			activateResponse := controller.Activate()
// 			Expect(activateResponse.Implements).ToNot(Equal(nil))
// 			Expect(len(activateResponse.Implements)).To(Equal(0))
// 		})
// 		It("errors when isMounted returns error", func() {
// 			fakeClient.IsMountedReturns(false, fmt.Errorf("checking if mounted failed"))
// 			activateResponse := controller.Activate()
// 			Expect(activateResponse.Implements).ToNot(Equal(nil))
// 			Expect(len(activateResponse.Implements)).To(Equal(0))
// 		})

// 		Context("on successful activate", func() {
// 			BeforeEach(func() {
// 				activateResponse := controller.Activate()
// 				Expect(activateResponse.Implements).ToNot(Equal(nil))
// 				Expect(len(activateResponse.Implements)).To(Equal(1))
// 				Expect(activateResponse.Implements[0]).To(Equal("VolumeDriver"))
// 			})
// 			Context(".Create", func() {
// 				It("does not error on create with valid opts", func() {
// 					fakeClient.CreateReturns(nil)
// 					createRequest := &models.CreateRequest{Name: "dockerVolume1", Opts: map[string]interface{}{"Filesystem": "gpfs1"}}
// 					createResponse := controller.Create(createRequest)
// 					Expect(createResponse.Err).To(Equal(""))
// 					Expect(fakeClient.CreateCallCount()).To(Equal(1))
// 					name, _ := fakeClient.CreateArgsForCall(0)
// 					Expect(name).To(Equal("dockerVolume1"))
// 				})
// 				It("errors on create with valid opts if dockerVolume already exists", func() {
// 					dockerVolume := models.VolumeMetadata{Name: "dockerVolume1"}
// 					fakeClient.GetReturns(&dockerVolume, nil, nil)
// 					createRequest := &models.CreateRequest{Name: "dockerVolume1", Opts: map[string]interface{}{"Filesystem": "gpfs1"}}
// 					createResponse := controller.Create(createRequest)
// 					Expect(createResponse.Err).To(Equal("Volume already exists"))
// 				})
// 				It("does error on create when plugin fails to create dockerVolume", func() {
// 					fakeClient.CreateReturns(fmt.Errorf("Spectrum plugin internal error"))
// 					createRequest := &models.CreateRequest{Name: "dockerVolume1", Opts: map[string]interface{}{"Filesystem": "gpfs1"}}
// 					createResponse := controller.Create(createRequest)
// 					Expect(createResponse.Err).To(Equal("Spectrum plugin internal error"))
// 				})
// 				It("does error on create when plugin fails to list existing dockerVolume", func() {
// 					fakeClient.GetReturns(nil, nil, fmt.Errorf("Spectrum plugin internal error"))
// 					createRequest := &models.CreateRequest{Name: "dockerVolume1", Opts: map[string]interface{}{"Filesystem": "gpfs1"}}
// 					createResponse := controller.Create(createRequest)
// 					Expect(createResponse.Err).To(Equal("Spectrum plugin internal error"))
// 				})
// 			})
// 			Context(".Remove", func() {
// 				It("does not error when existing dockerVolume name is given", func() {
// 					dockerVolume := &models.VolumeMetadata{Name: "dockerVolume1"}
// 					fakeClient.GetReturns(dockerVolume, nil, nil)
// 					removeRequest := &models.GenericRequest{Name: "dockerVolume1"}
// 					removeResponse := controller.Remove(removeRequest)
// 					Expect(removeResponse.Err).To(Equal(""))
// 				})
// 				It("error when dockerVolume not found", func() {
// 					fakeClient.GetReturns(nil, nil, nil)
// 					removeRequest := &models.GenericRequest{Name: "dockerVolume1"}
// 					removeResponse := controller.Remove(removeRequest)
// 					Expect(removeResponse.Err).To(Equal("Volume not found"))
// 					Expect(fakeClient.RemoveCallCount()).To(Equal(0))
// 				})
// 				It("error when list dockerVolume returns an error", func() {
// 					fakeClient.GetReturns(nil, nil, fmt.Errorf("error listing volume"))
// 					removeRequest := &models.GenericRequest{Name: "dockerVolume1"}
// 					removeResponse := controller.Remove(removeRequest)
// 					Expect(removeResponse.Err).To(Equal("error listing volume"))
// 					Expect(fakeClient.RemoveCallCount()).To(Equal(0))
// 				})
// 				It("error when remove dockerVolume returns an error", func() {
// 					dockerVolume := &models.VolumeMetadata{Name: "dockerVolume1"}
// 					fakeClient.GetReturns(dockerVolume, nil, nil)
// 					fakeClient.RemoveReturns(fmt.Errorf("error removing volume"))
// 					removeRequest := &models.GenericRequest{Name: "dockerVolume1"}
// 					removeResponse := controller.Remove(removeRequest)
// 					Expect(removeResponse.Err).To(Equal("error removing volume"))
// 					Expect(fakeClient.RemoveCallCount()).To(Equal(1))
// 				})
// 			})
// 			Context(".List", func() {
// 				It("does not error when volumes exist", func() {
// 					dockerVolume := models.VolumeMetadata{Name: "dockerVolume1"}
// 					var dockerVolumes []models.VolumeMetadata
// 					dockerVolumes = append(dockerVolumes, dockerVolume)
// 					fakeClient.ListReturns(dockerVolumes, nil)
// 					listResponse := controller.List()
// 					Expect(listResponse.Err).To(Equal(""))
// 					Expect(listResponse.Volumes).ToNot(Equal(nil))
// 					Expect(len(listResponse.Volumes)).To(Equal(1))
// 				})
// 				It("does not error when no volumes exist", func() {
// 					var dockerVolumes []models.VolumeMetadata
// 					fakeClient.ListReturns(dockerVolumes, nil)
// 					listResponse := controller.List()
// 					Expect(listResponse.Err).To(Equal(""))
// 					Expect(listResponse.Volumes).ToNot(Equal(nil))
// 					Expect(len(listResponse.Volumes)).To(Equal(0))
// 				})
// 				It("errors when client fails to list dockerVolumes", func() {
// 					fakeClient.ListReturns(nil, fmt.Errorf("failed to list volumes"))
// 					listResponse := controller.List()
// 					Expect(listResponse.Err).To(Equal("failed to list volumes"))
// 				})
// 			})
// 			Context(".Get", func() {
// 				It("does not error when volume exist", func() {
// 					dockerVolume := &models.VolumeMetadata{Name: "dockerVolume1"}
// 					fakeClient.GetReturns(dockerVolume, nil, nil)
// 					getRequest := &models.GenericRequest{Name: "dockerVolume1"}
// 					getResponse := controller.Get(getRequest)
// 					Expect(getResponse.Err).To(Equal(""))
// 					Expect(getResponse.Volume).ToNot(Equal(nil))
// 					Expect(getResponse.Volume.Name).To(Equal("dockerVolume1"))
// 				})
// 				It("errors when list dockerVolume returns an error", func() {
// 					fakeClient.GetReturns(nil, nil, fmt.Errorf("failed listing volume"))
// 					getRequest := &models.GenericRequest{Name: "dockerVolume1"}
// 					getResponse := controller.Get(getRequest)
// 					Expect(getResponse.Err).To(Equal("failed listing volume"))
// 				})
// 				It("errors when volume does not exist", func() {
// 					getRequest := &models.GenericRequest{Name: "dockerVolume1"}
// 					getResponse := controller.Get(getRequest)
// 					Expect(getResponse.Err).To(Equal("volume does not exist"))
// 				})
// 			})
// 			Context(".Path", func() {
// 				It("does not error when volume exists and is mounted", func() {
// 					dockerVolume := &models.VolumeMetadata{Name: "dockerVolume1", Mountpoint: "some-mountpoint"}
// 					fakeClient.GetReturns(dockerVolume, nil, nil)
// 					pathRequest := &models.GenericRequest{Name: "dockerVolume1"}
// 					pathResponse := controller.Path(pathRequest)
// 					Expect(pathResponse.Err).To(Equal(""))
// 					Expect(pathResponse.Mountpoint).To(Equal("some-mountpoint"))
// 				})
// 				It("errors when volume exists but is not mounted", func() {
// 					dockerVolume := &models.VolumeMetadata{Name: "dockerVolume1"}
// 					fakeClient.GetReturns(dockerVolume, nil, nil)
// 					pathRequest := &models.GenericRequest{Name: "dockerVolume1"}
// 					pathResponse := controller.Path(pathRequest)
// 					Expect(pathResponse.Err).To(Equal("volume not mounted"))
// 				})
// 				It("errors when list dockerVolume returns an error", func() {
// 					fakeClient.GetReturns(nil, nil, fmt.Errorf("failed listing volume"))
// 					pathRequest := &models.GenericRequest{Name: "dockerVolume1"}
// 					pathResponse := controller.Path(pathRequest)
// 					Expect(pathResponse.Err).To(Equal("failed listing volume"))
// 				})
// 				It("errors when volume does not exist", func() {
// 					pathRequest := &models.GenericRequest{Name: "dockerVolume1"}
// 					pathResponse := controller.Path(pathRequest)
// 					Expect(pathResponse.Err).To(Equal("volume does not exist"))
// 				})
// 			})
// 			Context(".Mount", func() {
// 				It("does not error when volume exists and is not currently mounted", func() {
// 					dockerVolume := &models.VolumeMetadata{Name: "dockerVolume1"}
// 					fakeClient.GetReturns(dockerVolume, nil, nil)
// 					fakeClient.AttachReturns("some-mountpath", nil)
// 					mountRequest := &models.GenericRequest{Name: "dockerVolume1"}
// 					mountResponse := controller.Mount(mountRequest)
// 					Expect(mountResponse.Err).To(Equal(""))
// 					Expect(mountResponse.Mountpoint).To(Equal("some-mountpath"))
// 					Expect(fakeClient.AttachCallCount()).To(Equal(1))
// 				})
// 				It("errors when volume list returns error", func() {
// 					fakeClient.GetReturns(nil, nil, fmt.Errorf("error listing volume"))
// 					mountRequest := &models.GenericRequest{Name: "dockerVolume1"}
// 					mountResponse := controller.Mount(mountRequest)
// 					Expect(mountResponse.Err).To(Equal("error listing volume"))
// 				})
// 				It("errors when volume does not exist", func() {
// 					fakeClient.GetReturns(nil, nil, nil)
// 					mountRequest := &models.GenericRequest{Name: "dockerVolume1"}
// 					mountResponse := controller.Mount(mountRequest)
// 					Expect(mountResponse.Err).To(Equal("volume not found"))
// 				})
// 				It("errors when volume exists and LinkdockerVolume errors", func() {
// 					dockerVolume := &models.VolumeMetadata{Name: "dockerVolume1"}
// 					fakeClient.GetReturns(dockerVolume, nil, nil)
// 					fakeClient.AttachReturns("", fmt.Errorf("failed to link volume"))
// 					mountRequest := &models.GenericRequest{Name: "dockerVolume1"}
// 					mountResponse := controller.Mount(mountRequest)
// 					Expect(mountResponse.Err).To(Equal("failed to link volume"))
// 				})
// 			})
// 			Context(".Unmount", func() {
// 				It("does not error when volume exists and is currently mounted", func() {
// 					dockerVolume := &models.VolumeMetadata{Name: "dockerVolume1", Mountpoint: "some-mountpoint"}
// 					fakeClient.GetReturns(dockerVolume, nil, nil)
// 					unmountRequest := &models.GenericRequest{Name: "dockerVolume1"}
// 					unmountResponse := controller.Unmount(unmountRequest)
// 					Expect(unmountResponse.Err).To(Equal(""))
// 				})
// 				It("errors when volume list returns error", func() {
// 					fakeClient.GetReturns(nil, nil, fmt.Errorf("error listing volume"))
// 					unmountRequest := &models.GenericRequest{Name: "dockerVolume1"}
// 					unmountResponse := controller.Unmount(unmountRequest)
// 					Expect(unmountResponse.Err).To(Equal("error listing volume"))
// 				})
// 				It("errors when volume does not exist", func() {
// 					fakeClient.GetReturns(nil, nil, nil)
// 					unmountRequest := &models.GenericRequest{Name: "dockerVolume1"}
// 					unmountResponse := controller.Unmount(unmountRequest)
// 					Expect(unmountResponse.Err).To(Equal("volume not found"))
// 				})
// 				It("errors when volume exists and is currently not mounted", func() {
// 					dockerVolume := &models.VolumeMetadata{Name: "dockerVolume1"}
// 					fakeClient.GetReturns(dockerVolume, nil, nil)
// 					unmountRequest := &models.GenericRequest{Name: "dockerVolume1"}
// 					unmountResponse := controller.Unmount(unmountRequest)
// 					Expect(unmountResponse.Err).To(Equal("volume already unmounted"))
// 					Expect(fakeClient.DetachCallCount()).To(Equal(0))
// 				})
// 				It("errors when volume exists and UnLinkdockerVolume errors", func() {
// 					dockerVolume := &models.VolumeMetadata{Name: "dockerVolume1", Mountpoint: "some-mountpoint"}
// 					fakeClient.GetReturns(dockerVolume, nil, nil)
// 					fakeClient.DetachReturns(fmt.Errorf("failed to unlink volume"))
// 					unmountRequest := &models.GenericRequest{Name: "dockerVolume1"}
// 					unmountResponse := controller.Unmount(unmountRequest)
// 					Expect(unmountResponse.Err).To(Equal("failed to unlink volume"))
// 				})
// 			})
// 		})
// 	})
// })

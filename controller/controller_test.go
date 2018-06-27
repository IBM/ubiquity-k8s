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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"encoding/json"
	"fmt"

	ctl "github.com/IBM/ubiquity-k8s/controller"
	k8sresources "github.com/IBM/ubiquity-k8s/resources"
	"github.com/IBM/ubiquity/fakes"
	"github.com/IBM/ubiquity/resources"
)

var _ = Describe("Controller", func() {

	var (
		fakeClient         *fakes.FakeStorageClient
		controller         *ctl.Controller
		fakeExec           *fakes.FakeExecutor
		fakeMounterFactory *fakes.FakeMounterFactory
		fakeMounter        *fakes.FakeMounter
		ubiquityConfig     resources.UbiquityPluginConfig
		dat                map[string]interface{}
	)
	BeforeEach(func() {
		fakeExec = new(fakes.FakeExecutor)
		ubiquityConfig = resources.UbiquityPluginConfig{}
		fakeClient = new(fakes.FakeStorageClient)
		fakeMounterFactory = new(fakes.FakeMounterFactory)
		fakeMounter = new(fakes.FakeMounter)
		controller = ctl.NewControllerWithClient(testLogger, ubiquityConfig, fakeClient, fakeExec, fakeMounterFactory)
		byt := []byte(`{"Wwn":"fake"}`)
		if err := json.Unmarshal(byt, &dat); err != nil {
			panic(err)
		}

	})

	Context(".Init", func() {

		It("does not error when init is successful", func() {
			initResponse := controller.Init(ubiquityConfig)
			Expect(initResponse.Status).To(Equal("Success"))
			Expect(initResponse.Message).To(Equal("Plugin init successfully"))
			Expect(initResponse.Device).To(Equal(""))
		})

		//Context(".Attach", func() {
		//
		//	It("fails when attachRequest does not have volumeName", func() {
		//		fakeClient.GetVolumeReturns(resources.Volume{}, fmt.Errorf("GetVolume error"))
		//		attachRequest := map[string]string{"Filesystem": "gpfs1", "Size": "200m", "Fileset": "fs1", "Path": "myPath"}
		//		attachResponse := controller.Attach(attachRequest)
		//		Expect(attachResponse.Status).To(Equal("Failure"))
		//		Expect(fakeClient.GetVolumeCallCount()).To(Equal(0))
		//	})
		//
		//	It("fails when client fails to fetch volume", func() {
		//		fakeClient.GetVolumeReturns(resources.Volume{}, fmt.Errorf("GetVolume error"))
		//		attachRequest := map[string]string{"volumeName": "vol1", "Filesystem": "gpfs1", "Size": "200m", "Fileset": "fs1", "Path": "myPath"}
		//		attachResponse := controller.Attach(attachRequest)
		//		Expect(attachResponse.Status).To(Equal("Failure"))
		//		Expect(attachResponse.Message).To(Equal("Failed checking volume, call create before attach"))
		//		Expect(attachResponse.Device).To(Equal("vol1"))
		//	})
		//
		//	It("Succeeds when volume exists", func() {
		//		fakeClient.GetVolumeReturns(resources.Volume{}, nil)
		//		attachRequest := map[string]string{"volumeName": "vol1", "Filesystem": "gpfs1", "Size": "200m", "Fileset": "fs1", "Path": "myPath"}
		//		attachResponse := controller.Attach(attachRequest)
		//		Expect(attachResponse.Status).To(Equal("Success"))
		//		Expect(attachResponse.Message).To(Equal("Volume already attached"))
		//		Expect(attachResponse.Device).To(Equal("vol1"))
		//		Expect(fakeClient.CreateVolumeCallCount()).To(Equal(0))
		//	})
		//})
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
	})
	Context(".Mount", func() {
		It("should fail if k8s version < 1.6 (doMount)", func() {
			mountRequest := k8sresources.FlexVolumeMountRequest{MountPath: "fake", MountDevice: "pv1", Opts: map[string]string{"Wwn": "fake"}, Version: k8sresources.KubernetesVersion_1_5}

			mountResponse := controller.Mount(mountRequest)

			Expect(mountResponse.Message).To(MatchRegexp(ctl.SupportK8sVesion))
			Expect(mountResponse.Status).To(Equal(ctl.FlexFailureStr))
			Expect(mountResponse.Device).To(Equal(""))
		})
		It("should fail in GetVolume if volume not in ubiqutiyDB (doMount)", func() {
			fakeClient.GetVolumeReturns(resources.Volume{}, fmt.Errorf("error not found in DB"))
			mountRequest := k8sresources.FlexVolumeMountRequest{MountPath: "/pod/pv1", MountDevice: "pv1", Opts: map[string]string{}}

			mountResponse := controller.Mount(mountRequest)

			Expect(fakeClient.GetVolumeCallCount()).To(Equal(1))
			Expect(mountResponse.Message).To(MatchRegexp(".*error not found in DB"))
			Expect(mountResponse.Status).To(Equal(ctl.FlexFailureStr))
			Expect(mountResponse.Device).To(Equal(""))
		})
		It("should fail if cannot get mounter for the PV (doMount, GetMounterPerBackend)", func() {
			errstr := "ERROR backend"
			fakeClient.GetVolumeReturns(resources.Volume{Name: "pv1", Backend: "XXX", Mountpoint: "fake"}, nil)
			fakeMounterFactory.GetMounterPerBackendReturns(fakeMounter, fmt.Errorf(errstr))
			controller = ctl.NewControllerWithClient(testLogger, ubiquityConfig, fakeClient, fakeExec, fakeMounterFactory)
			mountRequest := k8sresources.FlexVolumeMountRequest{MountPath: "/pod/pv1", MountDevice: "pv1", Opts: map[string]string{}}

			mountResponse := controller.Mount(mountRequest)

			Expect(mountResponse.Device).To(Equal(""))
			Expect(fakeClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeMounterFactory.GetMounterPerBackendCallCount()).To(Equal(1))
			Expect(fakeClient.GetVolumeConfigCallCount()).To(Equal(0))
			Expect(mountResponse.Message).To(MatchRegexp(errstr))
			Expect(mountResponse.Status).To(Equal(ctl.FlexFailureStr))
		})
		It("should fail to prepareUbiquityMountRequest if GetVolumeConfig failed (doMount)", func() {
			errstr := "error GetVolumeConfig"
			fakeClient.GetVolumeReturns(resources.Volume{Name: "pv1", Backend: "aaaa", Mountpoint: "fake"}, nil)
			byt := []byte(`{"":""}`)
			var dat map[string]interface{}
			if err := json.Unmarshal(byt, &dat); err != nil {
				panic(err)
			}
			fakeClient.GetVolumeConfigReturns(dat, fmt.Errorf(errstr))
			mountRequest := k8sresources.FlexVolumeMountRequest{MountPath: "/pod/pv1", MountDevice: "pv1", Opts: map[string]string{}}

			mountResponse := controller.Mount(mountRequest)

			Expect(fakeClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeMounterFactory.GetMounterPerBackendCallCount()).To(Equal(1))
			Expect(fakeClient.GetVolumeConfigCallCount()).To(Equal(1))
			Expect(fakeMounter.MountCallCount()).To(Equal(0))
			Expect(mountResponse.Message).To(MatchRegexp(errstr))
			Expect(mountResponse.Status).To(Equal(ctl.FlexFailureStr))
		})
		It("should fail to prepareUbiquityMountRequest if GetVolumeConfig does not contain Wwn key (doMount)", func() {
			fakeClient.GetVolumeReturns(resources.Volume{Name: "pv1", Backend: "scbe", Mountpoint: "fake"}, nil)
			byt := []byte(`{"fake":"fake"}`)
			var dat map[string]interface{}
			if err := json.Unmarshal(byt, &dat); err != nil {
				panic(err)
			}
			fakeClient.GetVolumeConfigReturns(dat, nil)
			mountRequest := k8sresources.FlexVolumeMountRequest{MountPath: "/pod/pv1", MountDevice: "pv1", Opts: map[string]string{}}

			mountResponse := controller.Mount(mountRequest)

			Expect(fakeClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeMounterFactory.GetMounterPerBackendCallCount()).To(Equal(1))
			Expect(fakeClient.GetVolumeConfigCallCount()).To(Equal(1))
			Expect(fakeMounter.MountCallCount()).To(Equal(0))
			Expect(mountResponse.Message).To(Equal(ctl.MissingWwnMountRequestErrorStr))
			Expect(mountResponse.Status).To(Equal(ctl.FlexFailureStr))
		})

		It("should fail if mounter.Mount failed (doMount)", func() {
			errstr := "TODO set error in mounter"
			fakeClient.GetVolumeReturns(resources.Volume{Name: "pv1", Backend: "scbe", Mountpoint: "fake"}, nil)
			byt := []byte(`{"Wwn":"fake"}`)
			var dat map[string]interface{}
			if err := json.Unmarshal(byt, &dat); err != nil {
				panic(err)
			}
			fakeClient.GetVolumeConfigReturns(dat, nil)
			fakeMounter.MountReturns("fake device", fmt.Errorf(errstr))
			fakeMounterFactory.GetMounterPerBackendReturns(fakeMounter, nil)
			controller = ctl.NewControllerWithClient(testLogger, ubiquityConfig, fakeClient, fakeExec, fakeMounterFactory)
			mountRequest := k8sresources.FlexVolumeMountRequest{MountPath: "/pod/pv1", MountDevice: "pv1", Opts: map[string]string{"Wwn": "fake"}}

			mountResponse := controller.Mount(mountRequest)

			Expect(fakeClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeClient.GetVolumeConfigCallCount()).To(Equal(1))
			Expect(fakeMounter.MountCallCount()).To(Equal(1))
			Expect(fakeExec.LstatCallCount()).To(Equal(0))
			Expect(mountResponse.Message).To(Equal(errstr))
			Expect(mountResponse.Status).To(Equal(ctl.FlexFailureStr))
		})
		It("should fail to Mount if fail to lstat k8s-mountpoint (doAfterMount)", func() {
			errstr := "fakerror"
			errstrObj := fmt.Errorf(errstr)

			fakeClient.GetVolumeReturns(resources.Volume{Name: "pv1", Backend: "scbe", Mountpoint: "fake"}, nil)
			fakeClient.GetVolumeConfigReturns(dat, nil)
			fakeMounter.MountReturns("/ubiquity/wwn1", nil)
			fakeMounterFactory.GetMounterPerBackendReturns(fakeMounter, nil)
			controller = ctl.NewControllerWithClient(testLogger, ubiquityConfig, fakeClient, fakeExec, fakeMounterFactory)
			fakeExec.LstatReturns(nil, errstrObj)
			fakeExec.IsNotExistReturns(false)
			mountRequest := k8sresources.FlexVolumeMountRequest{MountPath: "/pod/pv1", MountDevice: "pv1", Opts: map[string]string{"Wwn": "fake"}, Version: k8sresources.KubernetesVersion_1_6OrLater}

			mountResponse := controller.Mount(mountRequest)

			Expect(fakeClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeClient.GetVolumeConfigCallCount()).To(Equal(1))
			Expect(fakeMounter.MountCallCount()).To(Equal(1))
			Expect(fakeExec.LstatCallCount()).To(Equal(1))
			Expect(fakeExec.IsNotExistCallCount()).To(Equal(1))
			Expect(fakeExec.IsNotExistArgsForCall(0)).To(Equal(errstrObj))
			Expect(fakeExec.IsDirCallCount()).To(Equal(0))
			Expect(mountResponse.Message).To(Equal(errstr))
			Expect(mountResponse.Status).To(Equal(ctl.FlexFailureStr))
		})
		It("should fail to Mount k8s-mountpoint dir not exist(idempotent) but and failed to slink (doAfterMount)", func() {
			errstr := "fakerror"
			errstrObj := fmt.Errorf(errstr)

			fakeClient.GetVolumeReturns(resources.Volume{Name: "pv1", Backend: "scbe", Mountpoint: "fake"}, nil)
			fakeClient.GetVolumeConfigReturns(dat, nil)
			fakeMounter.MountReturns("/ubiquity/wwn1", nil)
			fakeMounterFactory.GetMounterPerBackendReturns(fakeMounter, nil)
			controller = ctl.NewControllerWithClient(testLogger, ubiquityConfig, fakeClient, fakeExec, fakeMounterFactory)
			fakeExec.LstatReturns(nil, errstrObj)
			fakeExec.IsNotExistReturns(true)
			fakeExec.SymlinkReturns(errstrObj)
			mountRequest := k8sresources.FlexVolumeMountRequest{MountPath: "/pod/pv1", MountDevice: "pv1", Opts: map[string]string{"Wwn": "fake"}, Version: k8sresources.KubernetesVersion_1_6OrLater}

			mountResponse := controller.Mount(mountRequest)

			Expect(fakeClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeClient.GetVolumeConfigCallCount()).To(Equal(1))
			Expect(fakeMounter.MountCallCount()).To(Equal(1))
			Expect(fakeExec.LstatCallCount()).To(Equal(1))
			Expect(fakeExec.IsNotExistCallCount()).To(Equal(1))
			Expect(fakeExec.IsNotExistArgsForCall(0)).To(Equal(errstrObj))
			Expect(fakeExec.IsDirCallCount()).To(Equal(0))
			Expect(mountResponse.Message).To(Equal(errstr))
			Expect(mountResponse.Status).To(Equal(ctl.FlexFailureStr))
		})
		It("should succeed to Mount k8s-mountpoint dir not exist(idempotent) and slink succeed (doAfterMount)", func() {
			errstr := "fakerror"
			errstrObj := fmt.Errorf(errstr)

			fakeClient.GetVolumeReturns(resources.Volume{Name: "pv1", Backend: "scbe", Mountpoint: "fake"}, nil)
			fakeClient.GetVolumeConfigReturns(dat, nil)
			mountpoint := "/ubiquity/wwn1"
			fakeMounter.MountReturns(mountpoint, nil)
			fakeMounterFactory.GetMounterPerBackendReturns(fakeMounter, nil)
			controller = ctl.NewControllerWithClient(testLogger, ubiquityConfig, fakeClient, fakeExec, fakeMounterFactory)
			fakeExec.LstatReturns(nil, errstrObj)
			fakeExec.IsNotExistReturns(true)
			fakeExec.SymlinkReturns(nil)
			mountRequest := k8sresources.FlexVolumeMountRequest{MountPath: "/pod/pv1", MountDevice: "pv1", Opts: map[string]string{"Wwn": "fake"}, Version: k8sresources.KubernetesVersion_1_6OrLater}

			mountResponse := controller.Mount(mountRequest)

			Expect(fakeClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeClient.GetVolumeConfigCallCount()).To(Equal(1))
			Expect(fakeMounter.MountCallCount()).To(Equal(1))
			Expect(fakeExec.LstatCallCount()).To(Equal(1))
			Expect(fakeExec.IsNotExistCallCount()).To(Equal(1))
			Expect(fakeExec.SymlinkCallCount()).To(Equal(1))
			Expect(fakeExec.IsDirCallCount()).To(Equal(0))
			Expect(mountResponse.Message).To(Equal(""))
			Expect(mountResponse.Status).To(Equal(ctl.FlexSuccessStr))
			Expect(mountResponse.Device).To(Equal(""))
		})
		It("should fail to Mount because fail to Remove the k8s-mountpoint dir (doAfterMount)", func() {
			errstr := "fakerror"
			errstrObj := fmt.Errorf(errstr)

			fakeClient.GetVolumeReturns(resources.Volume{Name: "pv1", Backend: "scbe", Mountpoint: "fake"}, nil)
			fakeClient.GetVolumeConfigReturns(dat, nil)
			fakeMounter.MountReturns("/ubiquity/wwn1", nil)
			fakeMounterFactory.GetMounterPerBackendReturns(fakeMounter, nil)
			controller = ctl.NewControllerWithClient(testLogger, ubiquityConfig, fakeClient, fakeExec, fakeMounterFactory)
			mountRequest := k8sresources.FlexVolumeMountRequest{MountPath: "/tmp", MountDevice: "pv1", Opts: map[string]string{"Wwn": "fake"}, Version: k8sresources.KubernetesVersion_1_6OrLater}
			fakeExec.LstatReturns(nil, nil)
			fakeExec.IsDirReturns(true)
			fakeExec.RemoveReturns(errstrObj)

			mountResponse := controller.Mount(mountRequest)

			Expect(fakeClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeClient.GetVolumeConfigCallCount()).To(Equal(1))
			Expect(fakeMounter.MountCallCount()).To(Equal(1))
			Expect(fakeExec.LstatCallCount()).To(Equal(1))
			Expect(fakeExec.IsNotExistCallCount()).To(Equal(0))
			Expect(fakeExec.IsDirCallCount()).To(Equal(1))
			Expect(fakeExec.RemoveCallCount()).To(Equal(1))
			Expect(fakeExec.SymlinkCallCount()).To(Equal(0))
			Expect(mountResponse.Message).To(MatchRegexp(ctl.FailRemovePVorigDirErrorStr))
			Expect(mountResponse.Status).To(Equal(ctl.FlexFailureStr))
		})
		It("should fail to Mount because fail to create Symlink after Remove the k8s-mountpoint dir (doAfterMount)", func() {
			errstr := "fakerror"
			errstrObj := fmt.Errorf(errstr)

			fakeClient.GetVolumeReturns(resources.Volume{Name: "pv1", Backend: "scbe", Mountpoint: "fake"}, nil)
			fakeClient.GetVolumeConfigReturns(dat, nil)
			fakeMounter.MountReturns("/ubiquity/wwn1", nil)
			fakeMounterFactory.GetMounterPerBackendReturns(fakeMounter, nil)
			controller = ctl.NewControllerWithClient(testLogger, ubiquityConfig, fakeClient, fakeExec, fakeMounterFactory)
			mountRequest := k8sresources.FlexVolumeMountRequest{MountPath: "/tmp", MountDevice: "pv1", Opts: map[string]string{"Wwn": "fake"}, Version: k8sresources.KubernetesVersion_1_6OrLater}
			fakeExec.LstatReturns(nil, nil)
			fakeExec.IsDirReturns(true)
			fakeExec.RemoveReturns(nil)
			fakeExec.SymlinkReturns(errstrObj)

			mountResponse := controller.Mount(mountRequest)

			Expect(fakeClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeClient.GetVolumeConfigCallCount()).To(Equal(1))
			Expect(fakeMounter.MountCallCount()).To(Equal(1))
			Expect(fakeExec.LstatCallCount()).To(Equal(1))
			Expect(fakeExec.IsNotExistCallCount()).To(Equal(0))
			Expect(fakeExec.IsDirCallCount()).To(Equal(1))
			Expect(fakeExec.RemoveCallCount()).To(Equal(1))
			Expect(fakeExec.SymlinkCallCount()).To(Equal(1))
			Expect(fakeExec.IsSlinkCallCount()).To(Equal(0))
			Expect(mountResponse.Message).To(MatchRegexp(errstr))
			Expect(mountResponse.Status).To(Equal(ctl.FlexFailureStr))
		})

		It("should fail to Mount because fail to create Symlink after Remove the k8s-mountpoint dir (doAfterMount)", func() {
			fakeClient.GetVolumeReturns(resources.Volume{Name: "pv1", Backend: "scbe", Mountpoint: "fake"}, nil)
			fakeClient.GetVolumeConfigReturns(dat, nil)
			fakeMounter.MountReturns("/ubiquity/wwn1", nil)
			fakeMounterFactory.GetMounterPerBackendReturns(fakeMounter, nil)
			controller = ctl.NewControllerWithClient(testLogger, ubiquityConfig, fakeClient, fakeExec, fakeMounterFactory)
			mountRequest := k8sresources.FlexVolumeMountRequest{MountPath: "/tmp", MountDevice: "pv1", Opts: map[string]string{"Wwn": "fake"}, Version: k8sresources.KubernetesVersion_1_6OrLater}
			fakeExec.LstatReturns(nil, nil)
			fakeExec.IsDirReturns(true)
			fakeExec.RemoveReturns(nil)
			fakeExec.SymlinkReturns(nil)

			mountResponse := controller.Mount(mountRequest)

			Expect(fakeClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeClient.GetVolumeConfigCallCount()).To(Equal(1))
			Expect(fakeMounter.MountCallCount()).To(Equal(1))
			Expect(fakeExec.LstatCallCount()).To(Equal(1))
			Expect(fakeExec.IsNotExistCallCount()).To(Equal(0))
			Expect(fakeExec.IsDirCallCount()).To(Equal(1))
			Expect(fakeExec.RemoveCallCount()).To(Equal(1))
			Expect(fakeExec.SymlinkCallCount()).To(Equal(1))
			Expect(fakeExec.IsSlinkCallCount()).To(Equal(0))
			Expect(mountResponse.Message).To(Equal(""))
			Expect(mountResponse.Status).To(Equal(ctl.FlexSuccessStr))
		})

		It("should fail to Mount because fail to EvalSymlinks (idempotent) (doAfterMount)", func() {
			errstr := "fakerror"
			errstrObj := fmt.Errorf(errstr)

			fakeClient.GetVolumeReturns(resources.Volume{Name: "pv1", Backend: "scbe", Mountpoint: "fake"}, nil)
			fakeClient.GetVolumeConfigReturns(dat, nil)
			fakeMounter.MountReturns("/ubiquity/wwn1", nil)
			fakeMounterFactory.GetMounterPerBackendReturns(fakeMounter, nil)
			controller = ctl.NewControllerWithClient(testLogger, ubiquityConfig, fakeClient, fakeExec, fakeMounterFactory)
			mountRequest := k8sresources.FlexVolumeMountRequest{MountPath: "/tmp", MountDevice: "pv1", Opts: map[string]string{"Wwn": "fake"}, Version: k8sresources.KubernetesVersion_1_6OrLater}
			fakeExec.LstatReturns(nil, nil)
			fakeExec.IsDirReturns(false)
			fakeExec.IsSlinkReturns(true)
			fakeExec.EvalSymlinksReturns("", errstrObj)

			mountResponse := controller.Mount(mountRequest)

			Expect(fakeClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeClient.GetVolumeConfigCallCount()).To(Equal(1))
			Expect(fakeMounter.MountCallCount()).To(Equal(1))
			Expect(fakeExec.LstatCallCount()).To(Equal(1))
			Expect(fakeExec.IsNotExistCallCount()).To(Equal(0))
			Expect(fakeExec.IsDirCallCount()).To(Equal(1))
			Expect(fakeExec.IsSlinkCallCount()).To(Equal(1))
			Expect(fakeExec.EvalSymlinksCallCount()).To(Equal(1))
			Expect(mountResponse.Message).To(Equal(errstr))
			Expect(mountResponse.Status).To(Equal(ctl.FlexFailureStr))
		})
		It("should succeed to Mount after k8s-mountpoint is already slink(idempotent) and point to the right mountpath (doAfterMount)", func() {
			fakeClient.GetVolumeReturns(resources.Volume{Name: "pv1", Backend: "scbe", Mountpoint: "fake"}, nil)
			fakeClient.GetVolumeConfigReturns(dat, nil)
			mountPath := "/ubiquity/wwn1"

			fakeMounter.MountReturns(mountPath, nil)
			fakeMounterFactory.GetMounterPerBackendReturns(fakeMounter, nil)
			controller = ctl.NewControllerWithClient(testLogger, ubiquityConfig, fakeClient, fakeExec, fakeMounterFactory)
			mountRequest := k8sresources.FlexVolumeMountRequest{MountPath: "/pod/pv1", MountDevice: "pv1", Opts: map[string]string{"Wwn": "fake"}, Version: k8sresources.KubernetesVersion_1_6OrLater}
			fakeExec.LstatReturns(nil, nil)
			fakeExec.IsDirReturns(false)
			fakeExec.IsSlinkReturns(true)
			fakeExec.EvalSymlinksReturns(mountPath, nil)

			mountResponse := controller.Mount(mountRequest)

			Expect(fakeClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeClient.GetVolumeConfigCallCount()).To(Equal(1))
			Expect(fakeMounter.MountCallCount()).To(Equal(1))
			Expect(fakeExec.LstatCallCount()).To(Equal(1))
			Expect(fakeExec.IsNotExistCallCount()).To(Equal(0))
			Expect(fakeExec.IsDirCallCount()).To(Equal(1))
			Expect(fakeExec.IsSlinkCallCount()).To(Equal(1))
			Expect(fakeExec.EvalSymlinksCallCount()).To(Equal(1))
			Expect(mountResponse.Message).To(Equal(""))
			Expect(mountResponse.Status).To(Equal(ctl.FlexSuccessStr))
			Expect(mountResponse.Device).To(Equal(""))
		})

		It("should fail to Mount after k8s-mountpoint is already slink(idempotent) but point to wrong mountpath (doAfterMount)", func() {
			fakeClient.GetVolumeReturns(resources.Volume{Name: "pv1", Backend: "scbe", Mountpoint: "fake"}, nil)
			fakeClient.GetVolumeConfigReturns(dat, nil)
			mountPath := "/ubiquity/wwn1"

			fakeMounter.MountReturns(mountPath, nil)
			fakeMounterFactory.GetMounterPerBackendReturns(fakeMounter, nil)
			controller = ctl.NewControllerWithClient(testLogger, ubiquityConfig, fakeClient, fakeExec, fakeMounterFactory)
			mountRequest := k8sresources.FlexVolumeMountRequest{MountPath: "/pod/pv1", MountDevice: "pv1", Opts: map[string]string{"Wwn": "fake"}, Version: k8sresources.KubernetesVersion_1_6OrLater}
			fakeExec.LstatReturns(nil, nil)
			fakeExec.IsDirReturns(false)
			fakeExec.IsSlinkReturns(true)
			badMountPointForSlink := mountPath + "/bad"
			fakeExec.EvalSymlinksReturns(badMountPointForSlink, nil)

			mountResponse := controller.Mount(mountRequest)

			Expect(fakeClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeClient.GetVolumeConfigCallCount()).To(Equal(1))
			Expect(fakeMounter.MountCallCount()).To(Equal(1))
			Expect(fakeExec.LstatCallCount()).To(Equal(1))
			Expect(fakeExec.IsNotExistCallCount()).To(Equal(0))
			Expect(fakeExec.IsDirCallCount()).To(Equal(1))
			Expect(fakeExec.IsSlinkCallCount()).To(Equal(1))
			Expect(fakeExec.EvalSymlinksCallCount()).To(Equal(1))
			Expect(mountResponse.Message).To(MatchRegexp(ctl.WrongSlinkErrorStr))
			Expect(mountResponse.Status).To(Equal(ctl.FlexFailureStr))
		})
		It("should fail to Mount when k8s-mountpoint exist but not as dir nor as slink (idempotent) (doAfterMount)", func() {
			fakeClient.GetVolumeReturns(resources.Volume{Name: "pv1", Backend: "scbe", Mountpoint: "fake"}, nil)
			fakeClient.GetVolumeConfigReturns(dat, nil)
			mountPath := "/ubiquity/wwn1"

			fakeMounter.MountReturns(mountPath, nil)
			fakeMounterFactory.GetMounterPerBackendReturns(fakeMounter, nil)
			controller = ctl.NewControllerWithClient(testLogger, ubiquityConfig, fakeClient, fakeExec, fakeMounterFactory)
			mountRequest := k8sresources.FlexVolumeMountRequest{MountPath: "/pod/pv1", MountDevice: "pv1", Opts: map[string]string{"Wwn": "fake"}, Version: k8sresources.KubernetesVersion_1_6OrLater}
			fakeExec.LstatReturns(nil, nil)
			fakeExec.IsDirReturns(false)
			fakeExec.IsSlinkReturns(false)

			mountResponse := controller.Mount(mountRequest)

			Expect(fakeClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeClient.GetVolumeConfigCallCount()).To(Equal(1))
			Expect(fakeMounter.MountCallCount()).To(Equal(1))
			Expect(fakeExec.LstatCallCount()).To(Equal(1))
			Expect(fakeExec.IsNotExistCallCount()).To(Equal(0))
			Expect(fakeExec.IsDirCallCount()).To(Equal(1))
			Expect(fakeExec.IsSlinkCallCount()).To(Equal(1))
			Expect(fakeExec.EvalSymlinksCallCount()).To(Equal(0))
			Expect(mountResponse.Message).To(MatchRegexp(ctl.K8sPVDirectoryIsNotDirNorSlinkErrorStr))
			Expect(mountResponse.Status).To(Equal(ctl.FlexFailureStr))
		})

	})
	Context(".IsAttached", func() {
		It("should succeed if volume does not exist", func() {
			byt := []byte(`{"":""}`)
			var dat map[string]interface{}
			if err := json.Unmarshal(byt, &dat); err != nil {
				panic(err)
			}		
				
			fakeClient.GetVolumeConfigReturns(dat, fmt.Errorf("volume vol1 not found"))
			isAttachedRequest := k8sresources.FlexVolumeIsAttachedRequest{Name: "vol1", Host: "host1",
				Opts: map[string]string{"volumeName": "vol1"}}

			isAttachResponse := controller.IsAttached(isAttachedRequest)
			fmt.Println(isAttachResponse)
			Expect(isAttachResponse).To(Equal(k8sresources.FlexVolumeResponse{Status:"Success", Attached:false}))

		})
		It("should fail if other error is thrown from GetVolumeConfig", func() {
			byt := []byte(`{"":""}`)
			var dat map[string]interface{}
			if err := json.Unmarshal(byt, &dat); err != nil {
				panic(err)
			}		
				
			fakeClient.GetVolumeConfigReturns(dat, fmt.Errorf("Other error"))
			isAttachedRequest := k8sresources.FlexVolumeIsAttachedRequest{Name: "vol1", Host: "host1",
				Opts: map[string]string{"volumeName": "vol1"}}

			isAttachResponse := controller.IsAttached(isAttachedRequest)
			fmt.Println(isAttachResponse)
			Expect(isAttachResponse.Status).To(Equal("Failure"))
			Expect(isAttachResponse.Attached).To(Equal(false))
		})
		It("should succeed when no errors thrown", func() {
			isAttachedRequest := k8sresources.FlexVolumeIsAttachedRequest{Name: "vol1", Host: "host1",
				Opts: map[string]string{"volumeName": "vol1"}}

			isAttachResponse := controller.IsAttached(isAttachedRequest)
			fmt.Println(isAttachResponse)
			Expect(isAttachResponse).To(Equal(k8sresources.FlexVolumeResponse{Status:"Success", Attached:false}))
		})
	})

	/*
		Context(".Mount", func() {
			AfterEach(func() {

				err := os.RemoveAll("/tmp/test/mnt1")
				Expect(err).ToNot(HaveOccurred())

			})
			It("does not error when volume exists and is not currently mounted", func() {
				fakeClient.AttachReturns("/tmp/test/mnt1", nil)

				mountRequest := k8sresources.FlexVolumeMountRequest{MountPath: "/tmp/test/mnt2", MountDevice: "vol1", Opts: map[string]string{}}
				mountResponse := controller.Mount(mountRequest)
				Expect(mountResponse.Message).To(Equal("Volume mounted successfully to /tmp/test/mnt1"))
				Expect(mountResponse.Status).To(Equal("Success"))

				Expect(mountResponse.Device).To(Equal(""))
				Expect(fakeClient.AttachCallCount()).To(Equal(1))
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
	*/
})

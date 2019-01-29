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

package volume_test

import (
	"fmt"

	"github.com/IBM/ubiquity-k8s/volume"
	"github.com/IBM/ubiquity/fakes"
	"github.com/IBM/ubiquity/resources"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Provisioner", func() {
	var (
		fakeClient *fakes.FakeStorageClient
		// fakeKubeInterface *k8s_fake.FakeInterface
		provisioner    controller.Provisioner
		options        controller.VolumeOptions
		backends       []string
		ubiquityConfig resources.UbiquityPluginConfig
		err            error
	)

	BeforeEach(func() {
		fakeClient = new(fakes.FakeStorageClient)
		backends = []string{resources.SpectrumScale}
		ubiquityConfig = resources.UbiquityPluginConfig{Backends: backends}
		// fakeKubeInterface = new(k8s_fake.FakeInterface)
		provisioner, err = volume.NewFlexProvisioner(testLogger, fakeClient, ubiquityConfig)
	})

	Context(".Provision", func() {
		BeforeEach(func() {

			options = controller.VolumeOptions{PVName: "fakepv"}
		})
		It("fails when options does not contain pvc", func() {
			_, err = provisioner.Provision(options)
			Expect(err).To(HaveOccurred())
		})
		It("fails when options does not contain pvc", func() {
			_, err = provisioner.Provision(options)
			Expect(err).To(HaveOccurred())
		})

	})

	Context(".Delete", func() {

		It("fails when volume name is empty", func() {
			volume := v1.PersistentVolume{}
			err = provisioner.Delete(&volume)
			Expect(err).To(HaveOccurred())
			Expect(fakeClient.RemoveVolumeCallCount()).To(Equal(0))
		})

		It("fails when ubiquityClient returns an error", func() {
			fakeClient.RemoveVolumeReturns(fmt.Errorf("error removing volume"))
			objectMeta := metav1.ObjectMeta{Name: "vol1"}
			volume := v1.PersistentVolume{ObjectMeta: objectMeta}
			err = provisioner.Delete(&volume)
			Expect(err).To(HaveOccurred())
			Expect(fakeClient.RemoveVolumeCallCount()).To(Equal(1))
		})
		It("succeeds when volume  name exists and ubiquityClient does not return an error", func() {
			fakeClient.RemoveVolumeReturns(nil)
			objectMeta := metav1.ObjectMeta{Name: "vol1"}
			volume := v1.PersistentVolume{ObjectMeta: objectMeta}
			err = provisioner.Delete(&volume)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeClient.RemoveVolumeCallCount()).To(Equal(1))
		})
		It("succeeds when volume wasnot found in ubiquity DB", func() {
			fakeClient.GetVolumeReturns(resources.Volume{}, &resources.VolumeNotFoundError{"vol1"})
			objectMeta := metav1.ObjectMeta{Name: "vol1"}
			volume := v1.PersistentVolume{ObjectMeta: objectMeta}
			err = provisioner.Delete(&volume)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeClient.GetVolumeCallCount()).To(Equal(1))
			Expect(fakeClient.RemoveVolumeCallCount()).To(Equal(0))
		})

	})
})

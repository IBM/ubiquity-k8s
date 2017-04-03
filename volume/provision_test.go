package volume_test

import (
	"fmt"

	"github.com/kubernetes-incubator/external-storage/lib/controller"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/IBM/ubiquity-k8s/volume"
	"github.com/IBM/ubiquity/fakes"
)

var _ = Describe("Provisioner", func() {
	var (
		fakeClient *fakes.FakeStorageClient
		// fakeKubeInterface *k8s_fake.FakeInterface
		provisioner controller.Provisioner
		options     controller.VolumeOptions
		err         error
	)

	BeforeEach(func() {
		fakeClient = new(fakes.FakeStorageClient)
		// fakeKubeInterface = new(k8s_fake.FakeInterface)
		provisioner, err = volume.NewFlexProvisioner(testLogger, fakeClient, "/tmp")
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
			objectMeta := v1.ObjectMeta{Name: "vol1"}
			volume := v1.PersistentVolume{ObjectMeta: objectMeta}
			err = provisioner.Delete(&volume)
			Expect(err).To(HaveOccurred())
			Expect(fakeClient.RemoveVolumeCallCount()).To(Equal(1))
		})
		It("succeeds when volume  name exists and ubiquityClient does not return an error", func() {
			fakeClient.RemoveVolumeReturns(nil)
			objectMeta := v1.ObjectMeta{Name: "vol1"}
			volume := v1.PersistentVolume{ObjectMeta: objectMeta}
			err = provisioner.Delete(&volume)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeClient.RemoveVolumeCallCount()).To(Equal(1))
		})

	})
})

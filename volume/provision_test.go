package volume_test

import (
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.ibm.com/almaden-containers/ubiquity-k8s/volume"
	"github.ibm.com/almaden-containers/ubiquity/fakes"
)

var _ = Describe("Provisioner", func() {
	Context(".Provision", func() {
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
			provisioner, err = volume.NewFlexProvisioner(testLogger, fakeClient)
			options = controller.VolumeOptions{PVName: "fakepv"}
		})
		It("fails when options does not contain pvc", func() {
			_, err = provisioner.Provision(options)
			Expect(err).To(HaveOccurred())
		})
	})
})

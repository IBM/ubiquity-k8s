package flex

import (
	"context"
	"os"
	"time"

	gomock "github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	fakekubeclientset "k8s.io/client-go/kubernetes/fake"
	testcore "k8s.io/client-go/testing"

	flexmocks "github.com/IBM/ubiquity-k8s/sidecars/flex/mocks"
	"github.com/IBM/ubiquity-k8s/utils"
	"github.com/IBM/ubiquity/resources"
)

var _ = Describe("ServiceSyncer", func() {

	var ss *ServiceSyncer
	var kubeClient *fakekubeclientset.Clientset
	var realFlexConfigSyncer FlexConfigSyncer
	var ctx context.Context
	var cancelFunc context.CancelFunc
	var mockCtrl *gomock.Controller
	var mockFlexConfigSyncer *flexmocks.MockFlexConfigSyncer
	var ns = "ubiquity"
	var svc = &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      utils.UbiquityServiceName,
		},
		Spec: v1.ServiceSpec{
			ClusterIP: "1.2.3.4",
		},
	}

	BeforeEach(func() {

		os.Setenv("NAMESPACE", "ubiquity")
		ctx, cancelFunc = context.WithCancel(context.Background())
		kubeClient = fakekubeclientset.NewSimpleClientset()

		mockCtrl = gomock.NewController(GinkgoT())
		mockFlexConfigSyncer = flexmocks.NewMockFlexConfigSyncer(mockCtrl)
		realFlexConfigSyncer = defaultFlexConfigSyncer
		// mock the defaultFlexConfigSyncer
		defaultFlexConfigSyncer = mockFlexConfigSyncer

		ss, _ = NewServiceSyncer(kubeClient, ctx)
	})

	AfterEach(func() {
		os.Setenv("NAMESPACE", "")
		mockCtrl.Finish()
		defaultFlexConfigSyncer = realFlexConfigSyncer
	})

	Describe("test Sync", func() {

		JustBeforeEach(func() {
			go func() {
				// stop the Sync
				time.Sleep(30 * time.Millisecond)
				cancelFunc()
			}()
		})

		Context("ubiquity service does not exist at the beginning", func() {

			BeforeEach(func() {
				emptyConfig := &resources.UbiquityPluginConfig{}
				mockFlexConfigSyncer.EXPECT().GetCurrentFlexConfig().Return(emptyConfig, nil)
				mockFlexConfigSyncer.EXPECT().UpdateFlexConfig(gomock.Any())

				go func() {
					svcWatcher := watch.NewFake()
					kubeClient.PrependWatchReactor("services", testcore.DefaultWatchReactor(svcWatcher, nil))

					time.Sleep(20 * time.Millisecond)
					// create the service after the Sync starts
					kubeClient.CoreV1().Services(svc.Namespace).Create(svc)
					svcWatcher.Add(svc)
				}()
			})

			It("should call processService only once to get and update clusterIP", func(done Done) {
				err := ss.Sync()
				Ω(err).ShouldNot(HaveOccurred())
				close(done)
			})
		})

		Context("ubiquity service exists at the beginning and config has right cluterIP", func() {

			BeforeEach(func() {
				kubeClient.CoreV1().Services(svc.Namespace).Create(svc)
				configWithUbiquityIP := &resources.UbiquityPluginConfig{UbiquityServer: resources.UbiquityServerConnectionInfo{Address: "1.2.3.4"}}
				mockFlexConfigSyncer.EXPECT().GetCurrentFlexConfig().Return(configWithUbiquityIP, nil)
				mockFlexConfigSyncer.EXPECT().UpdateFlexConfig(gomock.Any()).Times(0)
			})

			It("should call processService only once but never update clusterIP", func(done Done) {
				err := ss.Sync()
				Ω(err).ShouldNot(HaveOccurred())
				close(done)
			})
		})

		Context("ubiquity service exists at the beginning and cluterIP never changes and config has no cluterIP", func() {

			BeforeEach(func() {
				kubeClient.CoreV1().Services(svc.Namespace).Create(svc)
				emptyConfig := &resources.UbiquityPluginConfig{}
				configWithUbiquityIP := &resources.UbiquityPluginConfig{UbiquityServer: resources.UbiquityServerConnectionInfo{Address: "1.2.3.4"}}
				mockFlexConfigSyncer.EXPECT().GetCurrentFlexConfig().Return(emptyConfig, nil)
				mockFlexConfigSyncer.EXPECT().GetCurrentFlexConfig().Return(configWithUbiquityIP, nil)
				mockFlexConfigSyncer.EXPECT().UpdateFlexConfig(gomock.Any())

				go func() {
					svcWatcher := watch.NewFake()
					kubeClient.PrependWatchReactor("services", testcore.DefaultWatchReactor(svcWatcher, nil))

					time.Sleep(20 * time.Millisecond)
					svcWatcher.Modify(svc)
				}()
			})

			It("should call processService twice but update clusterIP only once", func(done Done) {
				err := ss.Sync()
				Ω(err).ShouldNot(HaveOccurred())
				close(done)
			})
		})

		Context("ubiquity service exists at the beginning and cluterIP changes and config has right cluterIP", func() {

			BeforeEach(func() {
				kubeClient.CoreV1().Services(svc.Namespace).Create(svc)
				configWithUbiquityIP := &resources.UbiquityPluginConfig{UbiquityServer: resources.UbiquityServerConnectionInfo{Address: "1.2.3.4"}}
				mockFlexConfigSyncer.EXPECT().GetCurrentFlexConfig().Return(configWithUbiquityIP, nil).Times(2)
				mockFlexConfigSyncer.EXPECT().UpdateFlexConfig(gomock.Any())

				go func() {
					newSvc := svc.DeepCopy()
					newSvc.Spec.ClusterIP = "5.6.7.8"
					svcWatcher := watch.NewFake()
					kubeClient.PrependWatchReactor("services", testcore.DefaultWatchReactor(svcWatcher, nil))

					time.Sleep(20 * time.Millisecond)
					svcWatcher.Modify(newSvc)
				}()
			})

			It("should call processService twice but update clusterIP only once", func(done Done) {
				err := ss.Sync()
				Ω(err).ShouldNot(HaveOccurred())
				close(done)
			})
		})
	})
})

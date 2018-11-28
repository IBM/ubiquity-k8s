package hookexecutor

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	_ "github.com/IBM/ubiquity-k8s/cmd/hook-executor/logger"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakekubeclientset "k8s.io/client-go/kubernetes/fake"
)

var test_svcYaml = `
apiVersion: v1
kind: Service
metadata:
  name: ubiquity
  namespace: ubiquity
spec:
  clusterIP: 6.6.6.6
  ports:
  - port: 5432
    protocol: TCP
    targetPort: 5432
  selector:
    app: ubiquity
  type: ClusterIP
status:
  loadBalancer: {}
`

var test_daemonYaml = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
  name: ubiquity-k8s-flex
  namespace: ubiquity
spec:
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      name: ubiquity-k8s-flex
      product: ibm-storage-enabler-for-containers
  template:
    metadata:
      labels:
        name: ubiquity-k8s-flex
        product: ibm-storage-enabler-for-containers
    spec:
      containers:
      - command:
        - ./setup_flex.sh
        env:
        - name: UBIQUITY_IP_ADDRESS
          value: 0.0.0.0
        image: ibmcom/ibm-storage-flex-volume-for-kubernetes:1.2.0
        imagePullPolicy: IfNotPresent
        name: ubiquity-k8s-flex
        volumeMounts:
        - mountPath: /usr/libexec/kubernetes/kubelet-plugins/volume/exec
          name: host-k8splugindir
        - mountPath: /var/log
          name: flex-log-dir
      volumes:
      - hostPath:
          path: /usr/libexec/kubernetes/kubelet-plugins/volume/exec
          type: ""
        name: host-k8splugindir
      - hostPath:
          path: /var/log
          type: ""
        name: flex-log-dir
`

var _ = Describe("PostInstall", func() {

	var e Executor
	var kubeClient *fakekubeclientset.Clientset
	var svc *v1.Service
	var daemon *appsv1.DaemonSet

	BeforeEach(func() {

		os.Setenv("NAMESPACE", "ubiquity")

		svcObj, _ := FromYaml([]byte(test_svcYaml))
		svc = svcObj.(*v1.Service)

		daemonObj, _ := FromYaml([]byte(test_daemonYaml))
		daemon = daemonObj.(*appsv1.DaemonSet)

		kubeClient = fakekubeclientset.NewSimpleClientset(svc, daemon)

		e = PostInstallExecutor(kubeClient)
	})

	AfterEach(func() {
		os.Setenv("NAMESPACE", "")
	})

	Describe("test Execute", func() {

		Context("get Ubiquity serviceIP", func() {

			It("should return svc clusterIP", func() {
				Expect(e.(*postInstallExecutor).getUbiquityServiceIP()).To(Equal(svc.Spec.ClusterIP))
			})
		})

		Context("raise error if namespace not set", func() {

			BeforeEach(func() {
				os.Setenv("NAMESPACE", "")
			})

			It("should raise an error", func() {
				err := e.Execute()
				Ω(err).Should(HaveOccurred())
				Expect(err.Error()).To(Equal(ENVNamespaceNotSet))
			})
		})

		Context("raise error if Ubiquity serviceIP is empty", func() {

			BeforeEach(func() {
				newSvc := svc.DeepCopy()
				newSvc.Spec.ClusterIP = ""
				kubeClient.CoreV1().Services(newSvc.Namespace).Update(newSvc)
			})

			It("should raise an error", func() {
				err := e.Execute()
				Ω(err).Should(HaveOccurred())
				Expect(err.Error()).To(Equal(UbiquityServiceIPEmptyErrorStr))
			})
		})

		Context("update flex DaemonSet", func() {

			BeforeEach(func() {

				// IP before update is 0.0.0.0
				daemon, err := kubeClient.AppsV1().DaemonSets(daemon.Namespace).Get(daemon.Name, metav1.GetOptions{})
				Ω(err).ShouldNot(HaveOccurred())
				envs := daemon.Spec.Template.Spec.Containers[0].Env
				found := false
				for _, env := range envs {
					if env.Name == ubiquityIPAddressKey {
						Expect(env.Value).To(Equal("0.0.0.0"))
						found = true
					}
				}
				Expect(found).To(BeTrue())

				e.Execute()
			})

			It("should update ubiquityIPAddress to svc's clusterIP", func() {
				newDaemon, err := kubeClient.AppsV1().DaemonSets(daemon.Namespace).Get(daemon.Name, metav1.GetOptions{})
				Ω(err).ShouldNot(HaveOccurred())
				envs := newDaemon.Spec.Template.Spec.Containers[0].Env
				found := false
				for _, env := range envs {
					if env.Name == ubiquityIPAddressKey {
						Expect(env.Value).To(Equal(svc.Spec.ClusterIP))
						found = true
					}
				}
				Expect(found).To(BeTrue())
			})
		})
	})
})

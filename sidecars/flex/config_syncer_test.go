package flex

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var test_flexConfig = `
# This file was generated automatically by the ubiquity-k8s-flex Pod.

LogPath = "/usr/libexec/kubernetes/kubelet-plugins/volume/exec/ibm~ubiquity-k8s-flex"
LogRotateMaxSize = 0
Backends = ["scbe"]
LogLevel = "debug"

[DockerPlugin]
  Port = 0
  PluginsDirectory = ""

[UbiquityServer]
  Address = "1.2.3.4"
  Port = 9999

[SpectrumNfsRemoteConfig]
  ClientConfig = ""

[ScbeRemoteConfig]
  SkipRescanISCSI = false

[CredentialInfo]
  UserName = "ubiquity"
  Password = "ubiquity"
  Group = ""

[SslConfig]
  UseSsl = true
  SslMode = "require"
  VerifyCa = "/usr/libexec/kubernetes/kubelet-plugins/volume/exec/ibm~ubiquity-k8s-flex/ubiquity-trusted-ca.crt"
`

var _ = Describe("FlexConfigSyncer", func() {

	var realConfigFile string
	var tmpConfigFile *os.File

	BeforeEach(func() {
		var err error
		tmpConfigFile, err = ioutil.TempFile("", "")
		Ω(err).ShouldNot(HaveOccurred())

		realConfigFile = flexConfPath
		flexConfPath = tmpConfigFile.Name()

		tmpConfigFile.WriteString(test_flexConfig)
		err = tmpConfigFile.Close()
		Ω(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		flexConfPath = realConfigFile
		err := os.Remove(tmpConfigFile.Name())
		Ω(err).ShouldNot(HaveOccurred())

	})

	Context("test GetCurrentFlexConfig", func() {
		It("should be successful", func() {
			conf, err := defaultFlexConfigSyncer.GetCurrentFlexConfig()
			Ω(err).ShouldNot(HaveOccurred())
			Expect(conf.UbiquityServer.Address).To(Equal("1.2.3.4"))
		})
	})

	Context("test UpdateFlexConfig", func() {

		BeforeEach(func() {
			conf, err := defaultFlexConfigSyncer.GetCurrentFlexConfig()
			Ω(err).ShouldNot(HaveOccurred())
			conf.UbiquityServer.Address = "5.6.7.8"
			err = defaultFlexConfigSyncer.UpdateFlexConfig(conf)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should be updated for both cache and file", func() {
			// cache is updated
			Expect(defaultFlexConfigSyncer.(*flexConfigSyncer).cachedConfig.UbiquityServer.Address).To(Equal("5.6.7.8"))

			// reset the cached config
			defaultFlexConfigSyncer.(*flexConfigSyncer).cachedConfig = nil
			// get from file again
			newConf, err := defaultFlexConfigSyncer.GetCurrentFlexConfig()
			Ω(err).ShouldNot(HaveOccurred())
			// file is updated
			Expect(newConf.UbiquityServer.Address).To(Equal("5.6.7.8"))
		})
	})
})

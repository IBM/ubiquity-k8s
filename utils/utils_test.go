package utils_test

import (
	. "github.com/IBM/ubiquity-k8s/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"strconv"
)

var _ = Describe("Utils", func() {
	var (
		tests1 = []struct {
			envVar string
			in     string
			out    string
		}{
			{"LOG_LEVEL", "debug", "debug"},
			{"FLEX_LOG_ROTATE_MAXSIZE", "100", "100"},
			{"LOG_PATH", "/var/log", "/var/log"},
			{"BACKENDS", "SCBE", "SCBE"},
			{"UBIQUITY_PORT", "9999", "9999"},
			{"UBIQUITY_ADDRESS", "9.9.9.9", "9.9.9.9"},
			{"UBIQUITY_USERNAME", "ubiquity", "ubiquity"},
			{"UBIQUITY_PASSWORD", "ubiquity", "ubiquity"},
		}

		tests2 = []struct {
			envVar string
			in     string
			out    string
		}{
			{"LOG_LEVEL", "debug", "debug"},
			{"FLEX_LOG_ROTATE_MAXSIZE", "", "50"},
			{"LOG_PATH", "/var/log", "/var/log"},
			{"BACKENDS", "SCBE", "SCBE"},
			{"UBIQUITY_PORT", "9999", "9999"},
			{"UBIQUITY_ADDRESS", "9.9.9.9", "9.9.9.9"},
			{"UBIQUITY_USERNAME", "ubiquity", "ubiquity"},
			{"UBIQUITY_PASSWORD", "ubiquity", "ubiquity"},
		}
	)

	Context("set the correct env val", func() {
		BeforeEach(func() {
			for _, data := range tests1 {
				os.Setenv(data.envVar, data.in)
			}
		})
		It("verify the vals which get from LoadConfig are expected", func() {
			ubiquityConfig, err := LoadConfig()
			Expect(err).To(Not(HaveOccurred()))
			loadData := []struct {
				envVar string
				val    string
			}{

				{"LOG_LEVEL", ubiquityConfig.LogLevel},
				{"FLEX_LOG_ROTATE_MAXSIZE", strconv.Itoa(ubiquityConfig.LogRotateMaxSize)},
				{"LOG_PATH", ubiquityConfig.LogPath},
				{"BACKENDS", ubiquityConfig.Backends[0]},
				{"UBIQUITY_PORT", strconv.Itoa(ubiquityConfig.UbiquityServer.Port)},
				{"UBIQUITY_ADDRESS", ubiquityConfig.UbiquityServer.Address},
				{"UBIQUITY_USERNAME", ubiquityConfig.CredentialInfo.UserName},
				{"UBIQUITY_PASSWORD", ubiquityConfig.CredentialInfo.Password},
			}
			for _, data := range tests1 {
				for _, loaddata := range loadData {
					if loaddata.envVar == data.envVar {
						Expect(loaddata.val).To(Equal(data.out))
					}
				}
			}
		})

	})

	Context("Without set FLEX_LOG_ROTATE_MAXSIZE", func() {
		BeforeEach(func() {
			for _, data := range tests2 {
				os.Setenv(data.envVar, data.in)
			}
		})
		It("verify the vals which get from LoadConfig are expected", func() {
			ubiquityConfig, err := LoadConfig()
			Expect(err).To(Not(HaveOccurred()))
			loadData := []struct {
				envVar string
				val    string
			}{

				{"LOG_LEVEL", ubiquityConfig.LogLevel},
				{"FLEX_LOG_ROTATE_MAXSIZE", strconv.Itoa(ubiquityConfig.LogRotateMaxSize)},
				{"LOG_PATH", ubiquityConfig.LogPath},
				{"BACKENDS", ubiquityConfig.Backends[0]},
				{"UBIQUITY_PORT", strconv.Itoa(ubiquityConfig.UbiquityServer.Port)},
				{"UBIQUITY_ADDRESS", ubiquityConfig.UbiquityServer.Address},
				{"UBIQUITY_USERNAME", ubiquityConfig.CredentialInfo.UserName},
				{"UBIQUITY_PASSWORD", ubiquityConfig.CredentialInfo.Password},
			}
			for _, data := range tests2 {
				for _, loaddata := range loadData {
					if loaddata.envVar == data.envVar {
						Expect(loaddata.val).To(Equal(data.out))
					}
				}
			}
		})

	})
})

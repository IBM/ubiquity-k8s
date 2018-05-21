package utils_test

import (
	"testing"
	"os"
	k8sutils "github.com/IBM/ubiquity-k8s/utils"
	"strconv"
	"fmt"
)

var tests1 = []struct{
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
	{"SCBE_SKIP_RESCAN_ISCSI", "true", "true"},
	{"UBIQUITY_USERNAME", "ubiquity", "ubiquity"},
	{"UBIQUITY_PASSWORD", "ubiquity", "ubiquity"},
}

var tests2 = []struct{
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
	{"SCBE_SKIP_RESCAN_ISCSI", "", "false"},
	{"UBIQUITY_USERNAME", "ubiquity", "ubiquity"},
	{"UBIQUITY_PASSWORD", "ubiquity", "ubiquity"},

}

//Test with correct config tests1
func TestUtils1(t *testing.T) {
	for _, data := range tests1 {
		os.Setenv(data.envVar, data.in)
	}

	ubiquityConfig, err := k8sutils.LoadConfig()
	if err != nil {
		fmt.Println("LoadConfig failed: ",err)
	}
	loadData := []struct{
		envVar string
		val    string
	}{

		{"LOG_LEVEL", ubiquityConfig.LogLevel},
		{"FLEX_LOG_ROTATE_MAXSIZE", strconv.Itoa(ubiquityConfig.LogRotateMaxSize)},
		{"LOG_PATH", ubiquityConfig.LogPath},
		{"BACKENDS", ubiquityConfig.Backends[0]},
		{"UBIQUITY_PORT", strconv.Itoa(ubiquityConfig.UbiquityServer.Port)},
		{"UBIQUITY_ADDRESS", ubiquityConfig.UbiquityServer.Address},
		{"SCBE_SKIP_RESCAN_ISCSI", strconv.FormatBool(ubiquityConfig.ScbeRemoteConfig.SkipRescanISCSI)},
		{"UBIQUITY_USERNAME", ubiquityConfig.CredentialInfo.UserName},
		{"UBIQUITY_PASSWORD",ubiquityConfig.CredentialInfo.Password},
	}
	for _, data := range tests1 {
		for _, loaddata := range loadData {
			if loaddata.envVar == data.envVar {
				if loaddata.val == data.out {
					t.Logf("Load env %s successful with val %s", loaddata.envVar, loaddata.val)
				} else {
					t.Logf("Load env %s fail with val %s", loaddata.envVar, loaddata.val)
					t.Fail()
				}
			}
		}
	}
}

//Test with special config tests2
//Don't assign val of FLEX_LOG_ROTATE_MAXSIZE and SCBE_SKIP_RESCAN_ISCSI
func TestUtils2(t *testing.T) {
	for _, data := range tests1 {
		os.Setenv(data.envVar, data.in)
	}

	ubiquityConfig, err := k8sutils.LoadConfig()
	if err != nil {
		fmt.Println("LoadConfig failed: ",err)
	}
	loadData := []struct{
		envVar string
		val    string
	}{

		{"LOG_LEVEL", ubiquityConfig.LogLevel},
		{"FLEX_LOG_ROTATE_MAXSIZE", strconv.Itoa(ubiquityConfig.LogRotateMaxSize)},
		{"LOG_PATH", ubiquityConfig.LogPath},
		{"BACKENDS", ubiquityConfig.Backends[0]},
		{"UBIQUITY_PORT", strconv.Itoa(ubiquityConfig.UbiquityServer.Port)},
		{"UBIQUITY_ADDRESS", ubiquityConfig.UbiquityServer.Address},
		{"SCBE_SKIP_RESCAN_ISCSI", strconv.FormatBool(ubiquityConfig.ScbeRemoteConfig.SkipRescanISCSI)},
		{"UBIQUITY_USERNAME", ubiquityConfig.CredentialInfo.UserName},
		{"UBIQUITY_PASSWORD",ubiquityConfig.CredentialInfo.Password},
	}
	for _, data := range tests1 {
		for _, loaddata := range loadData {
			if loaddata.envVar == data.envVar {
				if loaddata.val == data.out {
					t.Logf("Load env %s successful with val %s", loaddata.envVar, loaddata.val)
				} else {
					t.Logf("Load env %s fail with val %s", loaddata.envVar, loaddata.val)
					t.Fail()
				}
			}
		}
	}
}




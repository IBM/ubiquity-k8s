package utils

import (
	"os"
	"strconv"
	"strings"

	"github.com/IBM/ubiquity/resources"
)

func LoadConfig() (resources.UbiquityPluginConfig, error) {

	config := resources.UbiquityPluginConfig{}
	config.LogLevel = os.Getenv("LOG_LEVEL")
	LogRotateMaxSize, err := strconv.Atoi(os.Getenv("FLEX_LOG_ROTATE_MAXSIZE"))
	if err != nil {
		if LogRotateMaxSize == 0 {
			LogRotateMaxSize = 50
		} else {
			return config, err
		}
	}
	config.LogRotateMaxSize = LogRotateMaxSize
	config.LogPath = os.Getenv("LOG_PATH")
	config.Backends = strings.Split(os.Getenv("BACKENDS"), ",")
	ubiquity := resources.UbiquityServerConnectionInfo{}
	port, err := strconv.ParseInt(os.Getenv("UBIQUITY_PORT"), 0, 32)
	if err != nil {
		return config, err
	}
	ubiquity.Port = int(port)
	ubiquity.Address = os.Getenv("UBIQUITY_ADDRESS")
	config.UbiquityServer = ubiquity

	spectrumNFSConfig := resources.SpectrumNfsRemoteConfig{}
	spectrumNFSConfig.ClientConfig = os.Getenv("SPECTRUM_NFS_REMOTE_CONFIG")
	config.SpectrumNfsRemoteConfig = spectrumNFSConfig

	bool, err := strconv.ParseBool(os.Getenv("SCBE_SKIP_RESCAN_ISCSI"))
	if err != nil {
		config.ScbeRemoteConfig.SkipRescanISCSI = false
	} else {
		config.ScbeRemoteConfig.SkipRescanISCSI = bool
	}

	config.CredentialInfo = resources.CredentialInfo{UserName: os.Getenv("UBIQUITY_USERNAME"), Password: os.Getenv("UBIQUITY_PASSWORD")}

	return config, nil
}

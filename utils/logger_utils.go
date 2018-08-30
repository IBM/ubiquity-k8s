package utils

import (
	"github.com/IBM/ubiquity/utils/logs"
	k8sresources "github.com/IBM/ubiquity-k8s/resources"
	"github.com/IBM/ubiquity/resources"
	"path"
)

func InitFlexLogger(config resources.UbiquityPluginConfig) func(){
	var logger_params = logs.LoggerParams{ShowGoid: false, ShowPid : true}
	deferFunction :=  logs.InitFileLogger(logs.GetLogLevelFromString(config.LogLevel), path.Join(config.LogPath, k8sresources.UbiquityFlexLogFileName), config.LogRotateMaxSize, logger_params)
	return deferFunction
}

func InitProvisionerLogger(ubiquityConfig resources.UbiquityPluginConfig) func(){
	deferFunction :=  logs.InitStdoutLogger(logs.GetLogLevelFromString(ubiquityConfig.LogLevel), logs.LoggerParams{ShowGoid: false, ShowPid : false})
	return deferFunction
}



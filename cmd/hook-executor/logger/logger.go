package logger

import (
	"fmt"
	"os"

	k8sutils "github.com/IBM/ubiquity-k8s/utils"
)

const (
	defaultlogPath  = "/var/log"
	defaultlogLevel = "info"
)

func init() {
	initLogger()
}

func initLogger() {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = defaultlogLevel
	}
	logPath := os.Getenv("LOG_PATH")
	if logPath == "" {
		logPath = defaultlogPath
	}

	err := os.MkdirAll(logPath, 0640)
	if err != nil {
		panic(fmt.Errorf("Failed to setup log dir"))
	}

	//logger := utils.SetupOldLogger(k8sresources.HookExecutorName)
	k8sutils.InitGenericLogger(logLevel)
}

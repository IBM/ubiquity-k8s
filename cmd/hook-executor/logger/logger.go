package logger

import (
	"os"

	k8sutils "github.com/IBM/ubiquity-k8s/utils"
)

const (
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
	k8sutils.InitGenericLogger(logLevel)
}

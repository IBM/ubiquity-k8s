package config_helpers

import (
	"os"
	"path/filepath"
	"runtime"
)

var BluemixTmpDir = bluemixTmpDir()

func bluemixTmpDir() string {
	d := filepath.Join(ConfigDir(), "tmp")
	os.MkdirAll(d, 0755)
	return d
}

func ConfigDir() string {
	if os.Getenv("BLUEMIX_HOME") != "" {
		return filepath.Join(os.Getenv("BLUEMIX_HOME"), ".bluemix")
	} else {
		return filepath.Join(UserHomeDir(), ".bluemix")
	}
}

func ConfigFilePath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

func PluginRepoDir() string {
	return filepath.Join(ConfigDir(), "plugins")
}

func PluginsConfigFilePath() string {
	return filepath.Join(PluginRepoDir(), "config.json")
}

func PluginDir(pluginName string) string {
	return filepath.Join(PluginRepoDir(), pluginName)
}

func PluginBinaryLocation(pluginName string) string {
	executable := filepath.Join(PluginDir(pluginName), pluginName)
	if runtime.GOOS == "windows" {
		executable = executable + ".exe"
	}
	return executable
}

func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}

	return os.Getenv("HOME")
}

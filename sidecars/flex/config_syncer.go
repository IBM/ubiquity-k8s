package flex

import (
	"os"

	"github.com/BurntSushi/toml"

	k8sresources "github.com/IBM/ubiquity-k8s/resources"
	"github.com/IBM/ubiquity/resources"
)

var cachedConfig *resources.UbiquityPluginConfig

func getCurrentFlexConfig() (*resources.UbiquityPluginConfig, error) {
	if cachedConfig == nil {
		cachedConfig = &resources.UbiquityPluginConfig{}
		if _, err := toml.DecodeFile(k8sresources.FlexConfPath, cachedConfig); err != nil {
			return nil, err
		}
	}
	return cachedConfig, nil
}

func updateFlexConfig(newConfig *resources.UbiquityPluginConfig) error {
	f, err := os.OpenFile(k8sresources.FlexConfPath, os.O_WRONLY, os.FileMode(0755))
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	f.WriteString("# This file was generated automatically by the ubiquity-k8s-flex Pod.\n\n")
	f.Sync()

	encoder := toml.NewEncoder(f)
	err = encoder.Encode(*newConfig)
	if err != nil {
		return err
	}
	return nil
}

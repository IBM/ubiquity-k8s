package flex

import (
	"os"

	"github.com/BurntSushi/toml"

	k8sresources "github.com/IBM/ubiquity-k8s/resources"
	"github.com/IBM/ubiquity/resources"
)

var flexConfPath = k8sresources.FlexConfPath

type FlexConfigSyncer interface {
	GetCurrentFlexConfig() (*resources.UbiquityPluginConfig, error)
	UpdateFlexConfig(newConfig *resources.UbiquityPluginConfig) error
}

type flexConfigSyncer struct {
	cachedConfig *resources.UbiquityPluginConfig
}

func (s *flexConfigSyncer) GetCurrentFlexConfig() (*resources.UbiquityPluginConfig, error) {
	if s.cachedConfig == nil {
		s.cachedConfig = &resources.UbiquityPluginConfig{}
		if _, err := toml.DecodeFile(flexConfPath, s.cachedConfig); err != nil {
			s.cachedConfig = nil
			return nil, err
		}
	}
	return s.cachedConfig, nil
}

func (s *flexConfigSyncer) UpdateFlexConfig(newConfig *resources.UbiquityPluginConfig) error {
	f, err := os.OpenFile(flexConfPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.FileMode(0755))
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	// write the file header first.
	f.WriteString("# This file was generated automatically by the ubiquity-k8s-flex Pod.\n\n")
	f.Sync()

	encoder := toml.NewEncoder(f)
	err = encoder.Encode(*newConfig)

	if err != nil {
		return err
	}
	return nil
}

var defaultFlexConfigSyncer FlexConfigSyncer = &flexConfigSyncer{}

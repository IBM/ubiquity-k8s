package configuration

import (
	cfconfiguration "github.com/cloudfoundry/cli/cf/configuration"

	"github.ibm.com/Bluemix/bluemix-cli-sdk/bluemix/configuration/core_config"
)

type FakeCFPersistor struct{}

func (f *FakeCFPersistor) Save(cfconfiguration.DataInterface) error { return nil }
func (f *FakeCFPersistor) Load(cfconfiguration.DataInterface) error { return nil }
func (f *FakeCFPersistor) Delete()                                  {}
func (f *FakeCFPersistor) Exists() bool                             { return false }

type FakeBXPersistor struct{}

func (f *FakeBXPersistor) Save(interface{}) error { return nil }
func (f *FakeBXPersistor) Load(interface{}) error { return nil }

func NewFakeCoreConfig() core_config.ReadWriter {
	config := core_config.NewCoreConfigFromPersistor(new(FakeCFPersistor), new(FakeBXPersistor), func(err error) { panic(err) })
	config.SetAPIVersion("3")
	return config
}

package configuration

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.ibm.com/Bluemix/bluemix-cli-sdk/common/file_helpers"
)

type Persistor interface {
	Save(interface{}) error
	Load(interface{}) error
}

type DiskPersistor struct {
	filePath string
}

func NewDiskPersistor(path string) *DiskPersistor {
	return &DiskPersistor{
		filePath: path,
	}
}

func (p *DiskPersistor) Load(retVal interface{}) (err error) {
	jsonBytes, err := ioutil.ReadFile(p.filePath)
	if err != nil {
		return
	}
	err = json.Unmarshal(jsonBytes, retVal)
	return
}

func (p *DiskPersistor) Save(data interface{}) (err error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return
	}

	if !file_helpers.FileExists(p.filePath) {
		err = os.MkdirAll(filepath.Dir(p.filePath), 0700)
		if err != nil {
			return
		}
	}

	return ioutil.WriteFile(p.filePath, jsonBytes, 0600)
}

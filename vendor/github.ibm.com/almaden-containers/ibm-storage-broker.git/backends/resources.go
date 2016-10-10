package backends

import (
	"github.ibm.com/almaden-containers/ibm-storage-broker.git/model"
)

//go:generate counterfeiter -o ../fakes/fake_storage_backend.go . StorageBackend
type StorageBackend interface {

	//CreateWithoutProvisioning(name string, opts map[string]interface{}) error
	//ExportNfs(name string, clientCIDR string) (string, error)
	//UnexportNfs(name string) error
	//RemoveWithoutDeletingVolume(string) error
	//GetFileSetForMountPoint(mountPoint string) (string, error)

	Activate() error
	CreateVolume(name string, opts map[string]interface{}) error
	RemoveVolume(name string, forceDelete bool) error
	ListVolumes() ([]model.VolumeMetadata, error)
	GetVolume(name string) (volumeMetadata model.VolumeMetadata, volumeConfigDetails model.SpectrumConfig, err error)
	Attach(name string) (string, error)
	Detach(name string) error
	GetPluginName() string
}

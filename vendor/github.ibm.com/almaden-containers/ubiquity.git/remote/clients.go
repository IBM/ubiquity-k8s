package remote

import (
	"fmt"
	"log"

	"github.ibm.com/almaden-containers/ubiquity.git/model"
)

func NewRemoteClient(logger *log.Logger, storageApiURL string, backendName string) (model.StorageClient, error) {
	if backendName == "spectrum-scale" {
		return NewSpectrumRemoteClient(logger, storageApiURL), nil
	}
	if backendName == "spectrum-scale-nfs" {
		return NewNFSRemoteClient(logger, storageApiURL, backendName), nil
	}
	return nil, fmt.Errorf("Backend not found: " + backendName)
}

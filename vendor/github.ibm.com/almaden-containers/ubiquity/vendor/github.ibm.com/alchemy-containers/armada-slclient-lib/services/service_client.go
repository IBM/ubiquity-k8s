package services

import (
	"fmt"
	"github.ibm.com/alchemy-containers/armada-slclient-lib/client"
	"github.ibm.com/alchemy-containers/armada-slclient-lib/interfaces"
)

const (
	API_URL = "api.softlayer.com/rest/v3"
)

type ServiceClient struct {
	softLayerStorageService interfaces.Softlayer_Storage_Service
	httpclient              interfaces.HttpClient
}

func NewServiceClientWithParam(httpclient interfaces.HttpClient) *ServiceClient {
	return &ServiceClient{
		httpclient: httpclient,
	}
}

func NewServiceClientWithAuth(username string, password string) *ServiceClient {
	httpclient := client.NewBasicHttpClient(username, password, API_URL)
	return &ServiceClient{
		httpclient: httpclient,
	}
}

func NewServiceClient() *ServiceClient {
	//Read details from ENV/Configuration file
	username := "dev-mon01-con.alchemy.dev"
	password := "6d9b159aff6459150a60f1d7dabff0ef26287f9b03a50310f36eb8708ac62cce"
	httpclient := client.NewBasicHttpClient(username, password, API_URL)
	return &ServiceClient{
		httpclient: httpclient,
	}
}

func (sc *ServiceClient) GetSoftLayerStorageService() interfaces.Softlayer_Storage_Service {
	if sc.softLayerStorageService == nil {
		//Create Softlayer storage service
		sc.softLayerStorageService = NewSoftLayer_File_Storage_Service(sc.httpclient)
		fmt.Printf("\nSoftlayer Storage Service creation is Successful\n")
	}
	return sc.softLayerStorageService
}

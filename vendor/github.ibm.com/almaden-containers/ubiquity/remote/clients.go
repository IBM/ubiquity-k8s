package remote

import (
	"fmt"
	"log"
	"github.ibm.com/almaden-containers/ubiquity/utils"
	"github.ibm.com/almaden-containers/ubiquity/model"
	"net/http"
)

var ClientDescriptors []model.StorageClientDescriptor = []model.StorageClientDescriptor{
	NfsRemoteClient,
	SpectrumRemoteClient,
}

func GetParams() []model.Parameter {
	params := make([]model.Parameter, 0, 20)
	for _, clientDescriptor := range ClientDescriptors {
		params = append(params, clientDescriptor.Params...)
	}
	return params
}

func NewRemoteClient(logger *log.Logger, backendName, storageApiURL string, paramValues map[string]interface{}) (model.StorageClient, error) {

	infoRemoteURL := utils.FormatURL(storageApiURL, backendName, "info")
	response, err := utils.HttpExecute(&http.Client{}, logger, "GET", infoRemoteURL, nil)
	if err != nil {
		logger.Printf("NewRemoteClient: Error in INFO remote call %#v", err)
		return nil, fmt.Errorf("Error in INFO remote call")
	}

	if response.StatusCode != http.StatusOK {
		logger.Printf("NewRemoteClient: Error in INFO remote call %#v", err)
		return nil, utils.ExtractErrorResponse(response)
	}

	infoResponse := model.InfoResponse{}
	err = utils.UnmarshalResponse(response, &infoResponse)
	if err != nil {
		logger.Printf("NewRemoteClient: Error in unmarshalling response for INFO remote call %#v for response %#v", err, response)
		return nil, fmt.Errorf("Error in INFO remote call")
	}

	remoteClientName := infoResponse.Info.RemoteClient

	for _, clientDescriptor := range ClientDescriptors {
		if clientDescriptor.Name == remoteClientName {

			for _, param := range clientDescriptor.Params{
				paramValue, ok := paramValues[param.Name].(*string)
				if !ok || (param.Required && (len(*paramValue) == 0)) {
					errorMsg := fmt.Sprintf("Missing required parameter %v for client %s / backend %s", param.Name, remoteClientName, backendName)
					logger.Printf("NewRemoteClient: Error: %s", errorMsg)
					return nil, fmt.Errorf(errorMsg)
				}
			}
			return clientDescriptor.New(logger, backendName, storageApiURL, paramValues)
		}
	}

	errorMsg := fmt.Sprintf("Missing required remote client %s for backend %s", remoteClientName, backendName)
	logger.Printf("NewRemoteClient: Error: %s", errorMsg)
	return nil, fmt.Errorf(errorMsg)
}

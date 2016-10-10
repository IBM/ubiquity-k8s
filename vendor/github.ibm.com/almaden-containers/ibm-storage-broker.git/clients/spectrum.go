package clients

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"os"
	"os/exec"

	"net/http"

	"github.ibm.com/almaden-containers/ibm-storage-broker.git/backends"
	"github.ibm.com/almaden-containers/ibm-storage-broker.git/model"

	"encoding/json"
	"path"

	"github.ibm.com/almaden-containers/ibm-storage-broker.git/utils"
)

type spectrumClient struct {
	logger        *log.Logger
	filesystem    string
	mountpoint    string
	isActivated   bool
	isMounted     bool
	HttpClient    *http.Client
	storageApiURL string
}

func NewSpectrumRemoteClient(logger *log.Logger, filesystem, mountpoint string, storageApiURL string) backends.StorageBackend {
	return &spectrumClient{logger: logger, filesystem: filesystem, mountpoint: mountpoint, storageApiURL: storageApiURL}
}
func (s *spectrumClient) Activate() error {
	s.logger.Println("MMCliFilesetClient: Activate start")
	defer s.logger.Println("MMCliFilesetClient: Activate end")

	if s.isActivated {
		return nil
	}

	//check if filesystem is mounted
	mounted, err := s.isSpectrumScaleMounted()

	if err != nil {
		s.logger.Println(err.Error())
		return err
	}

	if mounted == false {
		err = s.mount()

		if err != nil {
			s.logger.Println(err.Error())
			return err
		}
	}

	// call remote activate
	activateURL := path.Join(s.storageApiURL, s.GetPluginName(), "activate")
	request, err := http.NewRequest("POST", activateURL, nil)
	if err != nil {
		return fmt.Errorf("Error in creating activate request")
	}
	response, err := s.HttpClient.Do(request)
	if err != nil {
		return fmt.Errorf("Error in activate remote call")
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Error in activate remote call")
	}

	s.isActivated = true
	return nil
}

func (s *spectrumClient) GetPluginName() string {
	return "spectrum-scale"
}

func (s *spectrumClient) CreateVolume(name string, opts map[string]interface{}) (err error) {
	s.logger.Println("MMCliFilesetClient: create start")
	defer s.logger.Println("MMCliFilesetClient: create end")

	// call remote

	createRemoteURL := path.Join(s.storageApiURL, s.GetPluginName(), "volumes")
	createVolumeRequest := model.CreateRequest{Name: name, Opts: opts}

	bytesRequest, err := json.MarshalIndent(createVolumeRequest, "", " ")
	if err != nil {
		return fmt.Errorf("Internal error marshalling create params")
	}

	request, err := http.NewRequest("POST", createRemoteURL, bytes.NewReader(bytesRequest))
	if err != nil {
		return fmt.Errorf("Error in creating create request")
	}
	response, err := s.HttpClient.Do(request)
	if err != nil {
		return fmt.Errorf("Error in create volume remote call")
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Error in create volume remote call")
	}

	return nil
}

func (s *spectrumClient) RemoveVolume(name string, forceDelete bool) (err error) {
	s.logger.Println("MMCliFilesetClient: remove start")
	defer s.logger.Println("MMCliFilesetClient: remove end")

	removeRemoteURL := path.Join(s.storageApiURL, s.GetPluginName(), "volumes", name)

	removeRequest := model.RemoveRequest{Name: name, ForceDelete: forceDelete}

	bytesRequest, err := json.MarshalIndent(removeRequest, "", " ")
	if err != nil {
		return fmt.Errorf("Internal error marshalling remove params")
	}

	request, err := http.NewRequest("DELETE", removeRemoteURL, bytes.NewReader(bytesRequest))
	if err != nil {
		return fmt.Errorf("Error in creating remove request")
	}

	response, err := s.HttpClient.Do(request)
	if err != nil {
		return fmt.Errorf("Error in remove volume remote call")
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Error in remove volume remote call")
	}

	return nil
}

//GetVolume(string) (*model.VolumeMetadata, *string, *map[string]interface {}, error)
func (s *spectrumClient) GetVolume(name string) (model.VolumeMetadata, model.SpectrumConfig, error) {
	s.logger.Println("MMCliFilesetClient: get start")
	defer s.logger.Println("MMCliFilesetClient: get finish")

	getRemoteURL := path.Join(s.storageApiURL, s.GetPluginName(), "volumes", name)

	request, err := http.NewRequest("GET", getRemoteURL, nil)
	if err != nil {
		return model.VolumeMetadata{}, model.SpectrumConfig{}, fmt.Errorf("Error in creating get request")
	}
	response, err := s.HttpClient.Do(request)
	if err != nil {
		return model.VolumeMetadata{}, model.SpectrumConfig{}, fmt.Errorf("Error in get volume remote call")
	}

	if response.StatusCode != http.StatusOK {
		return model.VolumeMetadata{}, model.SpectrumConfig{}, fmt.Errorf("Error in get volume remote call")
	}

	getResponse := model.GetResponse{}
	err = utils.UnmarshalResponse(response, &getResponse)
	if err != nil {
		return model.VolumeMetadata{}, model.SpectrumConfig{}, fmt.Errorf("Error in unmarshalling response for get remote call")
	}

	return getResponse.Volume, getResponse.Config, nil
}

func (s *spectrumClient) Attach(name string) (string, error) {
	s.logger.Println("MMCliFilesetClient: attach start")
	defer s.logger.Println("MMCliFilesetClient: attach end")

	attachRemoteURL := path.Join(s.storageApiURL, s.GetPluginName(), "volumes", name, "attach")

	request, err := http.NewRequest("PUT", attachRemoteURL, nil)
	if err != nil {
		return "", fmt.Errorf("Error in creating attach request")
	}
	response, err := s.HttpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("Error in attach volume remote call")
	}

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Error in attach volume remote call")
	}

	mountResponse := model.MountResponse{}
	err = utils.UnmarshalResponse(response, &mountResponse)
	if err != nil {
		return "", fmt.Errorf("Error in unmarshalling response for attach remote call")
	}
	return mountResponse.Mountpoint, nil
}

func (s *spectrumClient) Detach(name string) error {
	s.logger.Println("MMCliFilesetClient: detach start")
	defer s.logger.Println("MMCliFilesetClient: detach end")

	detachRemoteURL := path.Join(s.storageApiURL, s.GetPluginName(), "volumes", name, "detach")

	request, err := http.NewRequest("PUT", detachRemoteURL, nil)
	if err != nil {
		return fmt.Errorf("Error in creating detach request")
	}
	response, err := s.HttpClient.Do(request)
	if err != nil {
		return fmt.Errorf("Error in detach volume remote call")
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Error in detach volume remote call")
	}

	return nil

}

func (s *spectrumClient) ListVolumes() ([]model.VolumeMetadata, error) {
	s.logger.Println("MMCliFilesetClient: list start")
	defer s.logger.Println("MMCliFilesetClient: list end")

	listRemoteURL := path.Join(s.storageApiURL, s.GetPluginName(), "volumes")

	request, err := http.NewRequest("GET", listRemoteURL, nil)
	if err != nil {
		return []model.VolumeMetadata{}, fmt.Errorf("Error in creating list request")
	}
	response, err := s.HttpClient.Do(request)
	if err != nil {
		return []model.VolumeMetadata{}, fmt.Errorf("Error in list volume remote call")
	}

	if response.StatusCode != http.StatusOK {
		return []model.VolumeMetadata{}, fmt.Errorf("Error in list volume remote call")
	}

	listResponse := model.ListResponse{}
	err = utils.UnmarshalResponse(response, &listResponse)
	if err != nil {
		return []model.VolumeMetadata{}, fmt.Errorf("Error in unmarshalling response for get remote call")
	}

	return listResponse.Volumes, nil

}

func (s *spectrumClient) mount() error {
	s.logger.Println("MMCliFilesetClient: mount start")
	defer s.logger.Println("MMCliFilesetClient: mount end")

	if s.isMounted == true {
		return nil
	}

	spectrumCommand := "/usr/lpp/mmfs/bin/mmmount"
	args := []string{s.filesystem, s.mountpoint}
	cmd := exec.Command(spectrumCommand, args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to mount filesystem")
	}
	s.logger.Println(output)
	s.isMounted = true
	return nil
}

func (s *spectrumClient) isSpectrumScaleMounted() (bool, error) {
	s.logger.Println("MMCliFilesetClient: isMounted start")
	defer s.logger.Println("MMCliFilesetClient: isMounted end")

	if s.isMounted == true {
		return s.isMounted, nil
	}

	spectrumCommand := "/usr/lpp/mmfs/bin/mmlsmount"
	args := []string{s.filesystem, "-L", "-Y"}
	cmd := exec.Command(spectrumCommand, args...)
	outputBytes, err := cmd.Output()
	if err != nil {
		s.logger.Printf("Error running command\n")
		s.logger.Println(err)
		return false, err
	}
	mountedNodes := extractMountedNodes(string(outputBytes))
	if len(mountedNodes) == 0 {
		//not mounted anywhere
		s.isMounted = false
		return s.isMounted, nil
	} else {
		// checkif mounted on current node -- compare node name
		currentNode, _ := os.Hostname()
		s.logger.Printf("MMCliFilesetClient: node name: %s\n", currentNode)
		for _, node := range mountedNodes {
			if node == currentNode {
				s.isMounted = true
				return s.isMounted, nil
			}
		}
	}
	s.isMounted = false
	return s.isMounted, nil
}

func extractMountedNodes(spectrumOutput string) []string {
	var nodes []string
	lines := strings.Split(spectrumOutput, "\n")
	if len(lines) == 1 {
		return nodes
	}
	for _, line := range lines[1:] {
		tokens := strings.Split(line, ":")
		if len(tokens) > 10 {
			if tokens[11] != "" {
				nodes = append(nodes, tokens[11])
			}
		}
	}
	return nodes
}

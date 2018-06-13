/**
 * Copyright 2017 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package volume

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	k8sresources "github.com/IBM/ubiquity-k8s/resources"
	"github.com/IBM/ubiquity/resources"
	"github.com/IBM/ubiquity/utils/logs"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"

	"k8s.io/api/core/v1"
)

const (

	// are we allowed to set this? else make up our own
	annCreatedBy = "kubernetes.io/createdby"
	createdBy    = k8sresources.UbiquityProvisionerName

	// Name of the file where an nfsProvisioner will store its identity
	identityFile = "k8sresources.UbiquityProvisionerName" + ".identity"

	// VolumeGidAnnotationKey is the key of the annotation on the PersistentVolume
	// object that specifies a supplemental GID.
	VolumeGidAnnotationKey = "pv.beta.kubernetes.io/gid"

	// A PV annotation for the identity of the flexProvisioner that provisioned it
	annProvisionerId = "Provisioner_Id"

	podIPEnv     = "POD_IP"
	serviceEnv   = "SERVICE_NAME"
	namespaceEnv = "POD_NAMESPACE"
	nodeEnv      = "NODE_NAME"
)

func NewFlexProvisioner(logger *log.Logger, ubiquityClient resources.StorageClient, config resources.UbiquityPluginConfig) (controller.Provisioner, error) {
	return newFlexProvisionerInternal(logger, ubiquityClient, config)
}

func newFlexProvisionerInternal(logger *log.Logger, ubiquityClient resources.StorageClient, config resources.UbiquityPluginConfig) (*flexProvisioner, error) {
	var identity types.UID
	identityPath := path.Join(config.LogPath, identityFile)
	request_context := logs.GetNewRequestContext()
	if _, err := os.Stat(identityPath); os.IsNotExist(err) {
		identity = uuid.NewUUID()
		err := ioutil.WriteFile(identityPath, []byte(identity), 0600)
		if err != nil {
			logger.Printf("Error writing identity file %s! %v", identityPath, err)
		}
	} else {
		read, err := ioutil.ReadFile(identityPath)
		if err != nil {
			logger.Printf("Error reading identity file %s! %v", config.LogPath, err)
		}
		identity = types.UID(strings.TrimSpace(string(read)))
	}
	provisioner := &flexProvisioner{
		logger:         logs.GetLogger(),
		identity:       identity,
		ubiquityClient: ubiquityClient,
		ubiquityConfig: config,
		podIPEnv:       podIPEnv,
		serviceEnv:     serviceEnv,
		namespaceEnv:   namespaceEnv,
		nodeEnv:        nodeEnv,
	}

	activateRequest := resources.ActivateRequest{Backends: config.Backends, Context: request_context}
	logger.Printf("activating backend %s\n", config.Backends)
	err := provisioner.ubiquityClient.Activate(activateRequest)

	return provisioner, err
}

type flexProvisioner struct {
	logger   logs.Logger
	identity types.UID
	// Whether the provisioner is running out of cluster and so cannot rely on
	// the existence of any of the pod, service, namespace, node env variables.
	outOfCluster bool

	ubiquityClient resources.StorageClient
	ubiquityConfig resources.UbiquityPluginConfig

	// Environment variables the provisioner pod needs valid values for in order to
	// put a service cluster IP as the server of provisioned NFS PVs, passed in
	// via downward API. If serviceEnv is set, namespaceEnv must be too.
	podIPEnv     string
	serviceEnv   string
	namespaceEnv string
	nodeEnv      string
}

// Provision creates a volume i.e. the storage asset and returns a PV object for
// the volume.
func (p *flexProvisioner) Provision(options controller.VolumeOptions) (*v1.PersistentVolume, error) {
	request_context := logs.GetNewRequestContext()
	go_id := logs.GetGoID()
	logs.GoIdToRequestIdMap.Store(go_id, request_context)
	defer logs.GetDeleteFromMapFunc(go_id)
	defer p.logger.Trace(logs.DEBUG)()
	if options.PVC == nil {
		return nil, fmt.Errorf("options missing PVC %#v", options)
	}

	// override volume name according to label
	pvName, ok := options.PVC.Labels["pv-name"]
	if ok {
		options.PVName = pvName
	}

	capacity, exists := options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	if !exists {
		return nil, fmt.Errorf("options.PVC.Spec.Resources.Requests does not contain capacity")
	}
	msg := fmt.Sprintf("PVC with capacity %d.", capacity.Value())
	p.logger.Info(msg)
	capacityMB := capacity.Value() / (1024 * 1024)

	volume_details, err := p.createVolume(options, capacityMB, request_context)
	if err != nil {
		return nil, err
	}

	annotations := make(map[string]string)
	annotations[annCreatedBy] = createdBy
	annotations[annProvisionerId] = k8sresources.UbiquityProvisionerName

	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        options.PVName,
			Labels:      map[string]string{},
			Annotations: annotations,
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: options.PersistentVolumeReclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)],
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				FlexVolume: &v1.FlexVolumeSource{
					Driver:    k8sresources.UbiquityK8sFlexVolumeDriverFullName,
					FSType:    "",
					SecretRef: nil,
					ReadOnly:  false,
					Options:   volume_details,
				},
			},
		},
	}

	return pv, nil
}

// Delete removes the directory that was created by Provision backing the given
// PV.
func (p *flexProvisioner) Delete(volume *v1.PersistentVolume) error {
	requestContext := logs.GetNewRequestContext()
	go_id := logs.GetGoID() 
	logs.GoIdToRequestIdMap.Store(go_id, requestContext)
	defer logs.GetDeleteFromMapFunc(go_id)
	defer p.logger.Trace(logs.DEBUG, logs.Args{{"volume name", volume.Name}})()

	if volume.Name == "" {
		return fmt.Errorf("volume name cannot be empty %#v", volume)
	}

	if volume.Spec.PersistentVolumeReclaimPolicy != v1.PersistentVolumeReclaimRetain {
		getVolumeRequest := resources.GetVolumeRequest{Name: volume.Name, Context: requestContext}
		volume, err := p.ubiquityClient.GetVolume(getVolumeRequest)
		if err != nil {
			fmt.Printf("error-retrieving-volume-info")
			return err
		}
		removeVolumeRequest := resources.RemoveVolumeRequest{Name: volume.Name, Context: requestContext}
		err = p.ubiquityClient.RemoveVolume(removeVolumeRequest)
		if err != nil {
			p.logger.Info("error removing volume")
			return err
		}
		return nil

	}

	return nil
}

func (p *flexProvisioner) createVolume(options controller.VolumeOptions, capacity int64, requestContext resources.RequestContext) (map[string]string, error) {
	defer p.logger.Trace(logs.DEBUG, logs.Args{{"volume name", options.PVName}})()

	ubiquityParams := make(map[string]interface{})
	if capacity != 0 {
		ubiquityParams["quota"] = fmt.Sprintf("%dM", capacity)    // SSc backend expect quota option
		ubiquityParams["size"] = fmt.Sprintf("%d", capacity/1024) // SCBE backend expect size option
	}
	for key, value := range options.Parameters {
		ubiquityParams[key] = value
	}
	backendName, exists := ubiquityParams["backend"]
	if !exists {
		return nil, fmt.Errorf("backend is not specified")
	}
	b := backendName.(string)
	createVolumeRequest := resources.CreateVolumeRequest{Name: options.PVName, Backend: b, Opts: ubiquityParams, Context: requestContext}
	err := p.ubiquityClient.CreateVolume(createVolumeRequest)
	if err != nil {
		return nil, fmt.Errorf("error creating volume: %v.", err)
	}

	getVolumeConfigRequest := resources.GetVolumeConfigRequest{Name: options.PVName, Context: requestContext}
	volumeConfig, err := p.ubiquityClient.GetVolumeConfig(getVolumeConfigRequest)
	if err != nil {
		return nil, fmt.Errorf("error getting volume config details: %v ", err)
	}

	flexVolumeConfig := make(map[string]string)
	flexVolumeConfig["volumeName"] = options.PVName
	for key, value := range volumeConfig {
		flexVolumeConfig[key] = fmt.Sprintf("%v", value)
	}

	return flexVolumeConfig, nil
}

package volume

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"github.ibm.com/almaden-containers/ubiquity/resources"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/pkg/api/v1"
)

const (

	// are we allowed to set this? else make up our own
	annCreatedBy = "kubernetes.io/createdby"
	createdBy    = "ubiquity-provisioner"

	// Name of the file where an nfsProvisioner will store its identity
	identityFile = "ubiquity-provisioner.identity"

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

func NewFlexProvisioner(logger *log.Logger, ubiquityClient resources.StorageClient, configPath string) (controller.Provisioner, error) {
	return newFlexProvisionerInternal(logger, ubiquityClient, configPath)
}

func newFlexProvisionerInternal(logger *log.Logger, ubiquityClient resources.StorageClient, configPath string) (*flexProvisioner, error) {
	var identity types.UID
	identityPath := path.Join(configPath, identityFile)
	if _, err := os.Stat(identityPath); os.IsNotExist(err) {
		identity = uuid.NewUUID()
		err := ioutil.WriteFile(identityPath, []byte(identity), 0600)
		if err != nil {
			logger.Printf("Error writing identity file %s! %v", identityPath, err)
		}
	} else {
		read, err := ioutil.ReadFile(identityPath)
		if err != nil {
			logger.Printf("Error reading identity file %s! %v", configPath, err)
		}
		identity = types.UID(strings.TrimSpace(string(read)))
	}
	provisioner := &flexProvisioner{
		logger:         logger,
		identity:       identity,
		ubiquityClient: ubiquityClient,
		podIPEnv:       podIPEnv,
		serviceEnv:     serviceEnv,
		namespaceEnv:   namespaceEnv,
		nodeEnv:        nodeEnv,
	}
	err := provisioner.ubiquityClient.Activate()

	return provisioner, err
}

type flexProvisioner struct {
	logger   *log.Logger
	identity types.UID
	// Whether the provisioner is running out of cluster and so cannot rely on
	// the existence of any of the pod, service, namespace, node env variables.
	outOfCluster bool

	ubiquityClient resources.StorageClient

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
	if options.PVC == nil {
		return nil, fmt.Errorf("options missing PVC %#v", options)
	}
	capacity, exists := options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	if !exists {
		return nil, fmt.Errorf("options.PVC.Spec.Resources.Requests does not contain capacity")
	}
	fmt.Printf("PVC with capacity %d", capacity.Value())
	capacityMB := capacity.Value() / (1024 * 1024)

	volume_details, err := p.createVolume(options, capacityMB)
	if err != nil {
		return nil, err
	}

	annotations := make(map[string]string)
	annotations[annCreatedBy] = createdBy
	annotations[annProvisionerId] = "ubiquity-provisioner"

	pv := &v1.PersistentVolume{
		ObjectMeta: v1.ObjectMeta{
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
					Driver:    "ibm/ubiquity",
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
	if volume.Name == "" {
		return fmt.Errorf("volume name cannot be empty %#v", volume)
	}
	if volume.Spec.PersistentVolumeReclaimPolicy != v1.PersistentVolumeReclaimRetain {

		err := p.ubiquityClient.RemoveVolume(volume.Name)
		if err != nil {
			return err
		}
		return nil

	}

	return nil
}

func (p *flexProvisioner) createVolume(options controller.VolumeOptions, capacity int64) (map[string]string, error) {
	ubiquityParams := make(map[string]interface{})
	if capacity != 0 {
		ubiquityParams["quota"] = fmt.Sprintf("%dM", capacity)
	}
	for key, value := range options.Parameters {
		ubiquityParams[key] = value
	}
	err := p.ubiquityClient.CreateVolume(options.PVName, ubiquityParams)
	if err != nil {
		return nil, fmt.Errorf("error creating volume: %v", err)
	}

	volumeConfig, err := p.ubiquityClient.GetVolumeConfig(options.PVName)
	if err != nil {
		return nil, fmt.Errorf("error getting volume config details: %v", err)
	}

	flexVolumeConfig := make(map[string]string)
	flexVolumeConfig["volumeName"] = options.PVName
	for key, value := range volumeConfig {
		flexVolumeConfig[key] = fmt.Sprintf("%v", value)
	}

	return flexVolumeConfig, nil
}

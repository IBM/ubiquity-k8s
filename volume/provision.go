package volume

import (
	"fmt"
	"log"

	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"github.ibm.com/almaden-containers/ubiquity/resources"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
)

const (

	// are we allowed to set this? else make up our own
	annCreatedBy = "kubernetes.io/createdby"
	createdBy    = "flex-dynamic-provisioner"

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

func NewFlexProvisioner(logger *log.Logger, client kubernetes.Interface, flexClient resources.StorageClient) (controller.Provisioner, error) {
	return newFlexProvisionerInternal(logger, client, flexClient)
}

func newFlexProvisionerInternal(logger *log.Logger, client kubernetes.Interface, flexClient resources.StorageClient) (*flexProvisioner, error) {

	provisioner := &flexProvisioner{
		logger:         logger,
		client:         client,
		ubiquityClient: flexClient,
		podIPEnv:       podIPEnv,
		serviceEnv:     serviceEnv,
		namespaceEnv:   namespaceEnv,
		nodeEnv:        nodeEnv,
	}
	err := provisioner.ubiquityClient.Activate()

	return provisioner, err
}

type flexProvisioner struct {
	logger *log.Logger
	// Client, needed for getting a service cluster IP to put as the NFS server of
	// provisioned PVs
	client kubernetes.Interface

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

var _ controller.Provisioner = &flexProvisioner{}

// Provision creates a volume i.e. the storage asset and returns a PV object for
// the volume.
func (p *flexProvisioner) Provision(options controller.VolumeOptions) (*v1.PersistentVolume, error) {
	capacity := options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	capacityMB := capacity.Value() / (1024 * 1024)

	volume_details, err := p.createVolume(options, capacityMB)
	if err != nil {
		return nil, err
	}

	annotations := make(map[string]string)
	annotations[annCreatedBy] = createdBy

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
					Driver:    "kubernetes.io/ubiquity",
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
	//remote.NewRemoteClient(log,backendName,url,config)
	err := p.ubiquityClient.RemoveVolume(volume.Name, true)

	if err != nil {
		return err
	}
	return nil
}

func (p *flexProvisioner) createVolume(options controller.VolumeOptions, capacity int64) (map[string]string, error) {
	ubiquityParams := make(map[string]interface{})
	ubiquityParams["quota"] = fmt.Sprintf("%dM", capacity)
	for key, value := range options.Parameters {
		ubiquityParams[key] = value
	}
	err := p.ubiquityClient.CreateVolume(options.PVName, ubiquityParams)
	if err != nil {
		return nil, fmt.Errorf("error creating volume: %v", err)
	}

	_, volumeConfig, err := p.ubiquityClient.GetVolume(options.PVName)
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

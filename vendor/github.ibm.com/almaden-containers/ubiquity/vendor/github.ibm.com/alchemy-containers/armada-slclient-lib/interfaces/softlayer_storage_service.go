package interfaces

import (
	"github.ibm.com/alchemy-containers/armada-slclient-lib/datatypes"
)

type Softlayer_Storage_Service interface {
	CreateStorage(size int, iops int, location string) (datatypes.SoftLayer_Storage, error)
	DeleteStorage(volumeId int) error
	GetBillingItem(volumeId int) (datatypes.Billing_Item, error)
	GetStorage(volumeId int) (datatypes.SoftLayer_Storage, error)
	AllowAccessFromSubnet(volumeId int, subnetId int) (bool, error)
	RemoveAccessFromSubnet(volumeId int, subnetId int) (bool, error)
	AllowAccessFromAllSubnets(volumeId int) (bool, error)
	RemoveAccessFromAllSubnets(volumeId int) (bool, error)
	GetAllowableSubnets(volumeId int) ([]datatypes.SoftLayer_Subnet, error)
	GetAllowedSubnets(volumeId int) ([]datatypes.SoftLayer_Subnet, error)
}

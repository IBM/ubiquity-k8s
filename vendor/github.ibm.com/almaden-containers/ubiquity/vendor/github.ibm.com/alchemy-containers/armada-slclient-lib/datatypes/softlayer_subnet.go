package datatypes

import (
	"time"
)

type SoftLayer_Subnet_Parameters struct {
	Parameters []SoftLayer_Subnet `json:"parameters"`
}
type SoftLayer_Subnet struct {
	BroadcastAddress     string    `json:"broadcastAddress,omitempty"`
	Cidr                 int       `json:"cidr,omitempty"`
	Gateway              string    `json:"gateway,omitempty"`
	Id                   int       `json:"id,omitempty"`
	IsCustomerOwned      bool      `json:"isCustomerOwned,omitempty"`
	IsCustomerRoutable   bool      `json:"isCustomerRoutable,omitempty"`
	ModifyDate           time.Time `json:"modifyDate,omitempty"`
	Netmask              string    `json:"netmask,omitempty"`
	NetworkIdentifier    string    `json:"networkIdentifier,omitempty"`
	NetworkVlanId        int       `json:"networkVlanId,omitempty"`
	Note                 string    `json:"note,omitempty"`
	SortOrder            string    `json:"sortOrder,omitempty"`
	SubnetType           string    `json:"subnetType,omitempty"`
	TotalIpAddresses     string    `json:"totalIpAddresses,omitempty"`
	UsableIpAddressCount string    `json:"usableIpAddressCount,omitempty"`
	Version              int       `json:"version,omitempty"`
}

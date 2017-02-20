package models

type QuotaFields struct {
	GUID                    string `json:"guid,omitempty"`
	Name                    string `json:"name"`
	MemoryLimitInMB         int64  `json:"memory_limit"`
	InstanceMemoryLimitInMB int64  `json:"instance_memory_limit"`
	RoutesLimit             int    `json:"total_routes"`
	ServicesLimit           int    `json:"total_services"`
	NonBasicServicesAllowed bool   `json:"non_basic_services_allowed"`
	AppInstanceLimit        int    `json:"app_instance_limit"`
}

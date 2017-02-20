package datatypes

type SoftLayer_Location struct {
	Id     		     	 int       `json:"id,omitempty"`
	LongName             string    `json:"longName,omitempty"`
	Name	             string    `json:"name,omitempty"`
	StatusId             int       `json:"statusId,omitempty"`
}

package mock

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type MockHttpClient struct {
	HTTPClient *http.Client
	username   string
	password   string
	apiUrl     string
	protocol   string
}

func NewMockHttpClient() *MockHttpClient {
	hClient := &MockHttpClient{
		username:   "dummy",
		password:   "dummy",
		apiUrl:     "api.softlayer.com/rest/v3",
		protocol:   "https",
		HTTPClient: http.DefaultClient,
	}
	return hClient
}

var mockResponses = []struct {
	testName    string
	requestType string
	url         string
	requestBody string
	response    string
	err         error
}{
	{
		"TestGetStorage",
		"GET",
		"/rest/v3/SoftLayer_Network_Storage/18061033/getObject.json?objectMask=accountId;capacityGb;createDate;id;username;billingItem.id;billingItem.orderItem.order.id;serviceResourceBackendIpAddress;totalBytesUsed;activeTransactions",
		"",
		`{"accountId":531277,"capacityGb":20,"createDate":"2016-12-31T04:37:43-05:00","id":18061033,"username":"IBM02SEV531277_3587","activeTransactions":[],"billingItem":{"id":143306907,"orderItem":{"order":{"id":11956981}}},"serviceResourceBackendIpAddress":"fsm-dal0901a-fz.service.softlayer.com","totalBytesUsed":"0"}`,
		nil,
	},
	{
		"TestGetBillingItem",
		"GET",
		"/rest/v3/SoftLayer_Network_Storage/18061033/getBillingItem.json",
		"",
		`{"allowCancellationFlag":1,"cancellationDate":null,"categoryCode":"storage_service_enterprise","createDate":"2016-12-31T04:36:41-05:00","cycleStartDate":null,"description":"Endurance Storage","id":143306907,"lastBillDate":"2017-01-01T03:20:17-05:00","modifyDate":null,"nextBillDate":"2017-02-01T01:00:00-05:00","orderItemId":170028889,"parentId":null,"recurringMonths":null,"serviceProviderId":null}`,
		nil,
	},
	{
		"TestGetAllowableSubnets",
		"GET",
		"/rest/v3/SoftLayer_Network_Storage/18061033/getAllowableSubnets",
		"",
		`[{"broadcastAddress":"10.130.231.255","cidr":25,"gateway":"10.130.231.129","id":1298475,"modifyDate":"2016-10-04T03:46:04-06:00","netmask":"255.255.255.128","networkIdentifier":"10.130.231.128","networkVlanId":1374055,"note":"Dev/Mex01p01/Rad01/B","sortOrder":"1","subnetType":"ADDITIONAL_PRIMARY","totalIpAddresses":"128","usableIpAddressCount":"125","version":4}]`,
		nil,
	},
	{
		"TestAllowAccessFromAllSubnets",
		"PUT",
		"/rest/v3/SoftLayer_Network_Storage/18061033/allowAccessFromSubnet",
		"",
		"true",
		nil,
	},
	{
		"TestGetAllowedSubnets",
		"GET",
		"/rest/v3/SoftLayer_Network_Storage/18061033/getAllowedSubnets",
		"",
		`[{"broadcastAddress":"10.130.231.255","cidr":25,"gateway":"10.130.231.129","id":1298475,"modifyDate":"2016-10-04T03:46:04-06:00","netmask":"255.255.255.128","networkIdentifier":"10.130.231.128","networkVlanId":1374055,"note":"Dev/Mex01p01/Rad01/B","sortOrder":"1","subnetType":"ADDITIONAL_PRIMARY","totalIpAddresses":"128","usableIpAddressCount":"125","version":4}]`,
		nil,
	},
	{
		"TestRemoveAccessFromAllSubnets",
		"PUT",
		"/rest/v3/SoftLayer_Network_Storage/18061033/removeAccessFromSubnet",
		"",
		"true",
		nil,
	},
	{
		"TestAllowAccessFromAllSubnetsWhenErrorIsNotNil",
		"PUT",
		"/rest/v3/SoftLayer_Network_Storage/18061034/allowAccessFromSubnet",
		"",
		"false",
		errors.New("Unable to allow access"),
	},
	{
		"TestRemoveAccessFromAllSubnetsWhenErrorIsNotNil",
		"PUT",
		"/rest/v3/SoftLayer_Network_Storage/18061034/removeAccessFromSubnet",
		"",
		"false",
		errors.New("Unable to remove access"),
	},
}

func (client *MockHttpClient) DoHttpRequest(path string, masks string, filters string, requestType string, requestBody *bytes.Buffer) ([]byte, error) {
	url := fmt.Sprintf("%s://%s:%s@%s/%s", client.protocol, client.username, client.password, client.apiUrl, path)
	if requestType == "GET" {
		if filters != "" && masks != "" {
			url += "?objectFilter=" + filters + "&objectMask=filteredMask[" + masks + "]"
		} else if filters != "" {
			url += "?objectFilter=" + filters
		} else if masks != "" {
			url += "?objectMask=" + masks
		}
	}
	return client.makeHttpRequest(url, requestType, requestBody)
}

func (client *MockHttpClient) makeHttpRequest(url string, requestType string, requestBody *bytes.Buffer) ([]byte, error) {

	for _, res := range mockResponses {
		fmt.Println(res.url)
		if (requestType == res.requestType) && strings.Contains(url, res.url) {
			return []byte(res.response), res.err
		}
	}
	return []byte{}, errors.New("Invalid request URL")
}

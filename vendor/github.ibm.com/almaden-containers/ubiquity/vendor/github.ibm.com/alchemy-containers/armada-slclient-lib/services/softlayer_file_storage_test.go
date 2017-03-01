package services

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.ibm.com/alchemy-containers/armada-slclient-lib/client/mock"
	"github.ibm.com/alchemy-containers/armada-slclient-lib/interfaces"
	"testing"
)

var sss interfaces.Softlayer_Storage_Service
var storageId = 18061033 //now dummy, with actual let this variable set from create storage testcase
var storageIdErr = 18061034

func init() {
	httpclient := mock.NewMockHttpClient()
	serviceclient := NewServiceClientWithParam(httpclient)
	//serviceclient := NewServiceClientWithAuth("user","pass")  //for Actual testing
	sss = serviceclient.GetSoftLayerStorageService()
	fmt.Println("successfully initialized the storage service")
}

func TestCreateStorage(t *testing.T) {

}

func TestDeleteStorage(t *testing.T) {

}

func TestGetBillingItem(t *testing.T) {
	billingItem, err := sss.GetBillingItem(storageId)
	if err != nil {
		t.Error(fmt.Sprintf("\nUnable to find Billing id  Error: %s", err.Error()))
	} else {
		fmt.Printf("\n Biliingid: %d", billingItem.Id)
	}
}

func TestGetStorage(t *testing.T) {

	storage, err := sss.GetStorage(storageId)
	if err != nil {
		t.Error(fmt.Sprintf("\nUnable to find Storage Error: %s", err.Error()))
	} else {
		fmt.Printf("\n Storage found. servername: %s mountpath: %s", storage.ServiceResourceBackendIpAddress, storage.Username)
	}
}

func TestAllowAccessFromSubnet(t *testing.T) {

}

func TestRemoveAccessFromSubnet(t *testing.T) {

}

func TestGetAllowableSubnets(t *testing.T) {
	result, _ := sss.GetAllowableSubnets(storageId)
	assert.NotEmpty(t, result)
}

func TestAllowAccessFromAllSubnets(t *testing.T) {
	result, _ := sss.AllowAccessFromAllSubnets(storageId)
	assert.Equal(t, true, result)
}

func TestGetAllowedSubnets(t *testing.T) {
	result, _ := sss.GetAllowedSubnets(storageId)
	assert.NotEmpty(t, result)
}

func TestRemoveAccessFromAllSubnets(t *testing.T) {
	result, _ := sss.RemoveAccessFromAllSubnets(storageId)
	assert.Equal(t, true, result)
}

func TestAllowAccessFromAllSubnetsWhenErrorIsNotNil(t *testing.T) {
	result, _ := sss.AllowAccessFromAllSubnets(storageIdErr)
	assert.Equal(t, false, result)
}

func TestRemoveAccessFromAllSubnetsWhenErrorIsNotNil(t *testing.T) {
	result, _ := sss.RemoveAccessFromAllSubnets(storageIdErr)
	assert.Equal(t, false, result)
}

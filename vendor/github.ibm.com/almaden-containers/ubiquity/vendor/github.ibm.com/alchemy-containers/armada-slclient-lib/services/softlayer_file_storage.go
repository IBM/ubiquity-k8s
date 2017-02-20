package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.ibm.com/alchemy-containers/armada-slclient-lib/datatypes"
	"github.ibm.com/alchemy-containers/armada-slclient-lib/interfaces"
	"strconv"
	"time"
)

const (
	NETWORK_ENDURANCE_STORAGE_PACKAGE_ID = 240
	SL_CREATE_STORAGE_TIMEOUT            = 600
	SL_CREATE_STORAGE_POLLING_INTERVAL   = 10
)

var IOPS_TIER = map[int]string{100: "LOW_INTENSITY_TIER", 200: "READHEAVY_TIER", 300: "WRITEHEAVY_TIER"}

type softLayer_File_Storage_Service struct {
	httpclient interfaces.HttpClient
}

func NewSoftLayer_File_Storage_Service(httpclient interfaces.HttpClient) *softLayer_File_Storage_Service {
	return &softLayer_File_Storage_Service{
		httpclient: httpclient,
	}
}

func (slfs *softLayer_File_Storage_Service) CreateStorage(size int, iops int, location string) (datatypes.SoftLayer_Storage, error) {

	//Get location id
	location_id, err := slfs.getLocation(location)
	if err != nil {
		return datatypes.SoftLayer_Storage{}, err
	}

	//Get Item prices
	itemPriceId1, err := slfs.getItemPriceIdBySizeAndIops(size, iops)
	if err != nil {
		return datatypes.SoftLayer_Storage{}, err
	}

	itemPriceId2, err := slfs.getItemPriceIdByItemKeyName(IOPS_TIER[iops])
	if err != nil {
		return datatypes.SoftLayer_Storage{}, err
	}

	itemPriceId3, err := slfs.getItemPriceIdByItemKeyName("FILE_STORAGE_2")
	if err != nil {
		return datatypes.SoftLayer_Storage{}, err
	}

	itemPriceId4, err := slfs.getItemPriceIdByItemKeyName("CODENAME_PRIME_STORAGE_SERVICE")
	if err != nil {
		return datatypes.SoftLayer_Storage{}, err
	}

	var requestBody = []byte(`{
						    "parameters": [{
						        "location":` + strconv.Itoa(location_id) + `,
						        "packageId":` + strconv.Itoa(NETWORK_ENDURANCE_STORAGE_PACKAGE_ID) + `,
						        "osFormatType": {
						            "keyName": "LINUX"
						        },
						        "complexType": "SoftLayer_Container_Product_Order_Network_Storage_Enterprise",
						        "prices": [{
						            "id":` + strconv.Itoa(itemPriceId1) + `
						        }, {
						            "id":` + strconv.Itoa(itemPriceId2) + `
						        }, {
						            "id":` + strconv.Itoa(itemPriceId3) + `
						        }, {
						            "id":` + strconv.Itoa(itemPriceId4) + `
						        }],
						        "quantity": 1
						    	}]
							}`)

	responseBytes, err := slfs.httpclient.DoHttpRequest("SoftLayer_Product_Order/placeOrder", "", "", "POST", bytes.NewBuffer(requestBody))
	if err != nil {
		return datatypes.SoftLayer_Storage{}, err
	}

	receipt := datatypes.SoftLayer_Product_Order_Receipt{}
	err = json.Unmarshal(responseBytes, &receipt)
	if err != nil {
		return datatypes.SoftLayer_Storage{}, err
	}

	var storage datatypes.SoftLayer_Storage
	MAX_RETRY_COUNT := SL_CREATE_STORAGE_TIMEOUT / SL_CREATE_STORAGE_POLLING_INTERVAL
	for i := 1; i <= MAX_RETRY_COUNT; i++ {
		storage, err = slfs.findStorage(receipt.OrderId)
		if err != nil {
			//log message
			fmt.Printf("\nFailed to find volume with orderId `%d` due to `%s`, retrying...", receipt.OrderId, err.Error())
		} else if len(storage.ActiveTransactions) > 0 {
			//log message
			errmsg := fmt.Sprintf("Storage found with orderId:%d volumeId:%d. Waiting for the %d active transactions to be completed. retrying...", receipt.OrderId, storage.Id, len(storage.ActiveTransactions))
			fmt.Println(errmsg)
			err = errors.New(errmsg)
		} else {
			break
		}
		time.Sleep(time.Duration(SL_CREATE_STORAGE_POLLING_INTERVAL) * time.Second)
	}

	if err != nil {
		err = errors.New(fmt.Sprintf("Failed to find volume with id `%d` after retry within `%d` seconds", receipt.OrderId, SL_CREATE_STORAGE_TIMEOUT))
	}
	return storage, err
}

func (slfs *softLayer_File_Storage_Service) getLocation(locationName string) (int, error) {
	var location []datatypes.SoftLayer_Location

	//Get Location
	location_id := 0
	response, err := slfs.httpclient.DoHttpRequest("SoftLayer_Location_Datacenter/getDatacenters", "", "", "GET", new(bytes.Buffer))
	if err != nil {
		return location_id, err
	}
	err = json.Unmarshal(response, &location)

	for i := range location {
		if location[i].Name == locationName {
			location_id = location[i].Id
			break
		}
	}

	if location_id == 0 {
		errorMessage := "Failed to get datacenter name '" + locationName + "'from SL"
		return location_id, errors.New(errorMessage)
	}
	return location_id, nil
}

func (slfs *softLayer_File_Storage_Service) findStorage(orderId int) (datatypes.SoftLayer_Storage, error) {
	path := "SoftLayer_Account/getNetworkStorage"
	objectFilter := string(`{"networkStorage":{"billingItem":{"orderItem":{"order":{"id":{"operation":` + strconv.Itoa(orderId) + `}}}}}}`)
	objectMasks := string(`accountId;capacityGb;createDate;id;username;billingItem.id;billingItem.orderItem.order.id;serviceResourceBackendIpAddress;activeTransactions`)

	responseBytes, err := slfs.httpclient.DoHttpRequest(path, objectMasks, objectFilter, "GET", new(bytes.Buffer))
	if err != nil {
		errorMessage := fmt.Sprintf("could not SoftLayer_Account#getIscsiNetworkStorage, error message '%s'", err.Error())
		return datatypes.SoftLayer_Storage{}, errors.New(errorMessage)
	}

	networkStorage := []datatypes.SoftLayer_Storage{}
	err = json.Unmarshal(responseBytes, &networkStorage)
	if err != nil {
		errorMessage := fmt.Sprintf("failed to decode JSON response, err message '%s'", err.Error())
		err := errors.New(errorMessage)
		return datatypes.SoftLayer_Storage{}, err
	}

	if len(networkStorage) == 1 {
		return networkStorage[0], nil
	}
	return datatypes.SoftLayer_Storage{}, errors.New(fmt.Sprintf("Cannot find storage with order id %d", orderId))
}

func (slfs *softLayer_File_Storage_Service) DeleteStorage(volumeId int) error {

	billingItem, err := slfs.GetBillingItem(volumeId)
	if err != nil {
		return err
	}
	if billingItem.Id > 0 {

		path := fmt.Sprintf("SoftLayer_Billing_Item/%d/cancelService.json", billingItem.Id)
		response, err := slfs.httpclient.DoHttpRequest(path, "", "", "GET", new(bytes.Buffer))
		if err != nil {
			return err
		}

		if res := string(response[:]); res != "true" {
			errorMessage := fmt.Sprintf("Could not do SoftLayer_Network_Storage#DeleteStorage with id: '%d'", volumeId)
			return errors.New(errorMessage)
		}
	}
	return nil
}

func (slfs *softLayer_File_Storage_Service) GetBillingItem(volumeId int) (datatypes.Billing_Item, error) {

	path := fmt.Sprintf("SoftLayer_Network_Storage/%d/getBillingItem.json", volumeId)
	response, err := slfs.httpclient.DoHttpRequest(path, "", "", "GET", new(bytes.Buffer))
	if err != nil {
		return datatypes.Billing_Item{}, err
	}

	billingItem := datatypes.Billing_Item{}
	err = json.Unmarshal(response, &billingItem)
	if err != nil {
		return datatypes.Billing_Item{}, err
	}

	return billingItem, nil
}

func (slfs *softLayer_File_Storage_Service) GetStorage(volumeId int) (datatypes.SoftLayer_Storage, error) {
	path := fmt.Sprintf("SoftLayer_Network_Storage/%d/getObject.json", volumeId)
	objectMasks := string(`accountId;capacityGb;createDate;id;username;billingItem.id;billingItem.orderItem.order.id;serviceResourceBackendIpAddress;totalBytesUsed;activeTransactions`)
	response, err := slfs.httpclient.DoHttpRequest(path, objectMasks, "", "GET", new(bytes.Buffer))

	if err != nil {
		return datatypes.SoftLayer_Storage{}, err
	}

	volume := datatypes.SoftLayer_Storage{}
	err = json.Unmarshal(response, &volume)
	if err != nil {
		return datatypes.SoftLayer_Storage{}, err
	}

	return volume, nil
}

func (slfs *softLayer_File_Storage_Service) AllowAccessFromSubnet(volumeId int, subnetId int) (bool, error) {
	return slfs.modifyAccessFromSubnet("allowAccessFromSubnet", volumeId, subnetId)
}

func (slfs *softLayer_File_Storage_Service) RemoveAccessFromSubnet(volumeId int, subnetId int) (bool, error) {
	return slfs.modifyAccessFromSubnet("removeAccessFromSubnet", volumeId, subnetId)
}

func (slfs *softLayer_File_Storage_Service) AllowAccessFromAllSubnets(volumeId int) (bool, error) {
	subnets, err := slfs.GetAllowableSubnets(volumeId)
	if err != nil {
		return false, err
	}
	return slfs.modifyAccessFromSubnets("allowAccessFromSubnet", volumeId, subnets)

}

func (slfs *softLayer_File_Storage_Service) RemoveAccessFromAllSubnets(volumeId int) (bool, error) {
	subnets, err := slfs.GetAllowedSubnets(volumeId)
	if err != nil {
		return false, err
	}
	return slfs.modifyAccessFromSubnets("removeAccessFromSubnet", volumeId, subnets)
}

func (slfs *softLayer_File_Storage_Service) GetAllowableSubnets(volumeId int) ([]datatypes.SoftLayer_Subnet, error) {
	return slfs.getAllSubnets("getAllowableSubnets", volumeId)
}

func (slfs *softLayer_File_Storage_Service) GetAllowedSubnets(volumeId int) ([]datatypes.SoftLayer_Subnet, error) {
	return slfs.getAllSubnets("getAllowedSubnets", volumeId)
}

func (slfs *softLayer_File_Storage_Service) modifyAccessFromSubnets(accessType string, volumeId int, subnets []datatypes.SoftLayer_Subnet) (bool, error) {

	result := false
	for i := range subnets {

		params := datatypes.SoftLayer_Subnet_Parameters{
			Parameters: []datatypes.SoftLayer_Subnet{
				subnets[i],
			},
		}

		requestBody, err1 := json.Marshal(params)
		if err1 != nil {
			return result, err1
		}

		path := fmt.Sprintf("SoftLayer_Network_Storage/%d/%s", volumeId, accessType)
		response, err2 := slfs.httpclient.DoHttpRequest(path, "", "", "PUT", bytes.NewBuffer(requestBody))
		if err2 != nil {
			return result, err2
		}

		err3 := json.Unmarshal(response, &result)
		if err3 != nil {
			return result, err3
		}
		if result == false {
			errorMessage := fmt.Sprintf("could not perform '%d' on subnet list  ", accessType)
			return result, errors.New(errorMessage)
		}
	}
	return result, nil
}

func (slfs *softLayer_File_Storage_Service) modifyAccessFromSubnet(accessType string, volumeId int, subnetId int) (bool, error) {

	//Get subnet
	subnet, err1 := slfs.getSubnet(subnetId)
	if err1 != nil {
		return false, err1
	}

	//Allow access from subnet
	parameters := datatypes.SoftLayer_Subnet_Parameters{
		Parameters: []datatypes.SoftLayer_Subnet{
			subnet,
		},
	}
	requestBody, err2 := json.Marshal(parameters)
	if err2 != nil {
		return false, err2
	}
	path := fmt.Sprintf("SoftLayer_Network_Storage/%d/%s", volumeId, accessType)
	response, err3 := slfs.httpclient.DoHttpRequest(path, "", "", "PUT", bytes.NewBuffer(requestBody))
	if err3 != nil {
		return false, err3
	}

	allowed, err4 := strconv.ParseBool(string(response[:]))
	if err4 != nil {
		return false, err4
	}
	return allowed, nil
}

func (slfs *softLayer_File_Storage_Service) getSubnet(subnetId int) (datatypes.SoftLayer_Subnet, error) {

	subnet := datatypes.SoftLayer_Subnet{}

	//Get subnet
	path := fmt.Sprintf("SoftLayer_Network_Subnet/%d/getObject.json", subnetId)
	response, err := slfs.httpclient.DoHttpRequest(path, "", "", "GET", new(bytes.Buffer))

	if err != nil {
		return subnet, err
	}

	err = json.Unmarshal(response, &subnet)
	fmt.Println(subnet)
	return subnet, nil
}

func (slfs *softLayer_File_Storage_Service) getAllSubnets(accessType string, volumeId int) ([]datatypes.SoftLayer_Subnet, error) {

	subnets := []datatypes.SoftLayer_Subnet{}

	path := fmt.Sprintf("SoftLayer_Network_Storage/%d/%s", volumeId, accessType)
	response, err := slfs.httpclient.DoHttpRequest(path, "", "", "GET", new(bytes.Buffer))

	if err != nil {
		return subnets, err
	}

	err1 := json.Unmarshal(response, &subnets)
	return subnets, err1
}

func (slfs *softLayer_File_Storage_Service) getItemPriceIdBySizeAndIops(size int, iops int) (int, error) {

	packageId := NETWORK_ENDURANCE_STORAGE_PACKAGE_ID
	filters := fmt.Sprintf(`{"itemPrices":{"item":{"capacity":{"operation":%d}},"attributes":{"value":{"operation":%d}},"categories":{"categoryCode":{"operation":"performance_storage_space"}}}}`, size, iops)
	objectMasks := string(`id;locationGroupId;item.id;item.keyName;item.units;item.description;item.capacity`)
	path := fmt.Sprintf("%s/%d/getItemPrices.json", "SoftLayer_Product_Package", packageId)

	response, err := slfs.httpclient.DoHttpRequest(path, objectMasks, filters, "GET", new(bytes.Buffer))
	if err != nil {
		return 0, err
	}

	itemPrices := []datatypes.SoftLayer_Product_Item_Price{}
	err = json.Unmarshal(response, &itemPrices)
	if err != nil {
		return 0, err
	}

	var currentItemId int

	if len(itemPrices) > 0 {
		for _, itemPrice := range itemPrices {
			if itemPrice.LocationGroupId == 0 {
				currentItemId = itemPrice.Id
				break
			}
		}
	}

	if currentItemId == 0 {
		return 0, errors.New(fmt.Sprintf("No item priceId found for size %d and iops ", size, iops))
	}
	return currentItemId, nil
}

func (slfs *softLayer_File_Storage_Service) getItemPriceIdByItemKeyName(keyName string) (int, error) {

	packageId := NETWORK_ENDURANCE_STORAGE_PACKAGE_ID
	filters := string(`{"itemPrices":{"item":{"keyName":{"operation":"` + keyName + `"}}}}`)
	objectMasks := string(`id;locationGroupId;item.id;item.keyName;item.units;item.description;item.capacity`)
	path := fmt.Sprintf("%s/%d/getItemPrices.json", "SoftLayer_Product_Package", packageId)

	response, err := slfs.httpclient.DoHttpRequest(path, objectMasks, filters, "GET", new(bytes.Buffer))
	if err != nil {
		return 0, err
	}

	itemPrices := []datatypes.SoftLayer_Product_Item_Price{}
	err = json.Unmarshal(response, &itemPrices)
	if err != nil {
		return 0, err
	}

	var currentItemId int
	if len(itemPrices) > 0 {
		for _, itemPrice := range itemPrices {
			if itemPrice.LocationGroupId == 0 {
				currentItemId = itemPrice.Id
				break
			}
		}
	}

	if currentItemId == 0 {
		return 0, errors.New(fmt.Sprintf("No item priceId found with item KeyName %s", keyName))
	}
	return currentItemId, nil
}

package main

import (
	"fmt"
	"github.ibm.com/alchemy-containers/armada-slclient-lib/services"
)

func main() {
	fmt.Println("Sample Test Start")

	//Get the service client
	username := "dev-mon01-con.alchemy.dev"
	password := "XXXXX"
	serviceClient := services.NewServiceClientWithAuth(username, password)
	
	//Get the Softlayer storage service
	sl_storage_provider := serviceClient.GetSoftLayerStorageService()
	fmt.Println(sl_storage_provider)
	
	//Create Storage
	fmt.Printf("\n*********** TEST CreateStorage ***********\n")
	//storage, err1 := sl_storage_provider.CreateStorage(20, 100, "dal09")
	storage, err1 := sl_storage_provider.CreateStorage(20, 100, "mex01")
	fmt.Println(storage)
	if err1 != nil {
		fmt.Printf("Storage creation failed with error: %s",err1)
		return
	} else {
		fmt.Printf("Storage created with id: %d",storage.Id)
	}
	
	//Retrieve Billing item
	storageId := storage.Id
//	storageId := 18475723
//	storageId := 18061033
//  storageId := 18067311
	fmt.Printf("\n*********** TEST GetBillingItem ***********")
	billingItem, err2 := sl_storage_provider.GetBillingItem(storageId)
	if err2 != nil {
		fmt.Printf("\nUnable to find Billing id for storageid: %d with Error: %s",storageId,err2)
		return
	} else {
		fmt.Printf("\n storageid: %d Biliingid: %d",storageId, billingItem.Id)
	}


	//Retrieve Storage
	fmt.Printf("\n*********** TEST GetStorage ***********")
	storage, err10 := sl_storage_provider.GetStorage(storageId)
	if err10 != nil {
		fmt.Printf("\nUnable to find Storage for storageid: %d with Error: %s",storageId,err10)
		return
	} else {
		fmt.Printf("\n Storage found. servername: %s mountpath: %s",storage.ServiceResourceBackendIpAddress, storage.Username)
	}

//	//Authorize Storage
//	fmt.Printf("\n*********** TEST Authorize ***********")
//	authorized, err11 := sl_storage_provider.AllowAccessFromSubnet(storageId,963953)
//	
//	if err11 != nil {
//		fmt.Printf("\nUnable to authorize Storage for storageid: %d with Error: %s",storageId,err11)
//		return
//	} else {
//		println(authorized)
//		fmt.Printf("\n Storage authorized with storageId: %s subnetid:963953",storageId)
//	}
//	
//	//UnAuthorize Storage
//	fmt.Printf("\n*********** TEST UnAuthorize ***********")
//	unauthorized, err12 := sl_storage_provider.RemoveAccessFromSubnet(storageId,963953)
//	
//	if err12 != nil {
//		fmt.Printf("\nUnable to unauthorize Storage for storageid: %d with Error: %s",storageId,err12)
//		return
//	} else {
//		println(authorized)
//		fmt.Printf("\n Storage unauthorized with storageId: %d subnetid:963953",storageId)
//	}
	
	//Authorize Storage
	fmt.Printf("\n*********** TEST Authorize All Subnets ***********")
	authorized, err13 := sl_storage_provider.AllowAccessFromAllSubnets(storageId)
	
	if err13 != nil {
		fmt.Printf("\nUnable to authorize all subnets for Storage: %d with Error: %s",storageId,err13)
		return
	} else {
		println(authorized)
		fmt.Printf("\nStorage authorized for all subnets. storageId: %s ",storageId)
	}
	
	//UnAuthorize Storage
	fmt.Printf("\n*********** TEST UnAuthorize All ***********")
	unauthorized, err14 := sl_storage_provider.RemoveAccessFromAllSubnets(storageId)
	
	if err14 != nil {
		fmt.Printf("\nUnable to unauthorize Storage for storageid: %d with Error: %s",storageId,err14)
		return
	} else {
		println(unauthorized)
		fmt.Printf("\n Storage unauthorized with storageId: %d subnetid:963953",storageId)
	}

	
	// Delete Storage
	fmt.Printf("\n*********** TEST DeleteStorage ***********")
	err3 := sl_storage_provider.DeleteStorage(storageId)
	if err3 != nil {
		fmt.Printf("\nDelete failed for storageid: %d with Error: %s",storageId,err3)
		return
	} else {
		fmt.Printf("\nDelete successful for storageid: %d",storageId)
	}
	
	fmt.Println("\nSample Test Ends")
}

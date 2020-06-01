package services

import (
	"github.com/EC-SEAL/perseal/externaldrive"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
)

var (
	mockUUID = "77358e9d-3c59-41ea-b4e3-04922657b30c" // As ID for DataStore
)

// Store Data on the corresponding PDS
func StoreCloudData(data sm.SessionMngrResponse, pds string, id string, filename string, cameFrom string) (password string, dataStore *externaldrive.DataStore, err *model.DashboardResponse) {
	uuid := mockUUID

	data, err = checkClientId(data, pds)
	if err != nil {
		return
	}
	if pds == "googleDrive" {
		password, dataStore, err = storeSessionDataGoogleDrive(data, uuid, id, filename, cameFrom) // No password
		return

	} else if pds == "oneDrive" {
		password, dataStore, err = storeSessionDataOneDrive(data, uuid, id, filename, cameFrom) // No password
	} else {
		err = &model.DashboardResponse{
			Code:    400,
			Message: "Wrong Module Or No Module Found in Credentials",
		}
		return
	}
	return
}

// Back-channel store may only be used for local Browser storing
func StoreLocalData(data sm.SessionMngrResponse, pds string, cipherPassword string) (dataStore *externaldrive.DataStore, err *model.DashboardResponse) {
	uuid := mockUUID
	if pds != "googleDrive" && pds != "oneDrive" {
		var erro error
		dataStore, erro = externaldrive.StoreSessionData(data, uuid, cipherPassword)
		if erro != nil {
			err = &model.DashboardResponse{
				Code:         500,
				Message:      "Couldn't Create New DataStore and Encrypt It",
				ErrorMessage: erro.Error(),
			}
			return
		}
		return
	} else {
		err = &model.DashboardResponse{
			Code:    400,
			Message: "Bad PDS Variable",
		}
		return
	}
}

package services

import (
	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/externaldrive"
	"github.com/EC-SEAL/perseal/model"
)

var (
	mockUUID = "77358e9d-3c59-41ea-b4e3-04922657b30c" // As ID for DataStore
)

// Store Data on the corresponding PDS
func StoreCloudData(dto dto.PersistenceDTO, filename string) (dataStore *externaldrive.DataStore, err *model.HTMLResponse) {
	if dto.PDS == "googleDrive" {
		dataStore, err = storeSessionDataGoogleDrive(dto, filename)
	} else if dto.PDS == "oneDrive" {
		dataStore, err = storeSessionDataOneDrive(dto, filename)
	} else {
		err = &model.HTMLResponse{
			Code:    400,
			Message: "Wrong Module Or No Module Found in Credentials",
		}
		return
	}
	return
}

// Back-channel store may only be used for local Browser storing
func StoreLocalData(dto dto.PersistenceDTO) (dataStore *externaldrive.DataStore, err *model.HTMLResponse) {
	if dto.PDS != "googleDrive" && dto.PDS != "oneDrive" {
		dataStore, _ = externaldrive.StoreSessionData(dto)
		return
	} else {
		err = &model.HTMLResponse{
			Code:    400,
			Message: "Bad PDS Variable",
		}
		return
	}
}

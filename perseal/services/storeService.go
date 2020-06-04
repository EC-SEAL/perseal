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
func StoreCloudData(dto dto.PersistenceDTO, filename string) (returningdto dto.PersistenceDTO, dataStore *externaldrive.DataStore, err *model.DashboardResponse) {
	dto.UUID = mockUUID

	dto.SMResp, err = checkClientId(dto)
	if err != nil {
		return
	}
	if dto.PDS == "googleDrive" {
		returningdto, dataStore, err = storeSessionDataGoogleDrive(dto, filename)
	} else if dto.PDS == "oneDrive" {
		returningdto, dataStore, err = storeSessionDataOneDrive(dto, filename)
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
func StoreLocalData(dto dto.PersistenceDTO) (dataStore *externaldrive.DataStore, err *model.DashboardResponse) {
	dto.UUID = mockUUID
	if dto.PDS != "googleDrive" && dto.PDS != "oneDrive" {
		var erro error
		dataStore, erro = externaldrive.StoreSessionData(dto)
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

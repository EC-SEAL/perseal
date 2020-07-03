package services

import (
	"encoding/json"
	"log"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/externaldrive"
	"github.com/EC-SEAL/perseal/model"
)

var (
	mockUUID = "77358e9d-3c59-41ea-b4e3-04922657b30c" // As ID for DataStore
)

// Save session data to the configured persistence mechanism (front channel)
func PersistenceStore(dto dto.PersistenceDTO) (response, err *model.HTMLResponse) {
	log.Println("persistanceStore")

	if dto.PDS == "Mobile" || dto.PDS == "Browser" {
		dto.IsLocalLoad = true
		dataStore, _ := externaldrive.StoreSessionData(dto)
		data, _ := json.Marshal(dataStore)

		response = &model.HTMLResponse{
			Code:               200,
			Message:            "Stored DataStore " + dataStore.ID,
			ClientCallbackAddr: dto.ClientCallbackAddr,
			DataStore:          string(data),
		}
	} else if dto.PDS == "googleDrive" || dto.PDS == "oneDrive" {
		var dataStore *externaldrive.DataStore
		dataStore, err = storeCloudData(dto, "datastore.seal")
		if err != nil {
			return
		}
		response = &model.HTMLResponse{
			Code:               200,
			Message:            "Stored DataStore " + dataStore.ID,
			ClientCallbackAddr: dto.ClientCallbackAddr,
		}
	}
	return

}

// Store Data on the corresponding PDS
func storeCloudData(dto dto.PersistenceDTO, filename string) (dataStore *externaldrive.DataStore, err *model.HTMLResponse) {
	if dto.PDS == "googleDrive" {
		dataStore, err = storeSessionDataGoogleDrive(dto, filename)
	} else if dto.PDS == "oneDrive" {
		dataStore, err = storeSessionDataOneDrive(dto, filename)
	}
	return
}

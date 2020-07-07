package services

import (
	"encoding/json"
	"log"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/externaldrive"
	"github.com/EC-SEAL/perseal/model"
)

// Save session data to the configured persistence mechanism (front channel)
func PersistenceStore(dto dto.PersistenceDTO) (response, err *model.HTMLResponse) {
	log.Println("persistanceStore")

	if dto.PDS == model.EnvVariables.Mobile_PDS || dto.PDS == model.EnvVariables.Browser_PDS {
		dto.IsLocalLoad = true
		response, err = BackChannelStorage(dto)
	} else if dto.PDS == model.EnvVariables.Google_Drive_PDS || dto.PDS == model.EnvVariables.One_Drive_PDS {
		var dataStore *externaldrive.DataStore
		dataStore, err = storeCloudData(dto)
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

func BackChannelStorage(dto dto.PersistenceDTO) (response, err *model.HTMLResponse) {
	dataStore, erro := externaldrive.StoreSessionData(dto)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         500,
			Message:      "Encryption Failed",
			ErrorMessage: erro.Error(),
		}
		return
	}
	data, _ := json.Marshal(dataStore)

	response = &model.HTMLResponse{
		Code:               200,
		Message:            "Stored DataStore " + dataStore.ID,
		ClientCallbackAddr: dto.ClientCallbackAddr,
		DataStore:          string(data),
	}
	return
}

// Store Data on the corresponding PDS
func storeCloudData(dto dto.PersistenceDTO) (dataStore *externaldrive.DataStore, err *model.HTMLResponse) {
	if dto.PDS == model.EnvVariables.Google_Drive_PDS {
		dataStore, err = storeSessionDataGoogleDrive(dto)
	} else if dto.PDS == model.EnvVariables.One_Drive_PDS {
		dataStore, err = storeSessionDataOneDrive(dto)
	}
	return
}

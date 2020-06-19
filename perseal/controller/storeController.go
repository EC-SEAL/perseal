package controller

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/externaldrive"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/services"
)

// Save session data to the configured persistence mechanism (front channel)
func PersistenceStore(w http.ResponseWriter, r *http.Request) {
	log.Println("persistanceStore")
	dto, err := recieveSessionIdAndPassword(r)
	if err != nil {
		writeResponseMessage(w, dto, *err)
	}

	log.Println(dto.ID)
	log.Println(dto.PDS)

	if dto.PDS != "googleDrive" && dto.PDS != "oneDrive" {
		dto.IsLocal = true
		dto.StoreAndLoad = false
		dataStore, _ := externaldrive.StoreSessionData(dto)
		data, _ := json.Marshal(dataStore)

		log.Println(dataStore)
		log.Println(dto.ClientCallbackAddr)
		if err != nil {
			log.Println(err)
			writeResponseMessage(w, dto, *err)
		}
		response := model.HTMLResponse{
			Code:               200,
			Message:            "Stored DataStore " + dataStore.ID,
			ClientCallbackAddr: dto.ClientCallbackAddr,
			DataStore:          string(data),
		}
		writeResponseMessage(w, dto, response)
	} else {

		dto.IsLocal = false
		var dataStore *externaldrive.DataStore
		dataStore, err = services.StoreCloudData(dto, "datastore.seal")
		log.Println(dto.ClientCallbackAddr)
		if err != nil {
			writeResponseMessage(w, dto, *err)
		}

		response := model.HTMLResponse{
			Code:               200,
			Message:            "Stored DataStore " + dataStore.ID,
			ClientCallbackAddr: dto.ClientCallbackAddr,
		}
		writeResponseMessage(w, dto, response)
	}

}

package controller

import (
	"fmt"
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/externaldrive"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/services"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/EC-SEAL/perseal/utils"
	"github.com/gorilla/mux"
)

// Save session data to the configured persistence mechanism (front channel)
func PersistenceStore(w http.ResponseWriter, r *http.Request) {
	fmt.Println("persistentStore")

	id, err := utils.ReadRequestBody(r)
	if err != nil {
		w.WriteHeader(err.Code)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}

	sessionData, err := sm.GetSessionData(id, "")

	var dataStore *externaldrive.DataStore
	var redirect string

	// For Development
	if sessionData.SessionData.SessionVariables["ClientCallbackAddr"] == "" {
		sm.UpdateSessionData(id, "https://vm.project-seal.eu:9053/swagger-ui.html", "ClientCallbackAddr")
		sessionData, _ = sm.GetSessionData(id, "")
	}

	dto := dto.PersistenceDTO{
		ID:                 id,
		PDS:                sessionData.SessionData.SessionVariables["PDS"],
		Method:             "store",
		ClientCallbackAddr: sessionData.SessionData.SessionVariables["ClientCallbackAddr"],
		SMResp:             sessionData,
	}
	dto, dataStore, err = services.StoreCloudData(dto, "datastore.seal")

	if dto.StopProcess == true {
		w.Header().Set("content-type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w = utils.WriteResponseMessage(w, dto.ClientCallbackAddr, 200)
		return
	}
	fmt.Println(dataStore)
	fmt.Println(redirect)
	fmt.Println(err)

	if redirect != "" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w = utils.WriteResponseMessage(w, redirect, 302)
		return
	} else if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	} else if dataStore != nil {
		w.Header().Set("content-type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w = utils.WriteResponseMessage(w, dto.ClientCallbackAddr, 200)
		log.Println("url: ", dto.ClientCallbackAddr)
	}
	return
}

// Handles /per/store/{sessionToken} request
func PersistenceStoreWithToken(w http.ResponseWriter, r *http.Request) {
	log.Println("persistenceStoreWithToken")

	sessionToken := mux.Vars(r)["sessionToken"]
	if sessionToken == "" {
		err := &model.DashboardResponse{
			Code:    400,
			Message: "Couldn't find Session Token",
		}
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}

	sessionData, err := sm.GetSessionData(sessionToken, "")
	log.Println(sessionData)
	if err != nil {
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}

	if err := sm.ValidateSessionMngrResponse(sessionData); err != nil {
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}

	var cipherPassword string
	if keys, ok := r.URL.Query()["cipherPassword"]; ok {
		cipherPassword = keys[0]
	}

	dto := dto.PersistenceDTO{
		ID:                 sessionToken,
		PDS:                sessionData.SessionData.SessionVariables["PDS"],
		Method:             "store",
		ClientCallbackAddr: sessionData.SessionData.SessionVariables["ClientCallbackAddr"],
		SMResp:             sessionData,
		Password:           cipherPassword,
	}

	dataStore, err := services.StoreLocalData(dto)

	if err != nil {
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}
	if dataStore != nil {
		w = utils.WriteResponseMessage(w, dataStore, 201)
		return
	}

}

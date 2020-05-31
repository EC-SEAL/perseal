package controller

import (
	"fmt"
	"log"
	"net/http"

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

	id, sessionData, err := getSessionDataFromMSToken(r)
	if err != nil {
		w.WriteHeader(err.Code)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}

	pds := sessionData.SessionData.SessionVariables["PDS"]
	log.Println(pds)
	var dataStore *externaldrive.DataStore
	var redirect string

	_, dataStore, err = services.StoreCloudData(sessionData, pds, id, "datastore.seal", "store")

	fmt.Println(dataStore)
	fmt.Println(redirect)
	fmt.Println(err)

	url := sessionData.SessionData.SessionVariables["ClientCallbackAddr"]

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
		w = utils.WriteResponseMessage(w, url, 200)
		if model.CurrentUser != nil {
			model.CurrentUser = nil
		}
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
	pds := sessionData.SessionData.SessionVariables["PDS"]

	if keys, ok := r.URL.Query()["cipherPassword"]; ok {
		cipherPassword = keys[0]
	}

	dataStore, err := services.StoreLocalData(sessionData, pds, cipherPassword)

	if err != nil {
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}
	if dataStore != nil {
		w = utils.WriteResponseMessage(w, dataStore, 201)
		return
	}

}

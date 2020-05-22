package controller

import (
	"encoding/json"
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

// Handles /per/store request
func PersistenceStore(w http.ResponseWriter, r *http.Request) {
	fmt.Println("persistentStore")

	msToken := r.FormValue("msToken")
	if msToken == "" {
		err := &model.DashboardResponse{
			Code:    400,
			Message: "Couldn't find msToken",
		}
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}

	sessionId, err := sm.ValidateToken(msToken)
	if err != nil {
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}

	sessionData, err := sm.GetSessionData(sessionId, "")
	if err != nil {
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}

	if err := sm.ValidateSessionMngrResponse(sessionData); err != nil {
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}

	log.Println(sessionData)

	pds := sessionData.SessionData.SessionVariables["PDS"]
	log.Println(pds)

	var clientId string
	if pds == "googleDrive" {
		clientId := sessionData.SessionData.SessionVariables["GoogleDriveAccessCreds"]
		log.Println(clientId)
	} else if pds == "oneDrive" {
		clientId := sessionData.SessionData.SessionVariables["OneDriveAccessToken"]
		log.Println(clientId)
	}

	var dataStore *externaldrive.DataStore
	var redirect string

	model.Filename = make(chan model.File)
	filename := <-model.Filename
	log.Println(filename)

	_, dataStore, err = services.StoreCloudData(sessionData, pds, clientId, sessionData.SessionData.SessionID, filename.Filename)

	fmt.Println(dataStore)
	fmt.Println(redirect)
	fmt.Println(err)

	if redirect != "" {
		w = utils.WriteResponseMessage(w, redirect, 302)
		return
	} else if err != nil {
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	} else if dataStore != nil {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(201)
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

//Auxiliary Method for Development: Resets Session Variables of a given SessionId
func Reset(w http.ResponseWriter, r *http.Request) {
	sessionToken := r.FormValue("sessionToken")
	if sessionToken == "" {
		err := &model.DashboardResponse{
			Code:    400,
			Message: "Couldn't find Session Token",
		}
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}

	sm.UpdateSessionData(sessionToken, "{}", "")
	ti, _ := sm.GetSessionData(sessionToken, "")
	w.WriteHeader(200)
	t, _ := json.MarshalIndent(ti, "", "\t")
	w.Write(t)
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/sm"
	"github.com/gorilla/mux"
)

func persistenceStore(w http.ResponseWriter, r *http.Request) {
	fmt.Println("persistentStore")

	msToken := r.FormValue("msToken")
	smResp, err := sm.ValidateToken(msToken)
	sessionData, err := sm.GetSessionData(smResp.SessionData.SessionID, "")
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	if err := sm.ValidateSessionMngrResponse(sessionData); err != nil {
		http.Error(w, err.Error(), 401)
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

	_, redirect, err := storeData(sessionData, pds, clientId, "")

	if redirect != nil {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(302)
		t, _ := json.MarshalIndent(redirect, "", "\t")
		w.Write(t)
	} else {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(200)
	}

}

func persistenceStoreWithToken(w http.ResponseWriter, r *http.Request) {
	log.Println("persistenceStoreWithToken")

	sessionToken := mux.Vars(r)["sessionToken"]
	sessionData, err := sm.GetSessionData(sessionToken, "")

	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	}

	if err := sm.ValidateSessionMngrResponse(sessionData); err != nil {
		http.Error(w, err.Error(), 401)
		return
	}

	var cipherPassword string
	pds := sessionData.SessionData.SessionVariables["PDS"]
	googleAccessToken := sessionData.SessionData.SessionVariables["GoogleDriveAccessCreds"]

	if keys, ok := r.URL.Query()["cipherPassword"]; ok {
		cipherPassword = keys[0]
	}

	dataStore, redirect, err := storeData(sessionData, pds, googleAccessToken, cipherPassword)

	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	}

	if redirect != nil {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(302)
		t, _ := json.MarshalIndent(redirect, "", "\t")
		w.Write(t)
	}

	if dataStore != nil {
		t, _ := json.MarshalIndent(dataStore, "", "\t")
		w.Write(t)
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(200)
	}

}

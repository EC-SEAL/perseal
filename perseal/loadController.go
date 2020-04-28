package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/sm"
)

func persistenceLoad(w http.ResponseWriter, r *http.Request) {
	log.Println("persistanceLoad")

	msToken := r.FormValue("msToken")
	smResp, err := sm.ValidateToken(msToken)
	smResp2, err := sm.GetSessionData(smResp.SessionData.SessionID, "")

	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	} else {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(200)
		t, _ := json.MarshalIndent(smResp2, "", "\t")
		w.Write(t)
	}
}

//In development
/*
func persistenceLoadWithToken(w http.ResponseWriter, r *http.Request) {
	log.Println("persistenceLoadWithToken")

	sessionToken := mux.Vars(r)["sessionToken"]
	sessionData, err := sm.GetSessionData(sessionToken, "")

	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	}

	dataSstr := r.PostFormValue("dataStore")

	log.Println(sessionData)
	log.Println(dataSstr)
	if dataSstr == "" {
		http.Error(w, "required param `dataStore` not sent", 400) // TODO remove verbose
		return
	}

	var dataStore DataStore
	err = json.Unmarshal([]byte(dataSstr), &dataStore)
	if err != nil {
		http.Error(w, "dataStore in invalid format\n"+err.Error(), 422) // TODO remove verbose
		return
	}

	cipherPassword := mux.Vars(r)["cipherPassword"]

	log.Println(cipherPassword)

	data := dataStore.ClearData
	uuid := dataStore.ID
	sessionId := sessionData.SessionData.SessionID
	pds := sessionData.SessionData.SessionVariables["PDS"]

	if pds == "googleDrive" {
		clientId := sessionData.SessionData.SessionVariables["GoogleDriveClientID"]
		_, _, _ = storeData(sessionData, pds, clientId, cipherPassword)

	}
	if pds == "oneDrive" {
		clientId := sessionData.SessionData.SessionVariables["OneDriveClientID"]

		if clientId == "" {
			sessionData.Error = "Session Data Not Correctly Set - One Drive Client Missing"
			establishOneDriveCredentials(sessionId)
		}

		_, _ = storeFileOneDriveClearText(sessionData, uuid, cipherPassword, data)
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(200)
}
*/

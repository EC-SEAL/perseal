package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/EC-SEAL/perseal/sm"
	"github.com/gorilla/mux"
)

var (
	mockUUID = "77358e9d-3c59-41ea-b4e3-04922657b30c" // As ID for DataStore
)

// persistenceStore handles /per/store request - Save session data to the configured persistence mechanism (front channel).
func persistenceStore(w http.ResponseWriter, r *http.Request) {
	fmt.Println("persistentStore")
	msToken := r.FormValue("msToken")
	platform := mux.Vars(r)["type"]
	sessionData, err := sm.ValidateToken(msToken)
	log.Println("Validated token and got sessionData: ", sessionData.SessionData)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	if err := sm.ValidateSessionMngrResponse(sessionData, msToken); err != nil {
		http.Error(w, err.Error(), 401)
		return
	}

	// Data to be stored
	data := sessionData
	uuid := mockUUID
	if platform == os.Getenv("GOOGLE_DRIVE") {
		//Validates if the session data contains the google drive authentication token
		//	if data.SessionData.SessionVariables.GoogleDrive == "" {
		//		sessionData.Error = "Session Data Not Correctly Set - Google Drive Oauth Missing"
		//		http.Error(w, sessionData.Error, 401)
		//		return
		//	}

		_, err = storeSessionDataGoogleDrive(data, uuid, "password") // No password
		if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}
	}
	if platform == os.Getenv("ONE_DRIVE") {
		//Validates if the session data contains the google drive authentication token
		if data.SessionData.SessionVariables.OneDriveClient == "" {
			sessionData.Error = "Session Data Not Correctly Set - One Drive Oauth Missing"
			http.Error(w, sessionData.Error, 401)
			log.Fatalln(sessionData.Error)
			data.SessionData.SessionVariables.OneDriveClient = os.Getenv("ONE_DRIVE_CLIENT")
			data.SessionData.SessionVariables.OneDriveScopes = os.Getenv("ONE_DRIVE_SCOPES")
			//	return
		}

		_, err = storeSessionDataOneDrive(data, uuid, "password") // No password
		if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(200)
}

// Save session data to the configured persistence mechanism (back channel). Might return the signed and possibly encrypted datastore
func persistenceStoreWithToken(w http.ResponseWriter, r *http.Request) {
	log.Println("persistenceStoreWithToken")
	sessionToken := mux.Vars(r)["sessionToken"]
	platform := mux.Vars(r)["type"]
	_, err := sm.ValidateToken(sessionToken)
	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	}
	//	log.Println("Validated session: ", session)
	sessionData, err := sm.GetSessionData(sessionToken, "") // TODO what is variable name?
	//Probably could be error message? Review Later
	log.Println("Validated token and got sessionData: ", sessionData.SessionData)
	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	}
	if err := sm.ValidateSessionMngrResponse(sessionData, sessionToken); err != nil {
		http.Error(w, err.Error(), 401)
		return
	}

	var cipherPassword string
	if keys, ok := r.URL.Query()["cipherPassword"]; ok {
		cipherPassword = keys[0]
	}

	// Data to be stored
	data := sessionData
	uuid := mockUUID
	var dataStore *DataStore
	if platform == os.Getenv("GOOGLE_DRIVE") {
		dataStore, err = storeSessionDataGoogleDrive(data, uuid, cipherPassword)
		if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}
	}
	if platform == os.Getenv("ONE_DRIVE") {
		if data.SessionData.SessionVariables.OneDriveClient == "" {
			sessionData.Error = "Session Data Not Correctly Set - One Drive Oauth Missing"
			http.Error(w, sessionData.Error, 401)
			log.Fatalln(sessionData.Error)
			data.SessionData.SessionVariables.OneDriveClient = os.Getenv("ONE_DRIVE_CLIENT")
			data.SessionData.SessionVariables.OneDriveScopes = os.Getenv("ONE_DRIVE_SCOPES")
			//	return
		}
		dataStore, err = storeSessionDataOneDrive(data, uuid, cipherPassword) // No password
		if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(200)
	t, err := json.MarshalIndent(dataStore, "", "\t")
	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	}
	w.Write(t)
}

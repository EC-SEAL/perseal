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

// persistenceLoad handles /per/load request
func persistenceLoad(w http.ResponseWriter, r *http.Request) {
	msToken := r.FormValue("msToken")
	//fmt.Fprintf(w, "persistenceLoad: %v", msToken)
	smResp, err := sm.ValidateToken(msToken)
	if err != nil {
		fmt.Fprintf(w, "error: %v", err)
	} else {
		fmt.Fprintf(w, "persistenceLoad: %v", smResp)
	}
}

/*  Mock dataStore value
{
  "id": "6c0f70a8-f32b-4535-b5f6-0d596c52813a",
  "encryptedData": "string",
  "signature": "string",
  "signatureAlgorithm": "string",
  "encryptionAlgorithm": "string",
  "clearData": [
    {
      "id": "6c0f70a8-f32b-4535-b5f6-0d596c52813a",
      "type": "string",
      "categories": [
        "string"
      ],
      "issuerId": "string",
      "subjectId": "string",
      "loa": "string",
      "issued": "2018-12-06T19:40:16Z",
      "expiration": "2018-12-06T19:45:16Z",
      "attributes": [
        {
          "name": "http://eidas.europa.eu/attributes/naturalperson/CurrentGivenName",
          "friendlyName": "CurrentGivenName",
          "encoding": "plain",
          "language": "ES_es",
          "isMandatory": true,
          "values": [
            "JOHN"
          ]
        }
      ],
      "properties": {
        "additionalProp1": "string",
        "additionalProp2": "string",
        "additionalProp3": "string"
      }
    }
  ]
}
*/

func persistenceLoadWithToken(w http.ResponseWriter, r *http.Request) {
	log.Println("persistenceLoadWithToken")
	sessionToken := mux.Vars(r)["sessionToken"]
	platform := mux.Vars(r)["type"]
	sessionData, err := sm.ValidateToken(sessionToken)
	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	}
	log.Println("Validated token: ", sessionData.SessionData)

	dataSstr := r.PostFormValue("dataStore")
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
	if err := sm.ValidateSessionMngrResponse(sessionData, sessionToken); err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	cipherPassword := r.PostFormValue("cipherPassword")
	// Data to be stored - in THIS case it is ClearData that comes from POST
	session := sessionData
	data := dataStore.ClearData
	uuid := dataStore.ID
	if platform == os.Getenv("GOOGLE_DRIVE") {
		_, err = storeSessionDataGoogleDrive(data, uuid, cipherPassword)
		if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}
	}
	if platform == os.Getenv("ONE_DRIVE") {
		if sessionData.SessionData.SessionVariables.OneDriveClient == "" {
			sessionData.Error = "Session Data Not Correctly Set - One Drive Oauth Missing"
			http.Error(w, sessionData.Error, 401)
			log.Fatalln(sessionData.Error)
			sessionData.SessionData.SessionVariables.OneDriveClient = os.Getenv("ONE_DRIVE_CLIENT")
			sessionData.SessionData.SessionVariables.OneDriveScopes = os.Getenv("ONE_DRIVE_SCOPES")
			//	return
		}
		_, err = storeFileOneDriveClearText(session, uuid, cipherPassword, data)
		if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(200)
}

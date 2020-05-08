package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/services"
	"github.com/EC-SEAL/perseal/sm"
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
		t, erro := json.MarshalIndent(err, "", "\t")
		if erro != nil {
			http.Error(w, erro.Error(), 404)
		}
		w.Write(t)
		return
	}

	sessionId, err := sm.ValidateToken(msToken)
	if err != nil {
		t, erro := json.MarshalIndent(err, "", "\t")
		if erro != nil {
			http.Error(w, erro.Error(), 404)
		}
		w.Write(t)
		return
	}

	sessionData, err := sm.GetSessionData(sessionId, "")
	if err != nil {
		t, erro := json.MarshalIndent(err, "", "\t")
		if erro != nil {
			http.Error(w, erro.Error(), 404)
		}
		w.Write(t)
		return
	}

	if err := sm.ValidateSessionMngrResponse(sessionData); err != nil {
		t, erro := json.MarshalIndent(err, "", "\t")
		if erro != nil {
			http.Error(w, erro.Error(), 404)
		}
		w.Write(t)
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

	dataStore, redirect, err := services.StoreData(sessionData, pds, clientId, "")
	fmt.Println(dataStore)
	fmt.Println(redirect)
	fmt.Println(err)
	if redirect != nil {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(302)
		t, err := json.MarshalIndent(redirect, "", "\t")
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		w.Write(t)
		return
	} else if err != nil {
		t, erro := json.MarshalIndent(err, "", "\t")
		if erro != nil {
			http.Error(w, erro.Error(), 404)
		}
		w.Write(t)
		return
	} else if dataStore != nil {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(201)
	}

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
		t, erro := json.MarshalIndent(err, "", "\t")
		if erro != nil {
			http.Error(w, erro.Error(), 404)
		}
		w.Write(t)
		return

	}

	sessionData, err := sm.GetSessionData(sessionToken, "")
	log.Println(sessionData)
	if err != nil {
		t, erro := json.MarshalIndent(err, "", "\t")
		if erro != nil {
			http.Error(w, erro.Error(), 404)
		}
		w.Write(t)
		return
	}

	if err := sm.ValidateSessionMngrResponse(sessionData); err != nil {
		t, erro := json.MarshalIndent(err, "", "\t")
		if erro != nil {
			http.Error(w, erro.Error(), 404)
		}
		w.Write(t)
		return
	}

	var cipherPassword string
	var clientId string
	pds := sessionData.SessionData.SessionVariables["PDS"]
	if pds == "googleDrive" {
		clientId = sessionData.SessionData.SessionVariables["GoogleDriveAccessCreds"]
		log.Println(clientId)
	} else if pds == "oneDrive" {
		clientId = sessionData.SessionData.SessionVariables["OneDriveAccessToken"]
		log.Println(clientId)
	}

	if keys, ok := r.URL.Query()["cipherPassword"]; ok {
		cipherPassword = keys[0]
	}

	dataStore, redirect, err := services.StoreData(sessionData, pds, clientId, cipherPassword)

	if err != nil {
		t, erro := json.MarshalIndent(err, "", "\t")
		if erro != nil {
			http.Error(w, erro.Error(), 404)
		}
		w.WriteHeader(err.Code)
		w.Write(t)
		return
	}

	if redirect != nil {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(302)
		t, _ := json.MarshalIndent(redirect, "", "\t")
		w.Write(t)
	}

	if dataStore != nil {
		t, err := json.MarshalIndent(dataStore, "", "\t")
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		w.Write(t)
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(201)
	}

}

// Recieve Request From Dashboard To Retreive a Cloud Token by providing with a code
func GetCodeFromDashboard(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	if code == "" {
		err := &model.DashboardResponse{
			Code:    400,
			Message: "Couldn't find Code",
		}
		w.WriteHeader(err.Code)
		t, erro := json.MarshalIndent(err, "", "\t")
		if erro != nil {
			http.Error(w, erro.Error(), 404)
		}
		w.Write(t)
		return
	}

	sessionID := r.FormValue("sessionId")
	if sessionID == "" {
		err := &model.DashboardResponse{
			Code:    400,
			Message: "Couldn't find Session Id",
		}
		w.WriteHeader(err.Code)
		t, erro := json.MarshalIndent(err, "", "\t")
		if erro != nil {
			http.Error(w, erro.Error(), 404)
		}
		w.Write(t)
		return
	}

	module := r.FormValue("module")
	if module == "" {
		err := &model.DashboardResponse{
			Code:    400,
			Message: "Couldn't find Module",
		}
		w.WriteHeader(err.Code)
		t, erro := json.MarshalIndent(err, "", "\t")
		if erro != nil {
			http.Error(w, erro.Error(), 404)
		}
		w.Write(t)
		return
	}
	err := services.PersistenceStoreWithCode(code, sessionID, module)
	if err != nil {
		w.WriteHeader(err.Code)
		t, erro := json.MarshalIndent(err, "", "\t")
		if erro != nil {
			http.Error(w, erro.Error(), 404)
		}
		w.Write(t)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(201)
}

//Auxiliary Method for Development: Resets Session Variables of a given SessionId
func Reset(w http.ResponseWriter, r *http.Request) {
	sessionToken := r.FormValue("sessionToken")
	if sessionToken == "" {
		err := &model.DashboardResponse{
			Code:    400,
			Message: "Couldn't find Session Token",
		}
		t, erro := json.MarshalIndent(err, "", "\t")
		if erro != nil {
			http.Error(w, erro.Error(), 404)
		}
		w.Write(t)
		return

	}
	sm.UpdateSessionData(sessionToken, "{}", "")
	ti, _ := sm.GetSessionData(sessionToken, "")
	w.WriteHeader(200)
	t, _ := json.MarshalIndent(ti, "", "\t")
	w.Write(t)
}

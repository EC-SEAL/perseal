package controller

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/externaldrive"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/services"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/gorilla/mux"
)

func PersistenceLoad(w http.ResponseWriter, r *http.Request) {
	log.Println("persistanceLoad")

	msToken := r.FormValue("msToken")
	if msToken == "" {
		errorToDash := &model.DashboardResponse{
			Code:    404,
			Message: "msToken not Found",
		}
		t, err := json.MarshalIndent(errorToDash, "", "\t")
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		w.Write(t)
		w.WriteHeader(errorToDash.Code)
		return
	}

	id, err := sm.ValidateToken(msToken)
	if err != nil {
		w.WriteHeader(err.Code)
		t, err := json.MarshalIndent(err, "", "\t")
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		w.Write(t)
		return
	}

	smResp2, err := sm.GetSessionData(id, "")
	if err != nil {
		w.WriteHeader(err.Code)
		t, err := json.MarshalIndent(err, "", "\t")
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		w.Write(t)
		return
	}

	pds := smResp2.SessionData.SessionVariables["PDS"]
	log.Println(smResp2)

	ds, err := services.FetchDataStore(pds, smResp2)
	if err != nil {
		w.WriteHeader(err.Code)
		t, err := json.MarshalIndent(err, "", "\t")
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		w.Write(t)
	} else if ds == nil {
		errorToDash := &model.DashboardResponse{
			Code:    404,
			Message: "dataStore not Found",
		}
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(404)
		t, err := json.MarshalIndent(errorToDash, "", "\t")
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		w.Write(t)
	} else {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(200)
		t, err := json.MarshalIndent(ds, "", "\t")
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		w.Write(t)
	}

}

func PersistenceLoadWithToken(w http.ResponseWriter, r *http.Request) {
	log.Println("persistenceLoadWithToken")
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
	if err != nil {
		t, err := json.MarshalIndent(err, "", "\t")
		if err != nil {
			http.Error(w, err.Error(), 404)
		}
		w.Write(t)
	}

	dataSstr := r.PostFormValue("dataStore")
	if dataSstr == "" {
		err := &model.DashboardResponse{
			Code:    400,
			Message: "Couldn't find DataStore",
		}
		t, erro := json.MarshalIndent(err, "", "\t")
		if erro != nil {
			http.Error(w, erro.Error(), 400)
		}
		w.Write(t)
		return
	}

	log.Println(sessionData)
	log.Println(dataSstr)

	var dataStore externaldrive.DataStore
	erro := json.Unmarshal([]byte(dataSstr), &dataStore)
	if erro != nil {
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

	cipherPassword := r.FormValue("cipherPassword")
	if cipherPassword == "" {
		err = &model.DashboardResponse{
			Code:         404,
			Message:      "Couldn't Parse Response Body from Get Session Data to Object",
			ErrorMessage: erro.Error(),
		}
		t, erro := json.MarshalIndent(err, "", "\t")
		if erro != nil {
			http.Error(w, erro.Error(), 404)
		}
		w.Write(t)
		return
	}

	log.Println(cipherPassword)
	erro = dataStore.Decrypt(cipherPassword)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         404,
			Message:      "Couldn't Decrypt DataStore",
			ErrorMessage: erro.Error(),
		}
		t, erro := json.MarshalIndent(err, "", "\t")
		if erro != nil {
			http.Error(w, erro.Error(), 404)
		}
		w.Write(t)
		return
	}

	jsonM, erro := json.Marshal(dataStore)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         404,
			Message:      "Couldn't Parse Response Body from Get Session Data to Object",
			ErrorMessage: erro.Error(),
		}
		t, erro := json.MarshalIndent(err, "", "\t")
		if erro != nil {
			http.Error(w, erro.Error(), 404)
		}
		w.Write(t)
		return
	}

	_, err = sm.UpdateSessionData(sessionToken, string(jsonM), "dataStore")
	if err != nil {
		t, erro := json.MarshalIndent(err, "", "\t")
		if erro != nil {
			http.Error(w, erro.Error(), 404)
		}
		w.Write(t)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(200)
}

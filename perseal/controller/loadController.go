package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/EC-SEAL/perseal/externaldrive"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/services"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/gorilla/mux"
)

// see https://github.com/EC-SEAL/interface-specs/blob/master/images/UC8_06_SP_Attribute_Retrieval_from_Browser_PDS_v2.png how does generateToken works

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

	var ds *externaldrive.DataStore
	var msTokenResp string
	if pds == "googleDrive" || pds == "oneDrive" {
		ds, err = services.FetchCloudDataStore(pds, smResp2)
		fmt.Println(err)
	} else if pds == "googleDrive" || pds == "oneDrive" {
		services.FetchLocalDataStore(pds, smResp2)
	}

	smRes, err := sm.GenerateToken("", "PERms001", "PERms001", id)
	if err != nil {
		w.WriteHeader(err.Code)
		t, err := json.MarshalIndent(err, "", "\t")
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		w.Write(t)
	}

	msTokenResp = smRes.AdditionalData
	if model.Local {
		err = services.DecryptAndMarshallDataStore(ds, id, "qwerty")
	} else {
		err = services.DecryptAndMarshallDataStore(ds, id, os.Getenv("PASS"))
	}
	if err != nil {
		w.WriteHeader(err.Code)
		t, err := json.MarshalIndent(err, "", "\t")
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		w.Write(t)
	}
	if ds == nil {
		w.WriteHeader(200)
		msTokenResp2 := msTokenResp + " DataStore Not Found"
		t, err := json.MarshalIndent(msTokenResp2, "", "\t")
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		w.Write(t)
	} else {
		t, err := json.MarshalIndent(ds, "", "\t")
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(200)
		w.Write(t)
	}

}

//see https://github.com/EC-SEAL/interface-specs/blob/master/images/UC8_03_SP_Attribute_Retrieval_from_Mobile_PDS_v5.png confusing

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
			Message: "Couldn't Unmarshal DataStore",
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

	err = services.DecryptAndMarshallDataStore(&dataStore, sessionToken, cipherPassword)
	if err != nil {
		t, erro := json.MarshalIndent(err, "", "\t")
		if erro != nil {
			http.Error(w, erro.Error(), 404)
		}
		w.Write(t)
		return
	}
	w.Write([]byte(sessionToken))
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(200)
}

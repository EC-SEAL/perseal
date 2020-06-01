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

// Setup a persistence mechanism and load a secure storage into session.
func PersistenceLoad(w http.ResponseWriter, r *http.Request) {
	log.Println("persistanceLoad")

	id, err := utils.ReadRequestBody(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(err.Code)
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}

	smResp, err := sm.GetSessionData(id, "")

	// For Development
	if smResp.SessionData.SessionVariables["ClienCallbackAddr"] == "" {
		sm.UpdateSessionData(id, "https://vm.project-seal.eu:9053/swagger-ui.html", "ClientCallbackAddr")
		smResp, _ = sm.GetSessionData(id, "")
	}

	// Initialize Variables
	var ds *externaldrive.DataStore
	var fetchedFromLocalData bool
	pds := smResp.SessionData.SessionVariables["PDS"]
	clientCallBack := smResp.SessionData.SessionVariables["ClientCallbackAddr"]
	var password, clientCallBackVerify string

	if pds == "googleDrive" || pds == "oneDrive" {
		ds, err = services.FetchCloudDataStore(smResp, pds, "datastore.seal")
	} else if pds == "Browser" || pds == "Mobile" {
		fetchedFromLocalData = services.FetchLocalDataStore(pds, clientCallBack, smResp)
	}

	// For UC 1.06. If no Files ares found, perform a store
	if err != nil {
		log.Println(err.Code)
		if err.Code == 302 {
			fmt.Println("No DataStore Found! Performing Store")
			password, ds, err = services.StoreCloudData(smResp, pds, id, "datastore.seal", "load&store")
			log.Println(ds)
			log.Println(err)
		}

	}

	// Validates signature of DataStore
	if !services.ValidateSignature(ds.EncryptedData, ds.Signature) {
		errorToDash := &model.DashboardResponse{
			Code:    500,
			Message: "Error Validating Signature",
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w = utils.WriteResponseMessage(w, errorToDash, errorToDash.Code)
		return
	}

	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}

	if fetchedFromLocalData && clientCallBack == clientCallBackVerify {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w = utils.WriteResponseMessage(w, smResp.SessionData.SessionID, 200)
		return

	} else if fetchedFromLocalData && clientCallBack != clientCallBackVerify {
		smRes, err := sm.GenerateToken("", "PERms001", "PERms001", id)
		if err != nil {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w = utils.WriteResponseMessage(w, err, err.Code)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w = utils.WriteResponseMessage(w, smRes.AdditionalData, 200)
		return

	} else {
		if err != nil {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w = utils.WriteResponseMessage(w, err, err.Code)
			return
		}

		// Waits for Password
		if password == "" {
			model.Password = make(chan string)
			password = <-model.Password
			log.Println(password)
			model.Password = nil
		}

		err = services.DecryptAndMarshallDataStore(ds, id, password)

		if err != nil {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w = utils.WriteResponseMessage(w, err, err.Code)
			return
		}

		w.Header().Set("content-type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w = utils.WriteResponseMessage(w, clientCallBack, 200)
	}
	return
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
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}

	sessionData, err := sm.GetSessionData(sessionToken, "")
	if err != nil {
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}

	dataSstr := r.PostFormValue("dataStore")
	if dataSstr == "" {
		err := &model.DashboardResponse{
			Code:    400,
			Message: "Couldn't find DataStore",
		}
		w = utils.WriteResponseMessage(w, err, err.Code)
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
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}

	cipherPassword := r.FormValue("cipherPassword")
	if cipherPassword == "" {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Couldn't Parse Response Body from Get Session Data to Object",
			ErrorMessage: erro.Error(),
		}
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}

	err = services.DecryptAndMarshallDataStore(&dataStore, sessionToken, cipherPassword)
	if err != nil {
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}
	w.Write([]byte(sessionToken))
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(200)
}

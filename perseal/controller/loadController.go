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
	"github.com/EC-SEAL/perseal/utils"
	"github.com/gorilla/mux"
)

// Setup a persistence mechanism and load a secure storage into session.
func PersistenceLoad(w http.ResponseWriter, r *http.Request) {
	log.Println("persistanceLoad")

	msToken := r.FormValue("msToken")
	if msToken == "" {
		errorToDash := &model.DashboardResponse{
			Code:    404,
			Message: "msToken not Found",
		}
		w = utils.WriteResponseMessage(w, errorToDash, errorToDash.Code)
		return
	}

	id, smResp, err := utils.GetSessionDataFromMSToken(msToken)
	if err != nil {
		w.WriteHeader(err.Code)
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}

	// Initialize Variables
	var ds *externaldrive.DataStore
	var fetchedFromLocalData bool
	pds := smResp.SessionData.SessionVariables["PDS"]
	clientCallBack := smResp.SessionData.SessionVariables["ClientCallbackAddr"]
	var password, clientCallBackVerify string

	//Send Current User to UI
	sm.CurrentUser = make(chan sm.SessionMngrResponse)
	sm.CurrentUser <- smResp

	//Request Filename
	model.Filename = make(chan model.File)
	filename := <-model.Filename
	log.Println(filename)

	if pds == "googleDrive" || pds == "oneDrive" {
		ds, err = services.FetchCloudDataStore(pds, smResp, &filename)
	} else if pds == "Browser" || pds == "Mobile" {
		fetchedFromLocalData = services.FetchLocalDataStore(pds, clientCallBack, smResp)
	}

	if err != nil {
		if err.Code == 302 {
			fmt.Println("No DataStore Found! Performing Store")
			password, ds, err = services.StoreCloudData(smResp, pds, id, filename.Filename)
			log.Println(ds)
			log.Println(err)
		}
	}

	if err != nil {
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}

	if model.Local {
		clientCallBackVerify = "https://vm.project-seal.eu:9053"
	} else {
		clientCallBackVerify = os.Getenv("CLIENT_CALLBACK_VERIFY")
	}

	if fetchedFromLocalData && clientCallBack == clientCallBackVerify {
		w = utils.WriteResponseMessage(w, smResp.SessionData.SessionID, 200)
		return

	} else if fetchedFromLocalData && clientCallBack != clientCallBackVerify {
		smRes, err := sm.GenerateToken("", "PERms001", "PERms001", id)
		if err != nil {
			w = utils.WriteResponseMessage(w, err, err.Code)
			return
		}
		w = utils.WriteResponseMessage(w, smRes.AdditionalData, 200)
		return

	} else {
		if err != nil {
			w = utils.WriteResponseMessage(w, err, err.Code)
			return
		}

		// Waits for Password
		if password == "" {
			model.Password = make(chan string)
			password = <-model.Password
			log.Println(password)
			close(model.Password)
		}

		err = services.DecryptAndMarshallDataStore(ds, id, password)

		if err != nil {
			w = utils.WriteResponseMessage(w, err, err.Code)
			return
		}

		w.Header().Set("content-type", "application/json")
		w = utils.WriteResponseMessage(w, ds, 200)
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
			Code:         404,
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

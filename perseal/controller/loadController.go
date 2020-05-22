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

// see https://github.com/EC-SEAL/interface-specs/blob/master/images/UC8_06_SP_Attribute_Retrieval_from_Browser_PDS_v2.png how does generateToken works

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

	smRes, err := sm.GenerateToken("", "PERms001", "PERms001", "70e26ae7-2687-4cc4-a3f2-ae1ab7ff1f6e")
	msToken = smRes.AdditionalData

	id, err := sm.ValidateToken(msToken)
	if err != nil {
		w.WriteHeader(err.Code)
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}

	smResp2, err := sm.GetSessionData(id, "")
	if err != nil {
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}

	pds := smResp2.SessionData.SessionVariables["PDS"]
	clientCallBack := smResp2.SessionData.SessionVariables["ClientCallbackAddr"]
	log.Println(smResp2)

	var ds *externaldrive.DataStore
	var fetchedFromLocalData bool

	if pds == "googleDrive" || pds == "oneDrive" {
		ds, err = services.FetchCloudDataStore(pds, smResp2)
		fmt.Println(err)
	} else if pds == "Browser" || pds == "Mobile" {
		fetchedFromLocalData = services.FetchLocalDataStore(pds, clientCallBack, smResp2)
	}

	if err != nil {
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}

	var clientCallBackVerify string
	if model.Local {
		clientCallBackVerify = "https://vm.project-seal.eu:9053"
	} else {
		clientCallBackVerify = os.Getenv("CLIENT_CALLBACK_VERIFY")
	}

	if fetchedFromLocalData && clientCallBack == clientCallBackVerify {
		w = utils.WriteResponseMessage(w, smResp2.SessionData.SessionID, 200)
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

		model.Password = make(chan string)
		password := <-model.Password
		log.Println(password)
		close(model.Password)
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

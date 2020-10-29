package controller

import (
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/services"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/EC-SEAL/perseal/utils"
	"github.com/gorilla/mux"
)

//Back-Channel request to Decrypt and Load User's Data
func BackChannelLoading(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)

	method := mux.Vars(r)["method"]
	msToken := r.FormValue("msToken")
	cipherPassword := getQueryParameter(r, "cipherPassword")
	if cipherPassword == "" {
		cipherPassword = r.FormValue("cipherPassword")
	}
	if model.Test {
		cipherPassword = utils.HashSUM256(cipherPassword)
	}

	dto, _, err := initialEPSetup(w, msToken, method, true)
	if err != nil {
		return
	}

	dataSstr := r.PostFormValue("dataStore")
	if dataSstr == "" {
		err := model.BuildResponse(http.StatusBadRequest, model.Messages.FailedFoundDataStore)
		dto.Response = *err
		sm.UpdateSessionData(dto.ID, err.Message+"!\n"+err.ErrorMessage, model.EnvVariables.SessionVariables.FinishedPersealBackChannel)
		writeBackChannelResponse(dto, w)
		return
	}

	dto.Password = cipherPassword
	if dto.Password == "" {
		err := model.BuildResponse(http.StatusBadRequest, model.Messages.NoPassword)
		w.WriteHeader(err.Code)
		w.Write([]byte(err.Message))
		return
	}

	response, err := services.BackChannelDecryption(dto, dataSstr)
	if err != nil {
		if err.FailedInput == "Password" {
			err := model.BuildResponse(http.StatusBadRequest, model.Messages.InvalidPassword)
			w.WriteHeader(err.Code)
			w.Write([]byte(err.Message))
			return
		}

		dto.Response = *err
		sm.UpdateSessionData(dto.ID, err.Message+"!\n"+err.ErrorMessage, model.EnvVariables.SessionVariables.FinishedPersealBackChannel)

	} else {
		dto.Response = *response
		sm.UpdateSessionData(dto.ID, model.Messages.LoadedDataStore, model.EnvVariables.SessionVariables.FinishedPersealBackChannel)
	}

	writeBackChannelResponse(dto, w)
	return
}

// Back-Channel Request to Encrypt User's Data
func backChannelStoring(w http.ResponseWriter, id, cipherPassword, method string, smResp sm.SessionMngrResponse) {
	obj, err := dto.PersistenceFactory(id, smResp, method)
	log.Println("Current Persistence Object: ", obj)

	obj.Password = cipherPassword
	if err != nil {
		obj.Response = *err
		writeBackChannelResponse(obj, w)
		return
	}

	response, err := services.BackChannelStorage(obj)
	if err != nil {
		obj.Response = *err
		sm.UpdateSessionData(obj.ID, err.Message+"!\n"+err.ErrorMessage, model.EnvVariables.SessionVariables.FinishedPersealBackChannel)
	} else {
		obj.Response = *response
		sm.UpdateSessionData(obj.ID, model.Messages.StoredDataStore, model.EnvVariables.SessionVariables.FinishedPersealBackChannel)
	}

	writeBackChannelResponse(obj, w)
	return
}

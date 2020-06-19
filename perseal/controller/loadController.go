package controller

import (
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/externaldrive"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/services"
)

// Setup a persistence mechanism and load a secure storage into session.
func PersistenceLoad(w http.ResponseWriter, r *http.Request) {
	log.Println("persistanceLoad")
	dto, err := recieveSessionIdAndPassword(r)
	if err != nil {
		writeResponseMessage(w, dto, *err)
	}

	ds := &externaldrive.DataStore{}
	log.Println(dto.SMResp.SessionData.SessionVariables["dataStore"])

	// Initialize Variables
	if dto.PDS == "googleDrive" || dto.PDS == "oneDrive" {
		ds, err = services.FetchCloudDataStore(dto, "datastore.seal")
	} else if dto.PDS == "Browser" || dto.PDS == "Mobile" {
		ds = services.FetchLocalDataStore(r)
	}

	if err != nil {
		writeResponseMessage(w, dto, *err)
	}

	// Validates signature of DataStore
	if !services.ValidateSignature(ds.EncryptedData, ds.Signature) {
		errorToDash := &model.HTMLResponse{
			Code:    500,
			Message: "Error Validating Signature",
		}
		writeResponseMessage(w, dto, *errorToDash)
	}
	/*
		if fetchedFromLocalData && dto.ClientCallbackAddr == clientCallBackVerify {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w = writeResponseMessage(w, dto.ID, 200)
			return

		} else if fetchedFromLocalData && dto.ClientCallbackAddr != clientCallBackVerify {
			smRes, err := sm.GenerateToken("", "PERms001", "PERms001", dto.ID)
			if err != nil {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w = writeResponseMessage(w, err, err.Code)
				return
			}
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w = writeResponseMessage(w, smRes.AdditionalData, 200)
			return

		} else {
	*/
	err = services.DecryptAndMarshallDataStore(ds, dto)
	log.Println(ds)

	if err != nil {
		writeResponseMessage(w, dto, *err)
	}
	response := model.HTMLResponse{
		Code:               200,
		Message:            "Loaded DataStore " + ds.ID,
		ClientCallbackAddr: dto.ClientCallbackAddr,
	}
	writeResponseMessage(w, dto, response)

	return
}

// UC 1.06 - Stores and Loads Datastore
func PersistenceStoreAndLoad(w http.ResponseWriter, r *http.Request) {

	dto, err := recieveSessionIdAndPassword(r)
	if err != nil {
		writeResponseMessage(w, dto, *err)
	}

	log.Println(dto.SMResp.SessionData.SessionVariables["dataStore"])
	log.Println(dto.ID)
	log.Println(dto.PDS)

	ds, err := services.StoreCloudData(dto, "datastore.seal")
	if err != nil {
		writeResponseMessage(w, dto, *err)
	}

	// Validates signature of DataStore
	if !services.ValidateSignature(ds.EncryptedData, ds.Signature) {
		errorToDash := &model.HTMLResponse{
			Code:    500,
			Message: "Error Validating Signature",
		}
		writeResponseMessage(w, dto, *errorToDash)
	}

	err = services.DecryptAndMarshallDataStore(ds, dto)
	log.Println(ds)
	if err != nil {
		writeResponseMessage(w, dto, *err)
	}

	response := model.HTMLResponse{
		Code:    200,
		Message: "Loaded DataStore " + ds.ID,
	}
	writeResponseMessage(w, dto, response)
}

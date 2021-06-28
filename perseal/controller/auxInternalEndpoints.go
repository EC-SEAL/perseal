package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/services"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/gorilla/mux"
)

func AuxiliaryEndpoints(w http.ResponseWriter, r *http.Request) {

	method := mux.Vars(r)["method"]

	id := getQueryParameter(r, "sessionId")
	smResp := getSessionData(id, w)

	if method == "save" {

		log.Println(r.URL.Path)

		//Downloads File for the localFile System
		log.Println("save")
		contents := getQueryParameter(r, "contents")
		w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(smResp.SessionData.SessionVariables["DSFilename"]+model.EnvVariables.DataStore_File_Ext))
		w.Header().Set("Content-Type", "application/octet-stream")
		json.NewEncoder(w).Encode(contents)

		return

	} else if method == "checkQrCodePoll" {

		finishedPersealBackChannel := smResp.SessionData.SessionVariables[model.EnvVariables.SessionVariables.FinishedPersealBackChannel]

		if finishedPersealBackChannel == "not finished" {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Operation Not Yet Finished"))
		} else if finishedPersealBackChannel == "" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Session Variable Not Set"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(finishedPersealBackChannel))
		}
		return

	} else if method == "qrCodePoll" {

		op := getQueryParameter(r, "operation")

		respMethod, dto, err := services.QRCodePoll(id, op)
		if err != nil {
			writeResponseMessage(w, dto, *err)
		}

		var resp *model.HTMLResponse
		if respMethod == model.Messages.LoadedDataStore || respMethod == model.Messages.StoredDataStore {
			resp = model.BuildResponse(http.StatusOK, respMethod, dto.ID)
		} else {
			resp = model.BuildResponse(http.StatusInternalServerError, respMethod, dto.ID)
		}

		writeResponseMessage(w, dto, *resp)
		return
	}
}

func PollToClientCallback(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)

	msToken := getQueryParameter(r, "msToken")
	tokinfo := getQueryParameter(r, "tokenInfo")

	dto, _, err := initialEPSetup(w, msToken, "", false)
	if err != nil {
		return
	}

	log.Println("Token to be sent in the ClientCallback: " + tokinfo)

	url := services.ClientCallbackAddrRedirect(tokinfo, dto.ClientCallbackAddr)
	http.Redirect(w, r, url, http.StatusFound)

	return
}

func GenerateQRCode(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)

	token := getQueryParameter(r, "msToken")

	dto, contents, err := initialEPSetup(w, token, "", false)
	if err != nil {
		return
	}

	var params sm.RequestParameters
	json.Unmarshal([]byte(contents), &params)
	log.Println(params)
	var smResp sm.SessionMngrResponse
	json.Unmarshal([]byte(params.Data), &smResp)
	var variables model.QRVariables
	json.Unmarshal([]byte(smResp.AdditionalData), &variables)
	dto.Method = variables.Method

	log.Println(variables)
	mobileQRCode(dto, variables, w)
}

// Recieves Token and SessionId from Cloud Redirect
// Creates Token with the Code and Stores it into Session
// Opens Insert Password
func RetrieveCode(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)

	id := getQueryParameter(r, "state")
	code := getQueryParameter(r, "code")

	smResp := getSessionData(id, w)
	dto, err := dto.PersistenceFactory(id, smResp)
	log.Println("Current Persistence Object: ", dto)
	fmt.Println(err)
	if err != nil {
		writeResponseMessage(w, dto, *err)
		return
	}

	dto, err = services.UpdateTokenFromCode(dto, code)
	redirectToOperation(dto, w, r)
}

package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/services"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/gorilla/mux"
)

func AuxiliaryEndpoints(w http.ResponseWriter, r *http.Request) {

	method := mux.Vars(r)["method"]

	if method == "save" {

		log.Println(r.URL.Path)
		token := getQueryParameter(r, "msToken")
		smResp, err := sm.ValidateToken(token)
		if err != nil {
			id := smResp.SessionData.SessionID
			dto, _ := dto.PersistenceFactory(id, sm.SessionMngrResponse{})
			writeResponseMessage(w, dto, *err)
			return
		}

		//Downloads File for the localFile System
		log.Println("save")

		smResp, _ = sm.GetSessionData(smResp.SessionData.SessionID)
		contents := getQueryParameter(r, "contents")
		w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(smResp.SessionData.SessionVariables["DSFilename"]+model.EnvVariables.DataStore_File_Ext))
		w.Header().Set("Content-Type", "application/octet-stream")
		json.NewEncoder(w).Encode(contents)

		return

	} else if method == "checkQrCodePoll" {

		id := getQueryParameter(r, "sessionId")
		smResp := getSessionData(id, w)

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

		id := getQueryParameter(r, "sessionId")
		op := getQueryParameter(r, "operation")

		respMethod, dto, err := services.QRCodePoll(id, op)
		if err != nil {
			writeResponseMessage(w, dto, *err)
		}

		var resp *model.HTMLResponse
		if respMethod == model.Messages.LoadedDataStore || respMethod == model.Messages.StoredDataStore {
			resp = model.BuildResponse(http.StatusOK, respMethod)
		} else {
			resp = model.BuildResponse(http.StatusInternalServerError, respMethod)
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
	services.ClientCallbackAddrPost(tokinfo, dto.ClientCallbackAddr)

	//TODO: Remove this section - SAML SP
	if strings.Contains(dto.ClientCallbackAddr, "/per/retrieve") {
		SimulateDashboard(w, r)
	}
	return
}

func GenerateQRCode(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)

	token := getQueryParameter(r, "msToken")

	dto, contents, err := initialEPSetup(w, token, "", false)
	if err != nil {
		return
	}

	var smResp sm.SessionMngrResponse
	json.Unmarshal([]byte(contents), &smResp)
	var variables model.QRVariables
	json.Unmarshal([]byte(smResp.AdditionalData), &variables)
	dto.Method = variables.Method

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
	if err != nil {
		writeResponseMessage(w, dto, *err)
		return
	}

	dto, err = services.UpdateTokenFromCode(dto, code)
	redirectToOperation(dto, w, r)
}

package controller

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strconv"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/services"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/EC-SEAL/perseal/utils"
	"github.com/gorilla/mux"
	"golang.org/x/net/webdav"
)

// Main Entry Point For Front-Channel Operations
func FrontChannelOperations(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)

	method := mux.Vars(r)["method"]
	cipherPassword := getQueryParameter(r, "cipherPassword")
	token := getQueryParameter(r, "msToken")

	smResp, err := sm.ValidateToken(token)
	id := smResp.SessionData.SessionID
	if err != nil {
		if cipherPassword != "" {
			w.WriteHeader(err.Code)
			w.Write([]byte(err.Message))
			return
		} else {
			dto, _ := dto.PersistenceFactory(id, sm.SessionMngrResponse{})
			writeResponseMessage(w, dto, *err)
			return
		}
	}

	smResp = getSessionData(id, w)
	// EXCEPTION: Mobile Storage can be enable if cipherPassword is sent immediatly in the GET request
	if cipherPassword != "" {
		backChannelStoring(w, id, cipherPassword, method, smResp)
		return
	}

	obj, err := dto.PersistenceFactory(id, smResp, method)
	if err != nil {
		writeResponseMessage(w, obj, *err)
		return
	}

	log.Println("Current Persistence Object: ", obj)
	url := redirectToOperation(obj, w)
	if url != "" {
		http.Redirect(w, r, url, http.StatusFound)
	}
}

//Handles DataStore operation (store or load) after password insertion
func DataStoreHandling(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)

	method := mux.Vars(r)["method"]
	dto, err := recieveSessionIdAndPassword(w, r, method)
	if err != nil {
		writeResponseMessage(w, dto, *err)
		return
	}

	log.Println("Current Persistence Object: ", dto)
	var response *model.HTMLResponse
	if dto.Method == model.EnvVariables.Store_Method {
		response, err = services.PersistenceStore(dto)
		if err != nil {
			writeResponseMessage(w, dto, *err)
			return
		}
	} else if dto.Method == model.EnvVariables.Load_Method {

		if dto.PDS == model.EnvVariables.Browser_PDS {
			dto.LocalFileBytes, err = fetchLocalDataStore(r)
			if err != nil {
				writeResponseMessage(w, dto, *err)
				return
			}
		}

		response, err = services.PersistenceLoad(dto)
	} else if dto.Method == model.EnvVariables.Store_Load_Method {
		response, err = services.PersistenceStoreAndLoad(dto)
	}

	if err != nil {
		writeResponseMessage(w, dto, *err)
		return
	} else if response != nil {
		writeResponseMessage(w, dto, *response)
		return
	}
}

//Back-Channel request to Decrypt and Load User's Data
func BackChannelLoading(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)

	method := mux.Vars(r)["method"]
	msToken := r.FormValue("msToken")
	smResp, err := sm.ValidateToken(msToken)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(model.Messages.NoMSTokenErrorMsg))
		return
	}

	id := smResp.SessionData.SessionID
	smResp = getSessionData(id, w)

	cipherPassword := getQueryParameter(r, "cipherPassword")

	if model.Test {
		cipherPassword = utils.HashSUM256(cipherPassword)
		log.Println(cipherPassword)
	}

	dto, err := dto.PersistenceFactory(id, smResp, method)
	log.Println("Current Persistence Object: ", dto)
	dto.Password = cipherPassword
	if dto.Password == "" {
		err := model.BuildResponse(http.StatusBadRequest, model.Messages.NoPassword)
		dto.Response = *err
		writeBackChannelResponse(dto, w)
		return
	}
	if err != nil {
		dto.Response = *err
		writeBackChannelResponse(dto, w)
		return
	}

	dataSstr := r.PostFormValue("dataStore")
	if dataSstr == "" {
		err := model.BuildResponse(http.StatusBadRequest, model.Messages.FailedFoundDataStore)
		dto.Response = *err
		writeBackChannelResponse(dto, w)
		return
	}

	response, err := services.BackChannelDecryption(dto, dataSstr)
	if err != nil {
		log.Println(err)
		dto.Response = *err
		writeBackChannelResponse(dto, w)
		return
	} else {
		/*
			if response.Code == http.StatusOK {
				rmURL := smResp.SessionData.SessionVariables["RMURL"]
				log.Println("RMURL: ", rmURL)
				if rmURL != "" {
					http.Redirect(w, r, rmURL, http.StatusFound)
				}
			}*/
		dto.Response = *response
		sm.UpdateSessionData(dto.ID, "finished", model.EnvVariables.SessionVariables.FinishedPersealBackChannel)
		writeBackChannelResponse(dto, w)

	}
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
		writeBackChannelResponse(obj, w)
	} else {
		sm.UpdateSessionData(obj.ID, "finished", model.EnvVariables.SessionVariables.FinishedPersealBackChannel)
		obj.Response = *response
		writeBackChannelResponse(obj, w)
	}
	return
}

func AuxiliaryEndpoints(w http.ResponseWriter, r *http.Request) {

	method := mux.Vars(r)["method"]

	if method != "checkQrCodePoll" && method != "qrCodePoll" {
		log.Println(r.URL.Path)
		token := getQueryParameter(r, "msToken")
		smResp, err := sm.ValidateToken(token)
		if err != nil {
			id := smResp.SessionData.SessionID
			dto, _ := dto.PersistenceFactory(id, sm.SessionMngrResponse{})
			writeResponseMessage(w, dto, *err)
			return
		}
	}

	if method == "save" {
		//Downloads File for the localFile System
		log.Println("save")

		contents := getQueryParameter(r, "contents")
		w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(model.EnvVariables.DataStore_File_Name))
		w.Header().Set("Content-Type", "application/octet-stream")
		json.NewEncoder(w).Encode(contents)

		return

	} else if method == "checkQrCodePoll" {

		id := getQueryParameter(r, "sessionId")
		smResp := getSessionData(id, w)
		finishedPersealBackChannel := smResp.SessionData.SessionVariables[model.EnvVariables.SessionVariables.FinishedPersealBackChannel]
		if finishedPersealBackChannel == "finished" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(finishedPersealBackChannel))
		} else if finishedPersealBackChannel == "not finished" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Operation Not Yet Finished"))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Session Variable Not Set"))
		}
		return
	} else if method == "qrCodePoll" {
		id := getQueryParameter(r, "sessionId")
		op := getQueryParameter(r, "operation")

		respMethod, dto, err := services.QRCodePoll(id, op)
		if err != nil {
			writeResponseMessage(w, dto, *err)
		}

		resp := model.BuildResponse(http.StatusOK, respMethod)
		writeResponseMessage(w, dto, *resp)
		return
	}
}

func PollToClientCallback(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)

	token := getQueryParameter(r, "msToken")
	tokinfo := getQueryParameter(r, "tokenInfo")

	smResp, err := sm.ValidateToken(token)
	id := smResp.SessionData.SessionID
	if err != nil {
		dto, _ := dto.PersistenceFactory(id, sm.SessionMngrResponse{})
		writeResponseMessage(w, dto, *err)
		return
	}

	smResp, err = sm.GetSessionData(id)
	dto, err := dto.PersistenceFactory(id, smResp)
	log.Println("Current Persistence Object: ", dto)
	if err != nil {
		dto.Response = *err
		writeBackChannelResponse(dto, w)
	}

	log.Println(tokinfo)
	log.Println(dto.ClientCallbackAddr)
	services.ClientCallbackAddrPost(tokinfo, dto.ClientCallbackAddr)
	return
}

func GenerateQRCode(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)

	token := getQueryParameter(r, "msToken")
	smResp, err := sm.ValidateToken(token)
	tokenContents := smResp.AdditionalData
	id := smResp.SessionData.SessionID

	json.Unmarshal([]byte(tokenContents), &smResp)
	var variables QRVariables
	json.Unmarshal([]byte(smResp.AdditionalData), &variables)

	smResp = getSessionData(id, w)

	dto, err := dto.PersistenceFactory(id, smResp, variables.Method)
	log.Println("Current Persistence Object: ", dto)
	if err != nil {
		writeResponseMessage(w, dto, *err)
		return
	}
	/*
		tok1, tok2 := services.BuildDataOfMSToken(variables.SessionId, "OK")
		log.Println(tok1)
		log.Println("\n\n", tok2)
	*/
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
	redirectToOperation(dto, w)
}

func TestWebDav(w http.ResponseWriter, r *http.Request) {

	dirFlag := flag.String("d", "./", "Directory to serve from. Default is CWD")

	flag.Parse()

	dir := *dirFlag

	srv := &webdav.Handler{
		FileSystem: webdav.Dir(dir),
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			if err != nil {
				log.Printf("WEBDAV [%s]: %s, ERROR: %s\n", r.Method, r.URL, err)
			} else {
				log.Printf("WEBDAV [%s]: %s \n", r.Method, r.URL)
			}
		},
	}

	r.Method = "GET"
	r.URL.Path = ""
	srv.ServeHTTP(w, r)
}

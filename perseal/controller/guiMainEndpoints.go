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
	msToken := getQueryParameter(r, "msToken")

	obj, _, err := initialEPSetup(w, msToken, method, false, cipherPassword)
	if err != nil {
		return
	}
	url := redirectToOperation(obj, w, r)
	if url != "" {
		http.Redirect(w, r, url, http.StatusFound)
	}
}

//Handles DataStore operation (store or load) after password insertion
func DataStoreHandling(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)

	method := mux.Vars(r)["method"]
	msToken := r.FormValue("msToken")

	dto, _, err := initialEPSetup(w, msToken, method, false)
	if err != nil {
		return
	}

	password := r.FormValue("password")
	if password == "" {
		err = model.BuildResponse(http.StatusBadRequest, model.Messages.NoPassword)
		err.FailedInput = "Password"
		writeResponseMessage(w, dto, *err)
		return
	}
	sha := utils.HashSUM256(password)
	dto.Password = sha

	dto.DataStoreFileName = r.FormValue("dataStoreName")
	sm.UpdateSessionData(dto.ID, dto.DataStoreFileName, "DSFilename")

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
		dto.Files.FileList = err.Files.FileList
		dto.Files.SizeList = err.Files.SizeList
		dto.Files.TimeList = err.Files.TimeList
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

			// TODO: Does it need to show an iframe to retry the pwd?
			/*
				dto.MenuOption = err.FailedInput
				dto.Response.DataStore = dataSstr
				openInternalHTML(dto, w, insertPasswordHTML)
			*/
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

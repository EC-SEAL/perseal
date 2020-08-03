package controller

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/services"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/EC-SEAL/perseal/utils"
	"github.com/gorilla/mux"
	"golang.org/x/net/webdav"
)

// Main Entry Point For Cloud. Verifies Token, Retrieves Session and checks if has Cloud Token.
//If not, redirects to Cloud Login Page
//If it does, presents page to Insert Password
func FrontChannelOperations(w http.ResponseWriter, r *http.Request) {
	log.Println("FrontChannelOperations")
	method := mux.Vars(r)["method"]
	token := getQueryParameter(r, "msToken")

	id := validateToken(token, w)
	smResp := getSessionData(id, w)
	log.Println("Current Session Id: " + id)
	// EXCEPTION: Mobile Storage can be enable if cipherPassword is sent immediatly in the GET request
	cipherPassword := getQueryParameter(r, "cipherPassword")
	if cipherPassword != "" {
		BackChannelStoring(w, id, cipherPassword, method, smResp)
		return
	}

	obj, err := dto.PersistenceBuilder(id, smResp, method)
	if err != nil {
		writeResponseMessage(w, obj, *err)
		return
	}

	log.Println("Current Persistence Object: ", obj)
	url := redirectToOperation(obj, w)
	if url != "" {
		http.Redirect(w, r, url, 302)
	}
}

func DataStoreHandling(w http.ResponseWriter, r *http.Request) {
	log.Println("DataStore Handling")

	method := mux.Vars(r)["method"]
	dto, err := recieveSessionIdAndPassword(w, r, method)
	if err != nil {
		writeResponseMessage(w, dto, *err)
		return
	}
	log.Println(sm.NewSearch(dto.ID))

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

func BackChannelLoading(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RequestURI)
	log.Println("persistenceLoadBackChannel")

	method := mux.Vars(r)["method"]
	msToken := r.FormValue("msToken")
	id := validateToken(msToken, w)
	log.Println(msToken)
	smResp := getSessionData(id, w)

	cipherPassword := getQueryParameter(r, "cipherPassword")

	if model.Test {
		cipherPassword = utils.HashSUM256(cipherPassword)
		log.Println(cipherPassword)
	}

	dto, err := dto.PersistenceBuilder(id, smResp, method)
	dto.Password = cipherPassword
	if err != nil {
		w.WriteHeader(err.Code)
		w.Write([]byte(err.Message))
		return
	}

	dataSstr := r.PostFormValue("dataStore")
	if dataSstr == "" {
		err := model.BuildResponse(http.StatusBadRequest, model.Messages.FailedFoundDataStore)
		w.WriteHeader(err.Code)
		w.Write([]byte(err.Message))
		return
	}

	response, err := services.BackChannelDecryption(dto, dataSstr)
	if err != nil {
		w.WriteHeader(err.Code)
		w.Write([]byte(err.Message))
	} else {
		w.WriteHeader(response.Code)
		w.Write([]byte(response.DataStore))
	}
	return
}

func BackChannelStoring(w http.ResponseWriter, id, cipherPassword, method string, smResp sm.SessionMngrResponse) {
	obj, err := dto.PersistenceBuilder(id, smResp, method)
	obj.Password = cipherPassword
	if err != nil {
		writeResponseMessage(w, obj, *err)
		return
	}

	response, err := services.BackChannelStorage(obj)
	token, err := sm.GenerateToken(model.EnvVariables.Perseal_Sender_Receiver, model.EnvVariables.Perseal_Sender_Receiver, obj.ID)
	response.MSToken = token.AdditionalData

	if err != nil {
		w.WriteHeader(err.Code)
		w.Write([]byte(err.Message))
	} else {
		w.WriteHeader(response.Code)
		w.Write([]byte(response.DataStore))
	}
	return
}
func AuxiliaryEndpoints(w http.ResponseWriter, r *http.Request) {

	log.Println("aux")
	method := mux.Vars(r)["method"]

	msToken := getQueryParameter(r, "msToken")
	id := validateToken(msToken, w)

	if method == "save" {
		//Downloads File for the localFile System
		log.Println("save")
		log.Println(msToken)
		contents := getQueryParameter(r, "contents")
		w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(model.EnvVariables.DataStore_File_Name))
		w.Header().Set("Content-Type", "application/octet-stream")
		json.NewEncoder(w).Encode(contents)

		return
	} else if method == "generateQRcode" {
		log.Println("generateQRcode")

		smResp := getSessionData(id, w)
		operation := getQueryParameter(r, "operation")
		log.Println(operation)

		dto, err := dto.PersistenceBuilder(id, smResp, operation)
		if err != nil {
			writeResponseMessage(w, dto, *err)
			return
		}

		mobileQRCode(dto, w)
	} else if method == "redirect" {

		token := getQueryParameter(r, "token")
		clientCallbackAddr := getQueryParameter(r, "clientCallbackAddr")

		hc := http.Client{}
		form := url.Values{}
		form.Add("msToken", token)

		log.Println(clientCallbackAddr)
		req, _ := http.NewRequest(http.MethodPost, clientCallbackAddr, strings.NewReader(form.Encode()))
		log.Println(req)
		log.Println(hc.Do(req))
		return
	}
}

// Recieves Token and SessionId from Cloud Redirect
// Creates Token with the Code and Stores it into Session
// Opens Insert Password
func RetrieveCode(w http.ResponseWriter, r *http.Request) {
	log.Println("recieveCode")
	id := getQueryParameter(r, "state")
	code := getQueryParameter(r, "code")

	smResp := getSessionData(id, w)
	dto, err := dto.PersistenceBuilder(id, smResp)
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

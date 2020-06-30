package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/services"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/gorilla/mux"
)

// Main Entry Point For Cloud. Verifies Token, Retrieves Session and checks if has Cloud Token.
//If not, redirects to Cloud Login Page
//If it does, presents page to Insert Password
func FrontChannelOperations(w http.ResponseWriter, r *http.Request) {
	log.Println("FrontChannelOperations")
	method := mux.Vars(r)["method"]
	token := getQueryParameter(r, "msToken")

	id := validateToken(token, w)
	sm.UpdateSessionData(id, method, "CurrentMethod")
	smResp := getSessionData(id, w)

	obj, err := dto.PersistenceBuilder(id, smResp, method)
	if err != nil {
		writeResponseMessage(w, obj, *err)
		return
	}

	log.Println(obj)
	url := redirectToOperation(obj, w)
	if url != "" {
		http.Redirect(w, r, url, 302)
	}
}

func DataStoreHandling(w http.ResponseWriter, r *http.Request) {
	log.Println("password inserted")

	dto, err := recieveSessionIdAndPassword(r)
	log.Println(dto)
	if err != nil {
		writeResponseMessage(w, dto, *err)
		return
	}

	var response *model.HTMLResponse
	if dto.Method == "store" {
		response, err = services.PersistenceStore(dto)
	} else if dto.Method == "load" {
		response, err = services.PersistenceLoad(dto, r)
	} else if dto.Method == "storeload" {
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

func Save(w http.ResponseWriter, r *http.Request) {

	method := mux.Vars(r)["method"]

	if method == "storeAndLoad" {

		// Activated When Cloud Drive does not have files, so it can store and load the dataStore
		log.Println("storeAndLoad")
		id := r.FormValue("sessionId")

		sessionData := getSessionData(id, w)

		dto, err := dto.PersistenceBuilder(id, sessionData)
		if err != nil {
			writeResponseMessage(w, dto, *err)
			return
		}
		dto.StoreAndLoad = true
		insertPassword(dto, w)
	} else if method == "save" {

		//Downloads File for the localFile System
		log.Println("save")
		contents := getQueryParameter(r, "contents")
		w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote("datastore.seal"))
		w.Header().Set("Content-Type", "application/octet-stream")
		json.NewEncoder(w).Encode(contents)

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

	sessionData := getSessionData(id, w)

	dto, err := dto.PersistenceBuilder(id, sessionData)
	if err != nil {
		writeResponseMessage(w, dto, *err)
		return
	}

	dto, err = services.UpdateTokenFromCode(dto, code)
	if err != nil {
		writeResponseMessage(w, dto, *err)
		return
	}
	log.Println(dto.Method)
	redirectToOperation(dto, w)
}

func BackChannelDecryption(w http.ResponseWriter, r *http.Request) {
	log.Println("persistenceLoadWithToken")

	id := mux.Vars(r)["sessionToken"]
	sessionData := getSessionData(id, w)

	cipherPassword := getQueryParameter(r, "cipherPassword")

	dto, err := dto.PersistenceWithPasswordBuilder(id, sessionData, cipherPassword)
	if err != nil {
		w.WriteHeader(err.Code)
		w.Write([]byte(err.Message))
		return
	}

	dataSstr := r.PostFormValue("dataStore")
	if dataSstr == "" {
		err := &model.HTMLResponse{
			Code:    400,
			Message: "Couldn't find DataStore",
		}
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
		w.Write([]byte(response.Message))
	}
	return
}

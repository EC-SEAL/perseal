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

// Main Entry Point For Cloud. Verifies Token, Retrieves Session and checks if has Cloud Token.
//If not, redirects to Cloud Login Page
//If it does, presents page to Insert Password
func FrontChannelOperations(w http.ResponseWriter, r *http.Request) {
	log.Println("FrontChannelOperations")
	method := mux.Vars(r)["method"]
	token := getQueryParameter(r, "msToken")

	id := validateToken(token, w)
	sm.UpdateSessionData(id, method, model.EnvVariables.SessionVariables.CurrentMethod)
	smResp := getSessionData(id, w)

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
	dto, err := recieveSessionIdAndPassword(w, r)
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

func AuxiliaryEndpoints(w http.ResponseWriter, r *http.Request) {

	log.Println("aux")
	method := mux.Vars(r)["method"]

	msToken := getQueryParameter(r, "msToken")
	id := validateToken(msToken, w)

	if method == "storeAndLoad" {
		// Activated When Cloud Drive does not have files, so it can store and load the dataStore
		log.Println("storeAndLoad")

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
	redirectToOperation(dto, w)
}

func BackChannelOperations(w http.ResponseWriter, r *http.Request) {
	log.Println("persistenceLoadWithToken")

	method := mux.Vars(r)["method"]
	msToken := r.FormValue("msToken")
	id := validateToken(msToken, w)
	sessionData, err := sm.GetSessionData(id)

	cipherPassword := getQueryParameter(r, "cipherPassword")

	if model.Test {
		cipherPassword = utils.HashSUM256(cipherPassword)
		log.Println(cipherPassword)
	}

	dto, err := dto.PersistenceWithPasswordBuilder(id, sessionData, cipherPassword)
	if err != nil {
		w.WriteHeader(err.Code)
		w.Write([]byte(err.Message))
		return
	}

	if method == model.EnvVariables.Store_Method {
		response, err := services.BackChannelStorage(dto)
		if err != nil {
			w.WriteHeader(err.Code)
			w.Write([]byte(err.Message))
		} else {
			w.WriteHeader(response.Code)
			w.Write([]byte(response.DataStore))
		}
		return
	} else if method == model.EnvVariables.Load_Method {
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
			w.Write([]byte(response.DataStore))
		}
		return
	} else {
		w.WriteHeader(404)
		w.Write([]byte("Invalid Method"))
	}
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

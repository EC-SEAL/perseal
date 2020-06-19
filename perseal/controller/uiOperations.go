package controller

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/EC-SEAL/perseal/sm"
	"github.com/gorilla/mux"
)

// Main Entry Point For Cloud. Verifies Token, Retrieves Session and checks if has Cloud Token.
//If not, redirects to Cloud Login Page
//If it does, presents page to Insert Password
func InitialCloudConfig(w http.ResponseWriter, r *http.Request) {
	log.Println("initial cloud config")
	method := mux.Vars(r)["method"]
	var token string
	if keys, ok := r.URL.Query()["msToken"]; ok {
		token = keys[0]
	}

	id, err := sm.ValidateToken(token)
	if err != nil {
		dto, err := persistenceBuilder(id, sm.SessionMngrResponse{})
		writeResponseMessage(w, dto, *err)
	}

	sm.UpdateSessionData(id, method, "CurrentMethod")
	log.Println(method)
	initialConfig(id, method, w, r)
}

// Main Entry Point. Verifies Token, Retrieves Session and checks if has Cloud Token.
//If not, redirects to Cloud Login Page
//If it does, presents page to Insert Password
func InitialLocalConfig(w http.ResponseWriter, r *http.Request) {
	log.Println("initial local config")
	method := mux.Vars(r)["method"]
	id := mux.Vars(r)["sessionToken"]

	sm.UpdateSessionData(id, method, "CurrentMethod")
	initialConfig(id, method, w, r)
}

// Activated When Cloud Drive does not have files, so it can store and load the dataStore
func InsertPasswordStoreAndLoad(w http.ResponseWriter, r *http.Request) {
	log.Println(r.FormValue("sessionId"))
	id := r.FormValue("sessionId")
	sessionData, err := sm.GetSessionData(id, "")
	if err != nil {
		dto, err := persistenceBuilder(id, sm.SessionMngrResponse{})
		writeResponseMessage(w, dto, *err)
	}

	dto, err := persistenceBuilder(id, sessionData)
	if err != nil {
		writeResponseMessage(w, dto, *err)
	}

	dto.Method = ""
	dto.StoreAndLoad = true
	t, _ := template.ParseFiles("ui/insertPassword.html")
	t.Execute(w, dto)
}

func Save(w http.ResponseWriter, r *http.Request) {
	log.Println("save")
	var contents string
	if keys, ok := r.URL.Query()["contents"]; ok {
		contents = keys[0]
	}
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote("datastore.seal"))
	w.Header().Set("Content-Type", "application/octet-stream")
	json.NewEncoder(w).Encode(contents)
	return
}

//OTHERS

/*
func GenerateToken(w http.ResponseWriter, r *http.Request) {
	log.Println("generateToken")

	var id, method string
	if keys, ok := r.URL.Query()["sessionId"]; ok {
		id = keys[0]
	}
	if id == "" {
		err := &model.HTMLResponse{
			Code:    400,
			Message: "Couldn't find Session Token",
		}
		w = writeResponseMessage(w, err, err.Code)
		return
	}

	if keys, ok := r.URL.Query()["method"]; ok {
		method = keys[0]
	}
	if id == "" {
		err := &model.HTMLResponse{
			Code:    400,
			Message: "Couldn't find Session Token",
		}
		w = writeResponseMessage(w, err, err.Code)
		return
	}

	smResp, err := utils.GenerateTokenAPI(method, id)

	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w = writeResponseMessage(w, err, err.Code)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w = writeResponseMessage(w, smResp.Payload, 200)
	return
}

func StartSession(w http.ResponseWriter, r *http.Request) {
	log.Println("startSession")
	smResp, err := utils.StartSession()
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w = writeResponseMessage(w, err, err.Code)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w = writeResponseMessage(w, smResp.Payload, 200)
	return
}

func UpdateSessionData(w http.ResponseWriter, r *http.Request) {
	var id string
	if keys, ok := r.URL.Query()["sessionId"]; ok {
		id = keys[0]
	}
	if id == "" {
		err := &model.HTMLResponse{
			Code:    400,
			Message: "Couldn't find Session Token",
		}
		w = writeResponseMessage(w, err, err.Code)
		return
	}
	sm.UpdateSessionData(id, "Moblie", "PDS")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w = writeResponseMessage(w, "", 200)
	return
}
*/

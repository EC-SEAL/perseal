package controller

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/services"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/EC-SEAL/perseal/utils"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
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
	log.Println(token)

	id, err := sm.ValidateToken(token)
	if err != nil {
		dto, _ := dto.PersistenceBuilder(id, sm.SessionMngrResponse{})
		log.Println(dto)
		writeResponseMessage(w, dto, *err)
		return
	} else {
		sm.UpdateSessionData(id, method, "CurrentMethod")
		log.Println(method)
		initialConfig(id, method, w, r)
	}
}

// Activated When Cloud Drive does not have files, so it can store and load the dataStore
func InsertPasswordStoreAndLoad(w http.ResponseWriter, r *http.Request) {
	log.Println(r.FormValue("sessionId"))
	id := r.FormValue("sessionId")
	sessionData, err := sm.GetSessionData(id, "")
	if err != nil {
		dto, err := dto.PersistenceBuilder(id, sm.SessionMngrResponse{})
		writeResponseMessage(w, dto, *err)
		return
	}

	dto, err := dto.PersistenceBuilder(id, sessionData)
	if err != nil {
		writeResponseMessage(w, dto, *err)
		return
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

// Recieves Token and SessionId from Cloud Redirect
// Creates Token with the Code and Stores it into Session
// Opens Insert Password
func RetrieveCode(w http.ResponseWriter, r *http.Request) {
	log.Println("recieveCode")

	var id, code string
	if keys, ok := r.URL.Query()["state"]; ok {
		id = keys[0]
	}
	if keys, ok := r.URL.Query()["code"]; ok {
		code = keys[0]
	}
	sessionData, err := sm.GetSessionData(id, "")
	if err != nil {
		dto, err := dto.PersistenceBuilder(id, sm.SessionMngrResponse{})
		writeResponseMessage(w, dto, *err)
		return
	}

	dto, err := dto.PersistenceBuilder(id, sessionData)
	if err != nil {
		writeResponseMessage(w, dto, *err)
		return
	}

	var token *oauth2.Token
	if dto.PDS == "googleDrive" {
		token, err = services.UpdateNewGoogleDriveTokenFromCode(dto.ID, code)
		b, _ := json.Marshal(token)
		dto.GoogleAccessCreds = string(b)
	} else if dto.PDS == "oneDrive" {
		token, err = services.UpdateNewOneDriveTokenFromCode(dto.ID, code)
		dto.OneDriveToken = *token

	}
	log.Println(dto.Method)
	redirectToOperation(dto, w)
}

// Main Entry Point For Cloud. Verifies Token, Retrieves Session and checks if has Cloud Token.
//If not, redirects to Cloud Login Page
//If it does, presents page to Insert Password
func Test(w http.ResponseWriter, r *http.Request) {
	log.Println("test")
	token := r.FormValue("msToken")
	method := mux.Vars(r)["method"]

	log.Println(token)
	id, err := sm.ValidateToken(token)
	if err != nil {
		dto, _ := dto.PersistenceBuilder(id, sm.SessionMngrResponse{})
		log.Println(dto)
		writeResponseMessage(w, dto, *err)
		return
	}

	sm.UpdateSessionData(id, method, "CurrentMethod")
	smResp, err := sm.GetSessionData(id, "")

	if err != nil {
		obj, err := dto.PersistenceBuilder(id, sm.SessionMngrResponse{}, "")
		writeResponseMessage(w, obj, *err)
		return
	}

	obj, err := dto.PersistenceBuilder(id, smResp)
	log.Println(obj.Method)
	if err != nil {
		writeResponseMessage(w, obj, *err)
		return
	}

	log.Println(obj)
	url, err := services.GetRedirectURL(obj)

	w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
	w.Write([]byte(url))
	return
}

func SimulateDashboard(w http.ResponseWriter, r *http.Request) {

	type testStruct struct {
		ID      string
		MSToken string
	}

	testing := testStruct{
		ID:      model.TestUser,
		MSToken: model.MSToken,
	}
	t, _ := template.ParseFiles("ui/simulateDashboard.html")
	t.Execute(w, testing)
}

func StartSession(w http.ResponseWriter, r *http.Request) {
	resp, _ := utils.StartSession()
	model.TestUser = resp.Payload
	var url string
	if model.Local {
		url = "http://localhost:8082"
	} else {
		url = "http://perseal.seal.eu:8082"
	}
	http.Redirect(w, r, url+"/per/simulateDashboard", 302)
}

func Token(w http.ResponseWriter, r *http.Request) {
	var method string
	if keys, ok := r.URL.Query()["method"]; ok {
		method = keys[0]
	}
	resp, _ := utils.GenerateTokenAPI(method, model.TestUser)
	model.MSToken = resp.Payload
	var url string
	if model.Local {
		url = "http://localhost:8082"
	} else {
		url = "http://perseal.seal.eu:8082"
	}
	http.Redirect(w, r, url+"/per/simulateDashboard", 302)
}

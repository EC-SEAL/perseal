package controller

import (
	"html/template"
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/services"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/EC-SEAL/perseal/utils"
	"github.com/gorilla/mux"
)

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
	smResp, err := sm.GetSessionData(id)

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
	url := services.GetRedirectURL(obj)

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
	if model.Test {
		url = "http://localhost:8082"
	} else {
		url = "http://perseal.seal.eu:8082"
	}
	http.Redirect(w, r, url+"/per/test/simulateDashboard", 302)
}

func Token(w http.ResponseWriter, r *http.Request) {
	var method string
	if keys, ok := r.URL.Query()["method"]; ok {
		method = keys[0]
	}
	resp, _ := utils.GenerateTokenAPI(method, model.TestUser)
	model.MSToken = resp.Payload
	var url string
	if model.Test {
		url = "http://localhost:8082"
	} else {
		url = "http://perseal.seal.eu:8082"
	}
	http.Redirect(w, r, url+"/per/test/simulateDashboard", 302)
}

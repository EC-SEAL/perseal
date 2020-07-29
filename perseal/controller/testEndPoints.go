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
)

// Main Entry Point For Cloud. Verifies Token, Retrieves Session and checks if has Cloud Token.
//If not, redirects to Cloud Login Page
//If it does, presents page to Insert Password
func Test(w http.ResponseWriter, r *http.Request) {
	log.Println("test")
	token := r.FormValue("msToken")

	log.Println(token)
	id, err := sm.ValidateToken(token)
	smResp := getSessionData(id, w)
	if err != nil {
		dto, _ := dto.PersistenceBuilder(id, smResp)
		log.Println(dto)
		writeResponseMessage(w, dto, *err)
		return
	}

	if err != nil {
		obj, err := dto.PersistenceBuilder(id, smResp)
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
	//	resp, _ := utils.StartSession()

	respo, _ := utils.StartSession("")
	log.Println(respo)
	model.TestUser = respo.Payload

	sm.NewAdd(model.TestUser, "this is a link request", "linkRequest")
	sm.UpdateSessionData(model.TestUser, "Mobile", model.EnvVariables.SessionVariables.UserDevice)

	var url string
	if model.Test {
		url = "http://localhost:8082"
	} else {
		url = "http://perseal.seal.eu:8082"
	}
	http.Redirect(w, r, url+"/per/test/simulateDashboard", 302)
}

func Token(w http.ResponseWriter, r *http.Request) {
	log.Println(sm.NewSearch(model.TestUser))
	var method string
	if keys, ok := r.URL.Query()["method"]; ok {
		method = keys[0]
	}

	var err *model.HTMLResponse
	model.MSToken, err = utils.GenerateTokenAPI(method, model.TestUser)
	if err != nil {
		log.Println(err)
	}
	log.Println(sm.GetSessionData(model.TestUser))

	/*
		resp, _ := sm.NewSearch(model.TestUser)
		log.Println("Search Before Delete: ", resp)
		resp, _ = sm.NewDelete(model.TestUser)
		log.Println("Delete Response: ", resp)
		log.Println("Waiting 3 seconds, just because")
		time.Sleep(3 * time.Second)
		resp, _ = sm.NewSearch(model.TestUser)
		log.Println("Search After Delete: ", resp)

		resp, _ = sm.NewAdd(model.TestUser, "NEW DATA", "dataSet")
		log.Println("Add Response: ", resp)
		resp, _ = sm.NewSearch(model.TestUser)
		log.Println("Search After Adding: ", resp)
	*/

	var url string
	if model.Test {
		url = "http://localhost:8082"
	} else {
		url = "http://perseal.seal.eu:8082"
	}
	http.Redirect(w, r, url+"/per/test/simulateDashboard", 302)
}

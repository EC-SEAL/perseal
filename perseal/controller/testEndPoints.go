package controller

import (
	"html/template"
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/EC-SEAL/perseal/utils"
)

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

	/*
		var s sm.NewUpdateDataRequest = sm.NewUpdateDataRequest{
			SessionId: model.TestUser,
			Type:      "linkRequest",
			Data:      "this is",
			ID:        uuid.New().String(),
		}
		var s2 sm.NewUpdateDataRequest = sm.NewUpdateDataRequest{
			SessionId: model.TestUser,
			Type:      "linkRequestButDuplicate",
			Data:      "this is yet another linkRequest",
			ID:        uuid.New().String(),
		}

		sm.NewAdd(s)
		sm.NewAdd(s2)
	*/
	/*
		var s sm.NewUpdateDataRequest = sm.NewUpdateDataRequest{
			SessionId: model.TestUser,
			Type:      "ahahachanged",
			Data:      "this is !!!!!!!!!!!!!!!!!!!",
			ID:        uuid.New().String(),
		}
		sm.NewAdd(s)
		var s2 sm.NewUpdateDataRequest = sm.NewUpdateDataRequest{
			SessionId: model.TestUser,
			Type:      "new tipo like what?",
			Data:      "this is smt very coiso",
			ID:        uuid.New().String(),
		}

		sm.NewAdd(s2)
		log.Println(sm.NewSearch(model.TestUser))
	*/
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
	log.Println(model.MSToken)
	var url string
	if model.Test {
		url = "http://localhost:8082"
	} else {
		url = "http://perseal.seal.eu:8082"
	}
	http.Redirect(w, r, url+"/per/test/simulateDashboard", 302)
}

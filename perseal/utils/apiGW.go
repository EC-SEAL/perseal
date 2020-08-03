package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/model"
)

type DashboardValidation struct {
	SessionId string `json:"sessionId"`
	Code      string `json:"code"`
	Message   string `json:"message"`
}

func StartSession(sessionId ...string) (tokenResp model.TokenResponse, err *model.HTMLResponse) {

	url := model.EnvVariables.TestURLs.APIGW_Endpoint + "/cl/session/start"

	req, erro := http.NewRequest(http.MethodGet, url, nil)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         500,
			Message:      "Couldn't Generate URL to Start Session",
			ErrorMessage: erro.Error(),
		}
		return
	}

	q := req.URL.Query()
	if len(sessionId) > 0 || sessionId != nil {
		q.Add(model.EnvVariables.SessionVariables.SessionId, sessionId[0])
	}
	req.URL.RawQuery = q.Encode()
	return apiGWRequest(req)
}

func GenerateTokenAPI(method string, id string) (msToken string, err *model.HTMLResponse) {

	url := model.EnvVariables.TestURLs.APIGW_Endpoint + "/cl/persistence/" + method + "/store"
	req, erro := http.NewRequest("GET", url, nil)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         500,
			Message:      "Couldn't Generate URL to Generate Token",
			ErrorMessage: erro.Error(),
		}
		return
	}

	q := req.URL.Query()
	q.Add("sessionID", id)
	req.URL.RawQuery = q.Encode()

	tok, err := apiGWRequest(req)
	if err != nil {
		return
	}
	log.Println(req)
	log.Println(tok)
	msToken = tok.Payload

	return
}

func apiGWRequest(req *http.Request) (tokenResp model.TokenResponse, err *model.HTMLResponse) {

	req.Header.Set("Accept", "application/json")

	var client http.Client
	log.Println(req)
	resp, erro := client.Do(req)
	log.Println(resp)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         404,
			Message:      "Couldn't Execute Request to Generate Token",
			ErrorMessage: erro.Error(),
		}
		return
	}

	body, erro := ioutil.ReadAll(resp.Body)

	if err != nil {
		err = &model.HTMLResponse{
			Code:         500,
			Message:      "Couldn't Read Response from Request to  Generate Token",
			ErrorMessage: erro.Error(),
		}
		return
	}

	var dat interface{}
	json.Unmarshal([]byte(body), &dat)
	jsonM, erro := json.Marshal(dat)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         500,
			Message:      "Couldn't Generate JSON From Response Body of Generate Token",
			ErrorMessage: erro.Error(),
		}
		return
	}

	tokenResp = model.TokenResponse{}
	json.Unmarshal(jsonM, &tokenResp)
	return
}

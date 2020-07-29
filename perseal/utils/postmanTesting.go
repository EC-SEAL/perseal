package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/model"
)

func StartSession(sessionId string) (tokenResp model.TokenResponse, err *model.HTMLResponse) {
	var url string
	if sessionId == "" {
		url = "https://vm.project-seal.eu:9154/cl/session/start"
	} else {
		url = "https://vm.project-seal.eu:9154/cl/session/start?sessionID=" + sessionId
	}

	req, erro := http.NewRequest("GET", url, nil)

	req.Header.Set("Accept", "application/json")
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         500,
			Message:      "Couldn't Generate URL to Generate Token",
			ErrorMessage: erro.Error(),
		}
		return
	}

	var client http.Client
	log.Println(req)
	resp, erro := client.Do(req)
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
	fmt.Println(tokenResp.Payload)
	return
}

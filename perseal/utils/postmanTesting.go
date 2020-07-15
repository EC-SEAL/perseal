package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/EC-SEAL/perseal/model"
)

func StartSession() (tokenResp model.TokenResponse, err *model.HTMLResponse) {
	url := "https://vm.project-seal.eu:9053/cl/session/start"

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

	fmt.Println(req.URL)
	var client http.Client
	resp, erro := client.Do(req)
	fmt.Println("\n", resp)
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
	fmt.Println("\n", dat)
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

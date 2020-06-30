package dto

import (
	"encoding/json"
	"log"

	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
	"golang.org/x/oauth2"
)

type PersistenceDTO struct {
	ID                 string
	PDS                string
	Method             string
	ClientCallbackAddr string
	SMResp             sm.SessionMngrResponse
	Password           string
	StoreAndLoad       bool
	GoogleAccessCreds  oauth2.Token
	OneDriveToken      oauth2.Token
	Response           model.HTMLResponse
	IsLocalLoad        bool
	Image              string
}

// Builds Standard Persistence DTO
func PersistenceBuilder(id string, sessionData sm.SessionMngrResponse, method ...string) (dto PersistenceDTO, err *model.HTMLResponse) {
	client := sessionData.SessionData.SessionVariables["ClientCallback"]
	if client == "" {
		client = "https://vm.project-seal.eu:9053/swagger-ui.html"
	}

	dto = PersistenceDTO{
		ID:                 id,
		PDS:                sessionData.SessionData.SessionVariables["PDS"],
		SMResp:             sessionData,
		ClientCallbackAddr: client,
		IsLocalLoad:        false,
		StoreAndLoad:       false,
	}
	googleTokenBytes, oneDriveTokenBytes, err := getGoogleAndOneDriveTokens(sessionData)
	if err != nil {
		return
	}

	json.Unmarshal(googleTokenBytes, &dto.GoogleAccessCreds)
	json.Unmarshal(oneDriveTokenBytes, &dto.OneDriveToken)

	if len(method) > 0 || method != nil {
		dto.Method = method[0]
	} else {
		dto.Method = sessionData.SessionData.SessionVariables["CurrentMethod"]
	}
	return
}

// Builds Persistence DTO With Password
func PersistenceWithPasswordBuilder(id string, sessionData sm.SessionMngrResponse, password string) (dto PersistenceDTO, err *model.HTMLResponse) {

	client := sessionData.SessionData.SessionVariables["ClientCallback"]
	if client == "" {
		client = "https://vm.project-seal.eu:9053/swagger-ui.html"
	}

	dto = PersistenceDTO{
		ID:                 id,
		PDS:                sessionData.SessionData.SessionVariables["PDS"],
		SMResp:             sessionData,
		ClientCallbackAddr: client,
		Password:           password,
		Method:             sessionData.SessionData.SessionVariables["CurrentMethod"],
		IsLocalLoad:        false,
		StoreAndLoad:       false,
	}

	googleTokenBytes, oneDriveTokenBytes, err := getGoogleAndOneDriveTokens(sessionData)
	if err != nil {
		return
	}

	json.Unmarshal(googleTokenBytes, &dto.GoogleAccessCreds)
	json.Unmarshal(oneDriveTokenBytes, &dto.OneDriveToken)
	return
}

func getGoogleAndOneDriveTokens(sessionData sm.SessionMngrResponse) (googleTokenBytes, oneDriveTokenBytes []byte, err *model.HTMLResponse) {
	var data interface{}
	json.Unmarshal([]byte(sessionData.SessionData.SessionVariables["OneDriveToken"]), &data)
	oneDriveTokenBytes, erro := json.Marshal(data)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         400,
			Message:      "Could Not Marshall One Drive Token",
			ErrorMessage: erro.Error(),
		}
		return
	}
	var data2 interface{}
	json.Unmarshal([]byte(sessionData.SessionData.SessionVariables["GoogleDriveAccessCreds"]), &data2)
	googleTokenBytes, erro = json.Marshal(data2)
	log.Println("\n\n\n\n", googleTokenBytes)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         400,
			Message:      "Could Not Marshall One Drive Token",
			ErrorMessage: erro.Error(),
		}
	}
	return
}

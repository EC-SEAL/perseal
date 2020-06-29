package dto

import (
	"encoding/json"

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
	GoogleAccessCreds  string
	OneDriveToken      oauth2.Token
	DoesNotHaveFiles   bool
	Response           model.HTMLResponse
	IsLocal            bool
	Image              string
}

// Builds Standard Persistence DTO
func PersistenceBuilder(id string, sessionData sm.SessionMngrResponse, method ...string) (PersistenceDTO, *model.HTMLResponse) {
	var data interface{}
	json.Unmarshal([]byte(sessionData.SessionData.SessionVariables["OneDriveToken"]), &data)
	jsonM, erro := json.Marshal(data)
	if erro != nil {
		err := &model.HTMLResponse{
			Code:         400,
			Message:      "Could Not Marshall One Drive Token",
			ErrorMessage: erro.Error(),
		}
		return PersistenceDTO{}, err
	}

	client := sessionData.SessionData.SessionVariables["ClientCallback"]

	if client == "" {
		client = "https://vm.project-seal.eu:9053/swagger-ui.html"
	}

	dto := PersistenceDTO{
		ID:                 id,
		PDS:                sessionData.SessionData.SessionVariables["PDS"],
		SMResp:             sessionData,
		ClientCallbackAddr: client,
		GoogleAccessCreds:  sessionData.SessionData.SessionVariables["GoogleDriveAccessCreds"],
	}
	json.Unmarshal(jsonM, &dto.OneDriveToken)

	if len(method) > 0 || method != nil {
		dto.Method = method[0]
	} else {
		dto.Method = sessionData.SessionData.SessionVariables["CurrentMethod"]
	}
	return dto, nil
}

// Builds Persistence DTO With Password
func PersistenceWithPasswordBuilder(id string, sessionData sm.SessionMngrResponse, password string) (PersistenceDTO, *model.HTMLResponse) {
	var data interface{}
	json.Unmarshal([]byte(sessionData.SessionData.SessionVariables["OneDriveToken"]), &data)
	jsonM, erro := json.Marshal(data)
	if erro != nil {
		err := &model.HTMLResponse{
			Code:         400,
			Message:      "Could Not Marshall One Drive Token",
			ErrorMessage: erro.Error(),
		}
		return PersistenceDTO{}, err
	}

	client := sessionData.SessionData.SessionVariables["ClientCallback"]

	if client == "" {
		client = "https://vm.project-seal.eu:9053/swagger-ui.html"
	}

	dto := PersistenceDTO{
		ID:                 id,
		PDS:                sessionData.SessionData.SessionVariables["PDS"],
		SMResp:             sessionData,
		ClientCallbackAddr: client,
		GoogleAccessCreds:  sessionData.SessionData.SessionVariables["GoogleDriveAccessCreds"],
		Password:           password,
	}
	json.Unmarshal(jsonM, &dto.OneDriveToken)
	return dto, nil
}

package controller

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/services"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/EC-SEAL/perseal/utils"
	"golang.org/x/oauth2"
)

//Opens HTML to display event message and redirect to ClientCallbackAddr
func writeResponseMessage(w http.ResponseWriter, dto dto.PersistenceDTO, response model.HTMLResponse) {
	dto.Response = response
	t, _ := template.ParseFiles("ui/message.html")
	t.Execute(w, response)
}

//Opens HTML of corresponding operation (store or load)
func redirectToOperation(dto dto.PersistenceDTO, w http.ResponseWriter) {
	if dto.PDS != "googleDrive" && dto.PDS != "oneDrive" {
		dto.IsLocal = true
		dto.StoreAndLoad = false
		log.Println(dto.IsLocal)
		t, _ := template.ParseFiles("ui/insertPassword.html")
		t.Execute(w, dto)
	} else {
		dto.IsLocal = false
		if dto.Method == "load" {
			files, _ := services.GetCloudFileNames(dto)
			log.Println(files)

			if files == nil || len(files) == 0 {
				dto.DoesNotHaveFiles = true
				t, _ := template.ParseFiles("ui/noFilesFound.html")
				t.Execute(w, dto)
			} else {
				dto.StoreAndLoad = false
				t, _ := template.ParseFiles("ui/insertPassword.html")
				t.Execute(w, dto)
			}
		} else {
			dto.StoreAndLoad = false
			t, _ := template.ParseFiles("ui/insertPassword.html")
			t.Execute(w, dto)
		}
	}
}

// Retrieves Password and SessionID from recieving request
func recieveSessionIdAndPassword(r *http.Request) (dto dto.PersistenceDTO, err *model.HTMLResponse) {
	password := r.FormValue("password")
	sha := utils.HashSUM256(password)
	id := r.FormValue("sessionId")
	sessionData, err := sm.GetSessionData(id, "")
	if err != nil {
		return
	}

	dto, err = persistenceWithPasswordBuilder(id, sessionData, sha)
	return
}

// Builds Standard Persistence DTO
func persistenceBuilder(id string, sessionData sm.SessionMngrResponse, method ...string) (dto.PersistenceDTO, *model.HTMLResponse) {
	var data interface{}
	json.Unmarshal([]byte(sessionData.SessionData.SessionVariables["OneDriveToken"]), &data)
	jsonM, erro := json.Marshal(data)
	if erro != nil {
		err := &model.HTMLResponse{
			Code:         400,
			Message:      "Could Not Marshall One Drive Token",
			ErrorMessage: erro.Error(),
		}
		return dto.PersistenceDTO{}, err
	}

	client := sessionData.SessionData.SessionVariables["ClientCallback"]

	if client == "" {
		client = "https://vm.project-seal.eu:9053/swagger-ui.html"
	}

	dto := dto.PersistenceDTO{
		ID:                 id,
		PDS:                sessionData.SessionData.SessionVariables["PDS"],
		SMResp:             sessionData,
		ClientCallbackAddr: client,
		Method:             method[0],
		GoogleAccessCreds:  sessionData.SessionData.SessionVariables["GoogleDriveAccessCreds"],
	}
	json.Unmarshal(jsonM, &dto.OneDriveToken)
	return dto, nil
}

// Builds Persistence DTO With Password
func persistenceWithPasswordBuilder(id string, sessionData sm.SessionMngrResponse, password string) (dto.PersistenceDTO, *model.HTMLResponse) {
	var data interface{}
	json.Unmarshal([]byte(sessionData.SessionData.SessionVariables["OneDriveToken"]), &data)
	jsonM, erro := json.Marshal(data)
	if erro != nil {
		err := &model.HTMLResponse{
			Code:         400,
			Message:      "Could Not Marshall One Drive Token",
			ErrorMessage: erro.Error(),
		}
		return dto.PersistenceDTO{}, err
	}

	client := sessionData.SessionData.SessionVariables["ClientCallback"]

	if client == "" {
		client = "https://vm.project-seal.eu:9053/swagger-ui.html"
	}

	dto := dto.PersistenceDTO{
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

func initialConfig(id, method string, w http.ResponseWriter, r *http.Request) {
	smResp, err := sm.GetSessionData(id, "")
	if err != nil {
		dto, err := persistenceBuilder(id, sm.SessionMngrResponse{}, "")
		writeResponseMessage(w, dto, *err)
	}

	dto, err := persistenceBuilder(id, smResp, method)
	log.Println(dto.Method)
	if err != nil {
		writeResponseMessage(w, dto, *err)
	}

	log.Println(dto)
	url, err := services.GetRedirectURL(dto)

	log.Println(url)
	if url != "" {
		http.Redirect(w, r, url, 302)
	} else {
		redirectToOperation(dto, w)
	}
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
		dto, err := persistenceBuilder(id, sm.SessionMngrResponse{})
		writeResponseMessage(w, dto, *err)
	}

	dto, err := persistenceBuilder(id, sessionData)
	if err != nil {
		writeResponseMessage(w, dto, *err)
	}

	if err != nil {
		writeResponseMessage(w, dto, *err)
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
	redirectToOperation(dto, w)
}

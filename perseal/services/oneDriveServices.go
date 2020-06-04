package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/externaldrive"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/EC-SEAL/perseal/utils"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
)

// ONE DRIVE SERVICE METHODS

//Attempts to Store the Session Data On OneDrive
func storeSessionDataOneDrive(dto dto.PersistenceDTO, filename string) (returningdto dto.PersistenceDTO, dataStore *externaldrive.DataStore, err *model.DashboardResponse) {

	returningdto, err = getOneDriveToken(dto)
	if returningdto.StopProcess == true {
		return
	}

	utils.RecieveCheckFirstAccess()
	// Request Password From UI
	returningdto.StopProcess, returningdto.Password = utils.RecievePassword()
	if returningdto.StopProcess == true {
		return
	}

	dataStore, erro := externaldrive.StoreSessionData(returningdto)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Couldn't Generate DataStore to be saved",
			ErrorMessage: erro.Error(),
		}
		return
	}

	var contents []byte
	contents, erro = dataStore.UploadingBlob(returningdto.Token)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Couldn't Generate Blob",
			ErrorMessage: erro.Error(),
		}
		return
	}
	var file *drive.File
	file, erro = dataStore.UploadOneDrive(returningdto.Token, contents, filename, "SEAL")
	fmt.Println(file)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Couldn't Upload DataStore One Drive",
			ErrorMessage: erro.Error(),
		}
	}
	return
}

// Fetches GoogleDrive Code
func loadSessionDataOneDrive(dto dto.PersistenceDTO, filename string) (returningdto dto.PersistenceDTO, file *http.Response, err *model.DashboardResponse) {
	returningdto = dto
	returningdto, err = getOneDriveToken(dto)
	if err != nil {
		return
	}
	if returningdto.StopProcess == true {
		return
	}

	fmt.Println(returningdto.Token)
	jsonM, _ := json.Marshal(returningdto.SMResp)
	smr := &sm.SessionMngrResponse{}
	json.Unmarshal(jsonM, smr)
	str := smr.SessionData.SessionVariables["ClientCallbackAddr"]
	utils.SendLink(str)

	checkFirstAccess := utils.RecieveCheckFirstAccess()

	if checkFirstAccess == true {
		err = &model.DashboardResponse{
			Code:    302,
			Message: "New Store Method",
		}
		return
	}

	file, erro := externaldrive.GetOneDriveItem(returningdto.Token, filename)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         404,
			Message:      "Couldn't Get One Drive Item",
			ErrorMessage: erro.Error(),
		}
		return
	}
	log.Println(file)
	return
}

func getOneDriveToken(dto dto.PersistenceDTO) (returningdto dto.PersistenceDTO, err *model.DashboardResponse) {
	returningdto = dto
	var link string
	var erro error

	creds, err := setOneDriveCreds(returningdto.SMResp)
	link, returningdto.Token, erro = externaldrive.GetOneDriveToken(creds)

	if erro != nil {
		err = &model.DashboardResponse{
			Code:         404,
			Message:      "Couldn't Get One Drive Token",
			ErrorMessage: erro.Error(),
		}
		return
	}

	// If no Token was Found
	if link != "" {
		modelRedirect := model.RedirectStruct{
			Redirect: true,
			URL:      link,
		}

		returningdto.StopProcess = utils.SendRedirect(modelRedirect)
		if returningdto.StopProcess == true {
			return
		}
		code := utils.RecieveCode()
		log.Println(code)

		returningdto.Token, err = updateNewOneDriveTokenFromCode(returningdto.ID, code, creds.OneDriveClientID)

		returningdto.SMResp, err = sm.GetSessionData(returningdto.ID, "")
		if err != nil {
			return
		}
		creds, err = setOneDriveCreds(returningdto.SMResp)
		if err != nil {
			return
		}

		log.Println("TOKEN ", returningdto.Token)
	} else {
		if returningdto.Method != "load&store" {
			modelRedirect := model.RedirectStruct{
				Redirect: false,
				URL:      "",
			}
			utils.SendRedirect(modelRedirect)
		}
	}
	if returningdto.Method != "load&store" {
		if sm.CurrentUser == nil {
			sm.CurrentUser = make(chan sm.SessionMngrResponse)
		}
		jsonM, _ := json.Marshal(returningdto.SMResp)
		smr := &sm.SessionMngrResponse{}
		json.Unmarshal(jsonM, smr)
		sm.CurrentUser <- *smr
		log.Println("tem user")
	}
	return
}

func updateNewOneDriveTokenFromCode(sessionId string, code string, id string) (oauthToken *oauth2.Token, err *model.DashboardResponse) {

	var erro error
	oauthToken, erro = externaldrive.RequestToken(code, id)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         404,
			Message:      "Couldn't Request One Drive Token",
			ErrorMessage: erro.Error(),
		}
		return
	}

	_, err = sm.UpdateSessionData(sessionId, oauthToken.AccessToken, "OneDriveAccessToken")
	if err != nil {
		return
	}

	_, err = sm.UpdateSessionData(sessionId, oauthToken.RefreshToken, "OneDriveRefreshToken")
	return
}

func setOneDriveCreds(data sm.SessionMngrResponse) (creds *externaldrive.OneDriveCreds, err *model.DashboardResponse) {
	creds, erro := externaldrive.SetOneDriveCreds(data)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Couldn't Set One Drive Credentials",
			ErrorMessage: erro.Error(),
		}
	}
	return
}

func establishOneDriveCredentials(clientID string) {
	if model.Local {
		sm.UpdateSessionData(clientID, "fff1cba9-7597-479d-b653-fd96c5d56b43", "OneDriveClientID")
		sm.UpdateSessionData(clientID, "offline_access files.read files.read.all files.readwrite files.readwrite.all", "OneDriveScopes")
	} else {
		sm.UpdateSessionData(clientID, "", "OneDriveAccessToken")
		sm.UpdateSessionData(clientID, "", "OneDriveRefreshToken")
		sm.UpdateSessionData(clientID, os.Getenv("ONE_DRIVE_CLIENT_ID"), "OneDriveClientID")
		sm.UpdateSessionData(clientID, os.Getenv("ONE_DRIVE_SCOPES"), "OneDriveScopes")
	}
}

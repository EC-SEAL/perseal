package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/EC-SEAL/perseal/externaldrive"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/EC-SEAL/perseal/utils"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
)

// ONE DRIVE SERVICE METHODS

//Attempts to Store the Session Data On OneDrive
func storeSessionDataOneDrive(data interface{}, uuid, id string, filename string, cameFrom string) (password string, dataStore *externaldrive.DataStore, err *model.DashboardResponse) {

	oauthToken, err := getOneDriveToken(data, id, cameFrom)

	dataStore, erro := externaldrive.StoreSessionData(data, uuid, password)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Couldn't Generate DataStore to be saved",
			ErrorMessage: erro.Error(),
		}
		return
	}

	var contents []byte
	contents, erro = dataStore.UploadingBlob(oauthToken)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Couldn't Generate Blob",
			ErrorMessage: erro.Error(),
		}
		return
	}
	var file *drive.File
	file, erro = dataStore.UploadOneDrive(oauthToken, contents, filename, "SEAL")
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
func loadSessionDataOneDrive(smResp interface{}, id string, filename string, cameFrom string) (file *http.Response, err *model.DashboardResponse) {
	var oauthToken *oauth2.Token
	oauthToken, err = getOneDriveToken(smResp, id, cameFrom)
	if err != nil {
		return
	}

	fmt.Println(oauthToken.AccessToken)

	checkFirstAccess := utils.RecieveCheckFirstAccess()

	if checkFirstAccess == true {
		err = &model.DashboardResponse{
			Code:    302,
			Message: "New Store Method",
		}
		return
	}

	file, erro := externaldrive.GetOneDriveItem(oauthToken, filename)
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

func getOneDriveToken(data interface{}, id string, cameFrom string) (oauthToken *oauth2.Token, err *model.DashboardResponse) {

	creds, err := setOneDriveCreds(data)
	link, oauthToken, erro := externaldrive.GetOneDriveToken(creds)

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

		utils.SendRedirect(modelRedirect)
		code := utils.RecieveCode()
		log.Println(code)

		oauthToken, err = updateNewOneDriveTokenFromCode(id, code, creds.OneDriveClientID)

		data, err = sm.GetSessionData(id, "")
		if err != nil {
			return
		}
		creds, err = setOneDriveCreds(data)
		if err != nil {
			return
		}

		log.Println("TOKEN ", oauthToken)
	} else {
		if cameFrom != "load&store" {
			modelRedirect := model.RedirectStruct{
				Redirect: false,
				URL:      "",
			}
			utils.SendRedirect(modelRedirect)
		}
	}
	if cameFrom != "load&store" {
		if sm.CurrentUser == nil {
			sm.CurrentUser = make(chan sm.SessionMngrResponse)
		}
		jsonM, _ := json.Marshal(data)
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

func setOneDriveCreds(data interface{}) (creds *externaldrive.OneDriveCreds, err *model.DashboardResponse) {
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

package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/externaldrive"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
)

// ONE DRIVE SERVICE METHODS

//Attempts to Store the Session Data On OneDrive
func storeSessionDataOneDrive(dto dto.PersistenceDTO, filename string) (dataStore *externaldrive.DataStore, err *model.HTMLResponse) {

	token, err := checkOneDriveTokenExpiry(dto)
	if err != nil {
		return
	}
	log.Println(token)

	dataStore, erro := externaldrive.StoreSessionData(dto)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         500,
			Message:      "Couldn't Generate DataStore to be saved",
			ErrorMessage: erro.Error(),
		}
		return
	}

	var contents []byte
	contents, erro = dataStore.UploadingBlob()
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         500,
			Message:      "Couldn't Generate Blob",
			ErrorMessage: erro.Error(),
		}
		return
	}
	var file *drive.File
	file, erro = dataStore.UploadOneDrive(token, contents, filename, "SEAL")
	fmt.Println(file)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         500,
			Message:      "Couldn't Upload DataStore One Drive",
			ErrorMessage: erro.Error(),
		}
	}
	return
}

// Fetches GoogleDrive Code
func loadSessionDataOneDrive(dto dto.PersistenceDTO, filename string) (file *http.Response, err *model.HTMLResponse) {

	token, err := checkOneDriveTokenExpiry(dto)
	if err != nil {
		return
	}

	file, erro := externaldrive.GetOneDriveItem(token, filename)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         404,
			Message:      "Couldn't Get One Drive Item",
			ErrorMessage: erro.Error(),
		}
		return
	}
	log.Println(file)
	return
}

func getOneDriveRedirectURL(dto dto.PersistenceDTO) (url string, err *model.HTMLResponse) {

	creds := externaldrive.SetOneDriveCreds()

	url, erro := externaldrive.GetOneDriveRedirectURL(dto.ID, creds)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         500,
			Message:      "Couldn't Generate One Drive Redirect URL",
			ErrorMessage: erro.Error(),
		}
	}
	return
}

// Fetches GoogleDrive Code
func checkOneDriveTokenExpiry(dto dto.PersistenceDTO) (token *oauth2.Token, err *model.HTMLResponse) {
	creds := externaldrive.SetOneDriveCreds()

	token, erro := externaldrive.CheckOneDriveTokenExpiry(dto.OneDriveToken, creds)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         404,
			Message:      "Error in Request to Refresh Token",
			ErrorMessage: erro.Error(),
		}
		return
	}

	return
}

func UpdateNewOneDriveTokenFromCode(id string, code string) (oauthToken *oauth2.Token, err *model.HTMLResponse) {

	creds := externaldrive.SetOneDriveCreds()

	var erro error
	oauthToken, erro = externaldrive.RequestToken(code, creds.OneDriveClientID)
	log.Println(oauthToken)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         404,
			Message:      "Couldn't Request One Drive Token",
			ErrorMessage: erro.Error(),
		}
		return
	}

	jsonM, erro := json.Marshal(oauthToken)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         404,
			Message:      "Couldn't Marshal One Drive Token",
			ErrorMessage: erro.Error(),
		}
		return
	}

	_, err = sm.UpdateSessionData(id, string(jsonM), "OneDriveToken")
	return
}

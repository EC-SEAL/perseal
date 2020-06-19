package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/externaldrive"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

// GOOGLE DRIVE SERVICE METHODS

//Attempts to Store the Session Data On Google Drive
func storeSessionDataGoogleDrive(dto dto.PersistenceDTO, filename string) (dataStore *externaldrive.DataStore, err *model.HTMLResponse) {
	token, client, err := getGoogleDriveClient(dto.GoogleAccessCreds)
	if err != nil {
		return
	}
	log.Println("TOKEN ", token)
	log.Println("ClIENT ", client)

	dataStore, erro := externaldrive.StoreSessionData(dto)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         500,
			Message:      "Couldn't Generate DataStore",
			ErrorMessage: erro.Error(),
		}
	}

	file, erro := dataStore.UploadGoogleDrive(token, client, filename)
	fmt.Println(file)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         500,
			Message:      "Couldn't Generate Uploading Blob",
			ErrorMessage: erro.Error(),
		}
	}
	return
}

//Attempts to Load a Datastore from GoogleDrive into Session
func loadSessionDataGoogleDrive(dto dto.PersistenceDTO, filename string) (file *http.Response, err *model.HTMLResponse) {

	_, client, err := getGoogleDriveClient(dto.GoogleAccessCreds)
	if err != nil {
		return
	}

	jsonM, erro := json.Marshal(dto.SMResp)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         500,
			Message:      "Couldn't Generate Marshal Session Data",
			ErrorMessage: erro.Error(),
		}
	}

	smr := &sm.SessionMngrResponse{}
	json.Unmarshal(jsonM, smr)

	file, erro = externaldrive.GetGoogleDriveFile(filename, client)
	log.Println(file)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         404,
			Message:      "Couldn't Get Google Drive File",
			ErrorMessage: erro.Error(),
		}
		return
	}
	return
}
func getGoogleRedirectURL(dto dto.PersistenceDTO) (url string, err *model.HTMLResponse) {

	var config *oauth2.Config
	config, err = establishGoogleDriveCreds()
	log.Println(config)
	if err != nil {
		return
	}
	url = externaldrive.GetGoogleLinkForDashboardRedirect(dto.ID, config)
	return
}

func getGoogleDriveClient(accessCreds string) (token *oauth2.Token, client *http.Client, err *model.HTMLResponse) {
	googleCreds, err := establishGoogleDriveCreds()
	if err != nil {
		return
	}

	token = &oauth2.Token{}
	erro := json.NewDecoder(strings.NewReader(accessCreds)).Decode(token)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         500,
			Message:      "Could not Decode Credentials to Token",
			ErrorMessage: erro.Error(),
		}
		return
	}

	fmt.Println(googleCreds)
	b2, erro := json.Marshal(googleCreds)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         500,
			Message:      "Google Creds JSON Malformed",
			ErrorMessage: erro.Error(),
		}
		return
	}

	config, erro := google.ConfigFromJSON([]byte(b2), drive.DriveFileScope)
	if err != nil {
		if erro != nil {
			err = &model.HTMLResponse{
				Code:         404,
				Message:      "Couldn't retrieve config from Google Creds JSON",
				ErrorMessage: erro.Error(),
			}
			return
		}
	}

	client = config.Client(context.Background(), token)
	return
}

// Uploads Google Drive Token to SessionVariables
func UpdateNewGoogleDriveTokenFromCode(id string, code string) (tok *oauth2.Token, err *model.HTMLResponse) {

	config, err := establishGoogleDriveCreds()
	if err != nil {
		return
	}

	tok, erro := config.Exchange(oauth2.NoContext, code)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         404,
			Message:      "Could not Fetch Google Drive Access Token",
			ErrorMessage: erro.Error(),
		}
		return
	}

	b, erro := json.Marshal(tok)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         500,
			Message:      "Couldn't Parse the Google Drive Access Token to byte array",
			ErrorMessage: erro.Error(),
		}
		return
	}

	_, err = sm.UpdateSessionData(id, string(b), "GoogleDriveAccessCreds")
	return
}

// Uploads new GoogleDrive data
func establishGoogleDriveCreds() (config *oauth2.Config, err *model.HTMLResponse) {

	googleCreds := externaldrive.SetGoogleDriveCreds()

	fmt.Println(googleCreds)
	b2, erro := json.Marshal(googleCreds)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         500,
			Message:      "Couldn't Parse the Google Drive Credentials to byte array",
			ErrorMessage: erro.Error(),
		}
		return
	}

	config, erro = google.ConfigFromJSON([]byte(b2), drive.DriveFileScope)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         500,
			Message:      "Couldn't Get Config from Google Creds",
			ErrorMessage: erro.Error(),
		}
	}

	log.Println(googleCreds)
	return

}

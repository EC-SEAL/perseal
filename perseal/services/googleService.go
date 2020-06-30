package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

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
	client := getGoogleDriveClient(dto.GoogleAccessCreds)
	log.Println(client)
	log.Println(dto.GoogleAccessCreds)
	dataStore, _ = externaldrive.StoreSessionData(dto)

	file, erro := dataStore.UploadGoogleDrive(&dto.GoogleAccessCreds, client, filename)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         500,
			Message:      "Couldn't Store DataStore",
			ErrorMessage: erro.Error(),
		}
		return
	}

	fmt.Println("file", file)
	return
}

//Attempts to Load a Datastore from GoogleDrive into Session
func loadSessionDataGoogleDrive(dto dto.PersistenceDTO, filename string) (file *http.Response, err *model.HTMLResponse) {

	client := getGoogleDriveClient(dto.GoogleAccessCreds)

	jsonM, _ := json.Marshal(dto.SMResp)
	smr := &sm.SessionMngrResponse{}
	json.Unmarshal(jsonM, smr)

	file, erro := externaldrive.GetGoogleDriveFile(filename, client)
	log.Println(file)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         404,
			Message:      "Couldn't Get Google Drive File",
			ErrorMessage: erro.Error(),
		}
	}
	return
}
func getGoogleRedirectURL(dto dto.PersistenceDTO) (url string) {

	var config *oauth2.Config
	config = establishGoogleDriveCreds()
	log.Println(config)
	url = externaldrive.GetGoogleLinkForDashboardRedirect(dto.ID, config)
	return
}

func getGoogleDriveClient(accessCreds oauth2.Token) (client *http.Client) {
	googleCreds := establishGoogleDriveCreds()

	b2, _ := json.Marshal(googleCreds)
	config, _ := google.ConfigFromJSON([]byte(b2), drive.DriveFileScope)
	client = config.Client(context.Background(), &accessCreds)
	return
}

// Uploads Google Drive Token to SessionVariables
func updateNewGoogleDriveTokenFromCode(id string, code string) (tok *oauth2.Token, err *model.HTMLResponse) {

	config := establishGoogleDriveCreds()

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
func establishGoogleDriveCreds() (config *oauth2.Config) {

	googleCreds := externaldrive.SetGoogleDriveCreds()

	fmt.Println(googleCreds)
	b2, _ := json.Marshal(googleCreds)

	config, _ = google.ConfigFromJSON([]byte(b2), drive.DriveFileScope)

	log.Println(googleCreds)
	return

}

package services

import (
	"context"
	"encoding/json"
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
func storeSessionDataGoogleDrive(dto dto.PersistenceDTO) (dataStore *externaldrive.DataStore, err *model.HTMLResponse) {
	client := getGoogleDriveClient(dto.GoogleAccessCreds)
	log.Println("Current Google Client: ", client)
	dataStore, erro := externaldrive.StoreSessionData(dto)
	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedEncryption, erro.Error())
		return
	}
	erro = dataStore.UploadGoogleDrive(client)
	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedDataStoreStoringInFile, erro.Error())
		return
	}
	return
}

//Attempts to Load a Datastore from GoogleDrive into Session
func loadSessionDataGoogleDrive(dto dto.PersistenceDTO, filename string) (file *http.Response, err *model.HTMLResponse) {

	client := getGoogleDriveClient(dto.GoogleAccessCreds)
	file, erro := getGoogleDriveFile(filename, client)
	if erro != nil {
		err = model.BuildResponse(http.StatusNotFound, model.Messages.FailedGetCloudFile+model.EnvVariables.Google_Drive_PDS, erro.Error())
	}
	return
}

func getGoogleRedirectURL(id string) (url string) {

	var config *oauth2.Config
	config = establishGoogleDriveCreds()
	log.Println(config)
	url = getGoogleLinkForDashboardRedirect(id, config)
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
		err = model.BuildResponse(http.StatusNotFound, model.Messages.FailedGetToken+model.EnvVariables.Google_Drive_PDS, erro.Error())
		return
	}

	b, erro := json.Marshal(tok)
	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedParseToken+model.EnvVariables.Google_Drive_PDS, erro.Error())
		return
	}

	err = sm.UpdateSessionData(id, string(b), model.EnvVariables.SessionVariables.GoogleDriveToken)
	return
}

// Uploads new GoogleDrive data
func establishGoogleDriveCreds() (config *oauth2.Config) {
	googleCreds := model.EnvVariables.GoogleDriveCreds
	b2, _ := json.Marshal(googleCreds)
	config, _ = google.ConfigFromJSON([]byte(b2), drive.DriveFileScope)
	return

}

func getGoogleDriveFile(filename string, client *http.Client) (file *http.Response, err error) {
	service, err := drive.New(client)
	if err != nil {
		return
	}

	list, err := service.Files.List().Do()
	if err != nil {
		return
	}
	var fileId string
	for _, v := range list.Files {
		if v.Name == filename {
			fileId = v.Id
		}
	}
	file, err = service.Files.Get(fileId).Download()
	return
}

// Requests a token from the web, then returns the retrieved token.
func getGoogleLinkForDashboardRedirect(id string, config *oauth2.Config) string {
	var authURL string
	if model.Test {
		authURL = config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("redirect_uri", model.EnvVariables.Redirect_URL), oauth2.SetAuthURLParam("state", id), oauth2.SetAuthURLParam("user_id", model.EnvVariables.Project_SEAL_Email))
	} else {
		authURL = config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("redirect_uri", model.EnvVariables.Redirect_URL), oauth2.SetAuthURLParam("state", id))
	}
	return authURL
}

func getGoogleDriveFiles(client *http.Client) (fileList []string, err error) {
	service, err := drive.New(client)
	if err != nil {
		return
	}

	list, err := service.Files.List().Do()
	if err != nil {
		return
	}
	fileList = make([]string, 0)
	for _, v := range list.Files {
		fileList = append(fileList, v.Name)
	}
	return
}

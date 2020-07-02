package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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

	file, erro := getGoogleDriveFile(filename, client)
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

	googleCreds := setGoogleDriveCreds()

	fmt.Println(googleCreds)
	b2, _ := json.Marshal(googleCreds)

	config, _ = google.ConfigFromJSON([]byte(b2), drive.DriveFileScope)

	log.Println(googleCreds)
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
		authURL = config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("redirect_uri", "http://localhost:8082/per/code"), oauth2.SetAuthURLParam("state", id), oauth2.SetAuthURLParam("user_id", "info@project-seal.eu"))
	} else {
		authURL = config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("redirect_uri", os.Getenv("REDIRECT_URL")), oauth2.SetAuthURLParam("state", id))
	}
	return authURL
}

func setGoogleDriveCreds() model.GoogleDriveCreds {
	googleCreds := &model.GoogleDriveCreds{}
	if model.Test {
		googleCreds.Web.ClientId = "425112724933-9o8u2rk49pfurq9qo49903lukp53tbi5.apps.googleusercontent.com"
		googleCreds.Web.ProjectId = "seal-274215"
		googleCreds.Web.AuthURI = "https://accounts.google.com/o/oauth2/auth"
		googleCreds.Web.TokenURI = "https://oauth2.googleapis.com/token"
		googleCreds.Web.AuthProviderx509CertUrl = "https://www.googleapis.com/oauth2/v1/certs"
		googleCreds.Web.ClientSecret = "0b3WtqfasYfWDmk31xa8UAht"
		googleCreds.Web.RedirectURIS = []string{"http://localhost:8082/per/code"}
	} else {
		googleCreds.Web.ClientId = os.Getenv("GOOGLE_DRIVE_CLIENT_ID")
		googleCreds.Web.ProjectId = os.Getenv("GOOGLE_DRIVE_CLIENT_PROJECT")
		googleCreds.Web.AuthURI = os.Getenv("GOOGLE_DRIVE_AUTH_URI")
		googleCreds.Web.TokenURI = os.Getenv("GOOGLE_DRIVE_TOKEN_URI")
		googleCreds.Web.AuthProviderx509CertUrl = os.Getenv("GOOGLE_DRIVE_AUTH_PROVIDER")
		googleCreds.Web.ClientSecret = os.Getenv("GOOGLE_DRIVE_CLIENT_SECRET")
		googleCreds.Web.RedirectURIS = strings.Split([]string{os.Getenv("GOOGLE_DRIVE_REDIRECT_URIS")}[0], ",")
	}
	return *googleCreds
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

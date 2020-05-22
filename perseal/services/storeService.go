package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/EC-SEAL/perseal/externaldrive"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

var (
	mockUUID = "77358e9d-3c59-41ea-b4e3-04922657b30c" // As ID for DataStore
)

// Store Data on the corresponding PDS
func StoreCloudData(data sm.SessionMngrResponse, pds string, clientID string, id string, filename string) (password string, dataStore *externaldrive.DataStore, err *model.DashboardResponse) {
	uuid := mockUUID

	if pds == "googleDrive" {
		// Validates if the session data contains the google drive authentication token
		if clientID == "" {
			data.Error = "Session Data Not Correctly Set - Google Drive Client Missing"
			establishGoogleCredentials(data.SessionData.SessionID)
		}
		data, err = sm.GetSessionData(id, "")
		if err != nil {
			return
		}
		password, dataStore, err = storeSessionDataGoogleDrive(data, uuid, id, filename) // No password
		return
	} else if pds == "oneDrive" {
		if data.SessionData.SessionVariables["OneDriveClientID"] == "" {
			data.Error = "Session Data Not Correctly Set - One Drive Client Missing"
			establishOneDriveCredentials(data.SessionData.SessionID)
		}
		id := data.SessionData.SessionID
		data, err = sm.GetSessionData(id, "")
		if err != nil {
			return
		}
		dataStore, err = storeSessionDataOneDrive(data, uuid, id) // No password
	} else {
		err = &model.DashboardResponse{
			Code:    404,
			Message: "Wrong Module Or No Module Found in Credentials",
		}
		return
	}
	return
}

// Back-channel store may only be used for local Browser storing
func StoreLocalData(data sm.SessionMngrResponse, pds string, cipherPassword string) (dataStore *externaldrive.DataStore, err *model.DashboardResponse) {
	uuid := mockUUID
	if pds == "googleDrive" || pds != "oneDrive" {
		var erro error
		dataStore, erro = externaldrive.StoreSessionData(data, uuid, cipherPassword)
		if erro != nil {
			err = &model.DashboardResponse{
				Code:         500,
				Message:      "Couldn't Create New DataStore and Encrypt It",
				ErrorMessage: erro.Error(),
			}
			return
		}
		return
	} else {
		err = &model.DashboardResponse{
			Code:    404,
			Message: "Bad PDS Variable",
		}
		return
	}
}

// GOOGLE DRIVE SERVICE METHODS

//Attempts to Store the Session Data On Google Drive
//May not find a token, in which it throws a redirect link for user login to the dashboard
func storeSessionDataGoogleDrive(data interface{}, uuid, id string, filename string) (password string, dataStore *externaldrive.DataStore, err *model.DashboardResponse) {

	log.Println("entered")
	config, err := refreshGoogleDriveCreds(data)

	if externaldrive.AccessCreds == "" {
		var authURL string
		authURL = externaldrive.GetGoogleLinkForDashboardRedirect(config)
		modelRedirect := model.RedirectStruct{
			Redirect: true,
			URL:      authURL,
		}

		log.Println("working on redirect")
		model.Redirect <- modelRedirect
		log.Println("modelredirect ", model.Redirect)
		model.Code = make(chan string)
		code := <-model.Code
		err = updateNewGoogleDriveTokenFromCode(config, id, code)
		if err != nil {
			return
		}

		data, err = sm.GetSessionData(id, "")
		if err != nil {
			return
		}
		config, err = refreshGoogleDriveCreds(data)
		if err != nil {
			return
		}
	}
	var oauthToken *oauth2.Token = &oauth2.Token{}
	erro := json.NewDecoder(strings.NewReader(externaldrive.AccessCreds)).Decode(oauthToken)
	log.Println(erro)

	fmt.Println(oauthToken.AccessToken)

	model.Password = make(chan string)
	password = <-model.Password
	log.Println(password)
	close(model.Password)

	dataStore, erro = externaldrive.StoreSessionData(data, uuid, password)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Couldn't Create New DataStore and Encrypt It",
			ErrorMessage: erro.Error(),
		}
		return
	}

	client := config.Client(context.Background(), oauthToken)
	file, erro := dataStore.UploadGoogleDrive(oauthToken, client, filename)
	fmt.Println(file)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Couldn't Generate Uploading Blob",
			ErrorMessage: erro.Error(),
		}
	}
	return
}

func updateNewGoogleDriveTokenFromCode(config *oauth2.Config, sessionId string, code string) (err *model.DashboardResponse) {

	tok, erro := config.Exchange(oauth2.NoContext, code)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Could not Fetch Google Drive Access Token",
			ErrorMessage: erro.Error(),
		}
		return
	}

	b, erro := json.Marshal(tok)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Couldn't Parse the Google Drive Access Token to byte array",
			ErrorMessage: erro.Error(),
		}
		return
	}

	_, err = sm.UpdateSessionData(sessionId, string(b), "GoogleDriveAccessCreds")
	return
}

func refreshGoogleDriveCreds(data interface{}) (config *oauth2.Config, err *model.DashboardResponse) {

	googleCreds := externaldrive.SetGoogleDriveCreds(data)

	fmt.Println(googleCreds)
	b2, erro := json.Marshal(googleCreds)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Couldn't Parse the Google Drive Credentials to byte array",
			ErrorMessage: erro.Error(),
		}
		return
	}

	config, erro = google.ConfigFromJSON([]byte(b2), drive.DriveFileScope)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Couldn't Get Config from Google Creds",
			ErrorMessage: erro.Error(),
		}
	}

	log.Println(googleCreds)
	log.Println(externaldrive.AccessCreds)
	return

}

// ONE DRIVE SERVICE METHODS

//Attempts to Store the Session Data On OneDrive
//May not find a token, in which it throws a redirect link for user login to the dashboard
func storeSessionDataOneDrive(data interface{}, uuid, id string) (dataStore *externaldrive.DataStore, err *model.DashboardResponse) {
	creds, err := setOneDriveCreds(data)
	link, oauthToken, erro := externaldrive.GetOneDriveToken(creds)

	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Couldn't Get One Drive Token",
			ErrorMessage: erro.Error(),
		}
		return
	}
	log.Println(oauthToken)

	model.Password = make(chan string)
	password := <-model.Password
	log.Println(password)
	close(model.Password)

	if link != "" {
		model.Code = make(chan string)
		code := <-model.Code
		err = updateNewOneDriveTokenFromCode(id, code, creds.OneDriveClientID)

		data, err = sm.GetSessionData(id, "")
		if err != nil {
			return
		}
		creds, err = setOneDriveCreds(data)
		if err != nil {
			return
		}

		link, oauthToken, erro = externaldrive.GetOneDriveToken(creds)
		log.Println("TOKEN ", oauthToken)
	}

	dataStore, erro = externaldrive.StoreSessionData(data, uuid, password)
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
	file, erro = dataStore.UploadOneDrive(oauthToken, contents, "SEAL")
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

func updateNewOneDriveTokenFromCode(sessionId string, code string, id string) (err *model.DashboardResponse) {

	oauthToken, erro := externaldrive.RequestToken(code, id)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
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

func establishGoogleCredentials(clientID string) {
	if model.Local {
		sm.UpdateSessionData(clientID, "425112724933-9o8u2rk49pfurq9qo49903lukp53tbi5.apps.googleusercontent.com", "GoogleDriveClientID")
		sm.UpdateSessionData(clientID, "seal-274215", "GoogleDriveClientProject")
		sm.UpdateSessionData(clientID, "https://accounts.google.com/o/oauth2/auth", "GoogleDriveAuthURI")
		sm.UpdateSessionData(clientID, "https://oauth2.googleapis.com/token", "GoogleDriveTokenURI")
		sm.UpdateSessionData(clientID, "https://www.googleapis.com/oauth2/v1/certs", "GoogleDriveAuthProviderx509CertUrl")
		sm.UpdateSessionData(clientID, "0b3WtqfasYfWDmk31xa8UAht", "GoogleDriveClientSecret")
		sm.UpdateSessionData(clientID, "http://localhost:4200/code,https://vm.project-seal.eu:4200/code,https://perseal.seal.eu:4200/code", "GoogleDriveRedirectUris")
	} else {
		sm.UpdateSessionData(clientID, os.Getenv("GOOGLE_DRIVE_CLIENT_ID"), "GoogleDriveClientID")
		sm.UpdateSessionData(clientID, os.Getenv("GOOGLE_DRIVE_CLIENT_PROJECT"), "GoogleDriveClientProject")
		sm.UpdateSessionData(clientID, os.Getenv("GOOGLE_DRIVE_AUTH_URI"), "GoogleDriveAuthURI")
		sm.UpdateSessionData(clientID, os.Getenv("GOOGLE_DRIVE_TOKEN_URI"), "GoogleDriveTokenURI")
		sm.UpdateSessionData(clientID, os.Getenv("GOOGLE_DRIVE_AUTH_PROVIDER"), "GoogleDriveAuthProviderx509CertUrl")
		sm.UpdateSessionData(clientID, os.Getenv("GOOGLE_DRIVE_CLIENT_SECRET"), "GoogleDriveClientSecret")
		sm.UpdateSessionData(clientID, os.Getenv("GOOGLE_DRIVE_REDIRECT_URIS"), "GoogleDriveRedirectUris")
	}
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

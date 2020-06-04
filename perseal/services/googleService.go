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
	"github.com/EC-SEAL/perseal/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

// GOOGLE DRIVE SERVICE METHODS

//Attempts to Store the Session Data On Google Drive
func storeSessionDataGoogleDrive(dto dto.PersistenceDTO, filename string) (returningdto dto.PersistenceDTO, dataStore *externaldrive.DataStore, err *model.DashboardResponse) {

	returningdto = dto
	returningdto, client, err := getGoogleToken(dto)
	if err != nil {
		return
	}
	if dto.StopProcess == true {
		return
	}
	fmt.Println(returningdto.Token.AccessToken)

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
			Message:      "Couldn't Create New DataStore and Encrypt It",
			ErrorMessage: erro.Error(),
		}
		return
	}

	file, erro := dataStore.UploadGoogleDrive(returningdto.Token, client, filename)
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

//Attempts to Load a Datastore from GoogleDrive into Session
func loadSessionDataGoogleDrive(dto dto.PersistenceDTO, filename string) (returningdto dto.PersistenceDTO, file *http.Response, err *model.DashboardResponse) {

	returningdto = dto

	returningdto, client, err := getGoogleToken(returningdto)
	if err != nil {
		return
	}
	if returningdto.StopProcess == true {
		return
	}

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
	file, erro := externaldrive.GetGoogleDriveFile(filename, client)
	log.Println(file)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         404,
			Message:      "Couldn't Get Google Drive File",
			ErrorMessage: erro.Error(),
		}
		return
	}
	return
}

// Fetches GoogleDrive Code
func getGoogleToken(dto dto.PersistenceDTO) (returningdto dto.PersistenceDTO, client *http.Client, err *model.DashboardResponse) {

	returningdto = dto
	config, err := refreshGoogleDriveCreds(dto.SMResp)
	if err != nil {
		return
	}
	log.Println("access cred ", externaldrive.AccessCreds)

	// If no Token was Found
	if externaldrive.AccessCreds == "" {

		var authURL string
		authURL = externaldrive.GetGoogleLinkForDashboardRedirect(config)
		modelRedirect := model.RedirectStruct{
			Redirect: true,
			URL:      authURL,
		}

		returningdto.StopProcess = utils.SendRedirect(modelRedirect)
		if returningdto.StopProcess == true {
			return
		}
		code := utils.RecieveCode()
		log.Println(code)

		log.Println(config)
		err = updateNewGoogleDriveTokenFromCode(config, returningdto.ID, code)
		if err != nil {
			return
		}

		returningdto.SMResp, err = sm.GetSessionData(returningdto.ID, "")
		if err != nil {
			return
		}

		config, err = refreshGoogleDriveCreds(returningdto.SMResp)
		if err != nil {
			return
		}

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
	}

	returningdto.Token = &oauth2.Token{}
	erro := json.NewDecoder(strings.NewReader(externaldrive.AccessCreds)).Decode(returningdto.Token)

	log.Println(erro)

	client = config.Client(context.Background(), returningdto.Token)
	return
}

func getGoogleDriveClient(smResp sm.SessionMngrResponse) (client *http.Client, err *model.DashboardResponse) {
	googleCreds := externaldrive.SetGoogleDriveCreds(smResp)
	var token *oauth2.Token = &oauth2.Token{}
	erro := json.NewDecoder(strings.NewReader(externaldrive.AccessCreds)).Decode(token)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Could not Decode Credentials to Token",
			ErrorMessage: erro.Error(),
		}
		return
	}

	fmt.Println(googleCreds)
	b2, erro := json.Marshal(googleCreds)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Google Creds JSON Malformed",
			ErrorMessage: erro.Error(),
		}
		return
	}

	config, erro := google.ConfigFromJSON([]byte(b2), drive.DriveFileScope)
	if err != nil {
		if erro != nil {
			err = &model.DashboardResponse{
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
func updateNewGoogleDriveTokenFromCode(config *oauth2.Config, sessionId string, code string) (err *model.DashboardResponse) {

	tok, erro := config.Exchange(oauth2.NoContext, code)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         404,
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

// Uploads new GoogleDrive data
func refreshGoogleDriveCreds(data sm.SessionMngrResponse) (config *oauth2.Config, err *model.DashboardResponse) {

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

package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/EC-SEAL/perseal/sm"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v2"
)

var (
	mockUUID = "77358e9d-3c59-41ea-b4e3-04922657b30c" // As ID for DataStore
)

func persistenceStoreWithCode(code string, sessionId string, module string) {
	sessionData, _ := sm.GetSessionData(sessionId, "")
	uuid := mockUUID
	if module == "googleDrive" {
		storeSessionDataGoogleDriveWithCode(sessionData, uuid, "", sessionId, code)
	}
	if module == "oneDrive" {
		storeSessionDataOneDriveWithCode(sessionData, uuid, "", sessionId, code)
	}
}

// persistenceStore handles /per/store request - Save session data to the configured persistence mechanism (front channel).
func storeData(data sm.SessionMngrResponse, pds string, clientID string, cipherPassword string) (dataStore *DataStore, redirect *Redirect, err error) {
	uuid := mockUUID

	if pds == "googleDrive" {
		// Validates if the session data contains the google drive authentication token
		if clientID == "" {
			data.Error = "Session Data Not Correctly Set - Google Drive Client Missing"
			establishGoogleCredentials(data.SessionData.SessionID)
		}

		data, err = sm.GetSessionData(data.SessionData.SessionID, "")
		dataStore, redirect, err := storeSessionDataGoogleDrive(data, uuid, clientID, cipherPassword) // No password

		log.Println(dataStore)
		log.Println(redirect)
		log.Println(err)

		if redirect != nil {
			return nil, redirect, nil
		}
		if err != nil {
			return nil, nil, err
		}
		return dataStore, nil, nil
	}

	if pds == "oneDrive" {

		if data.SessionData.SessionVariables["OneDriveClientID"] == "" {
			data.Error = "Session Data Not Correctly Set - One Drive Client Missing"
			establishOneDriveCredentials(data.SessionData.SessionID)
		}

		dataStore, redirect, err = storeSessionDataOneDrive(data, uuid, cipherPassword) // No password
		if redirect != nil {
			redirect.SessionID = data.SessionData.SessionID
			return nil, redirect, nil
		}

		if err != nil {
			return nil, nil, err
		}

		return dataStore, nil, nil
	}
	return
}

func establishGoogleCredentials(clientID string) {
	//sm.UpdateSessionData(clientID, "425112724933-9o8u2rk49pfurq9qo49903lukp53tbi5.apps.googleusercontent.com", "GoogleDriveClientID")
	//sm.UpdateSessionData(clientID, "seal-274215", "GoogleDriveClientProject")
	//sm.UpdateSessionData(clientID, "https://accounts.google.com/o/oauth2/auth", "GoogleDriveAuthURI")
	//sm.UpdateSessionData(clientID, "https://oauth2.googleapis.com/token", "GoogleDriveTokenURI")
	//sm.UpdateSessionData(clientID, "https://www.googleapis.com/oauth2/v1/certs", "GoogleDriveAuthProviderx509CertUrl")
	//sm.UpdateSessionData(clientID, "0b3WtqfasYfWDmk31xa8UAht", "GoogleDriveClientSecret")
	//sm.UpdateSessionData(clientID, "https://localhost:8082/per/code,https://vm.project-seal.eu:8082/per/code,https://perseal.seal.eu:8082/per/code", "GoogleDriveRedirectUris")

	sm.UpdateSessionData(clientID, os.Getenv("GOOGLE_DRIVE_CLIENT_ID"), "GoogleDriveClientID")
	sm.UpdateSessionData(clientID, os.Getenv("GOOGLE_DRIVE_CLIENT_PROJECT"), "GoogleDriveClientProject")
	sm.UpdateSessionData(clientID, os.Getenv("GOOGLE_DRIVE_AUTH_URI"), "GoogleDriveAuthURI")
	sm.UpdateSessionData(clientID, os.Getenv("GOOGLE_DRIVE_TOKEN_URI"), "GoogleDriveTokenURI")
	sm.UpdateSessionData(clientID, os.Getenv("GOOGLE_DRIVE_AUTH_PROVIDER"), "GoogleDriveAuthProviderx509CertUrl")
	sm.UpdateSessionData(clientID, os.Getenv("GOOGLE_DRIVE_CLIENT_SECRET"), "GoogleDriveClientSecret")
	sm.UpdateSessionData(clientID, os.Getenv("GOOGLE_DRIVE_REDIRECT_URIS"), "GoogleDriveRedirectUris")

}

func establishOneDriveCredentials(clientID string) {
	//sm.UpdateSessionData(clientID, "fff1cba9-7597-479d-b653-fd96c5d56b43", "OneDriveClientID")
	//sm.UpdateSessionData(clientID, "offline_access files.read files.read.all files.readwrite files.readwrite.all", "OneDriveScopes")

	sm.UpdateSessionData(clientID, os.Getenv("ONE_DRIVE_CLIENT_ID"), "OneDriveClientID")
	sm.UpdateSessionData(clientID, os.Getenv("ONE_DRIVE_SCOPES"), "OneDriveScopes")
}

func storeSessionDataGoogleDrive(data interface{}, uuid, id string, cipherPassword string) (dataStore *DataStore, redirect *Redirect, err error) {
	setGoogleDriveCreds(data)

	b2, _ := json.Marshal(googleCreds)
	config, err := google.ConfigFromJSON([]byte(b2), drive.DriveFileScope)

	oauthToken, err := TokenFromSessionData()
	log.Println(oauthToken.AccessToken)

	if oauthToken.AccessToken != "" {
		dataStore, err = storeSessionData(data, uuid, cipherPassword)
		client := config.Client(context.Background(), oauthToken)
		_, err = dataStore.UploadGoogleDrive(oauthToken, client)
		return
	} else {

		log.Println(googleCreds)
		if err != nil {
			desc, authURL := getGoogleLinkForDashboardRedirect(config)
			log.Println(desc)
			redirect = &Redirect{
				SessionID:   id,
				Description: desc,
				Link:        authURL,
				Module:      "googleDrive",
				//Module:    os.Getenv("GOOGLE_DRIVE")
			}
			return
		}
		return
	}
}

func storeSessionDataGoogleDriveWithCode(data interface{}, uuid string, cipherPassword string, sessionId string, code string) (datastore *DataStore) {
	setGoogleDriveCreds(data)

	log.Println(googleCreds.Web.RedirectURIS[2])
	b2, _ := json.Marshal(googleCreds)
	config, err := google.ConfigFromJSON([]byte(b2), drive.DriveFileScope)
	log.Println(config)

	tok, err := config.Exchange(oauth2.NoContext, code)
	b, _ := json.Marshal(tok)
	sm.UpdateSessionData(sessionId, string(b), "GoogleDriveAccessCreds")
	log.Println(string(b))
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}

	datastore, err = storeSessionData(data, uuid, cipherPassword)
	client := config.Client(context.Background(), tok)
	_, err = datastore.UploadGoogleDrive(tok, client)
	return
}

func storeSessionDataOneDrive(data interface{}, uuid, cipherPassword string) (datastore *DataStore, redirect *Redirect, err error) {
	datastore, _ = storeSessionData(data, uuid, cipherPassword)
	setOneDriveCreds(data)
	redirect, oauthToken := getOneDriveToken()

	if redirect != nil {
		return nil, redirect, nil
	} else {
		contents, _ := datastore.UploadingBlob(oauthToken)
		_, err = datastore.UploadOneDrive(oauthToken, contents)
	}
	return
}

func storeSessionDataOneDriveWithCode(data interface{}, uuid, cipherPassword string, sessionId string, code string) (datastore *DataStore, err error) {
	datastore, _ = storeSessionData(data, uuid, cipherPassword)
	setOneDriveCreds(data)
	oauthToken := requestToken(code, creds.OneDriveClientID)
	sm.UpdateSessionData(sessionId, oauthToken.AccessToken, "OneDriveAccessToken")
	sm.UpdateSessionData(sessionId, oauthToken.RefreshToken, "OneDriveRefreshToken")
	contents, _ := datastore.UploadingBlob(oauthToken)
	_, err = datastore.UploadOneDrive(oauthToken, contents)
	return
}

/*
func storeFileOneDriveClearText(data sm.SessionMngrResponse, uuid, cipherPassword string, contents interface{}) (datastore *DataStore, err error) {
	datastore, _ = storeSessionData(data, uuid, cipherPassword)
	setOneDriveCreds(data)
	redirect, oauthToken := getOneDriveToken()
	if redirect != nil {

	}
	b, _ := json.Marshal(contents)
	_, err = datastore.UploadOneDrive(oauthToken, b)
	return
}

*/

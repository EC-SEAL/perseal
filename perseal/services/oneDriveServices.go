package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

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

	token, err := checkOneDriveTokenExpiry(dto.OneDriveToken)
	if err != nil {
		return
	}
	dataStore, _ = externaldrive.StoreSessionData(dto)

	var contents []byte
	contents, _ = dataStore.UploadingBlob()

	var file *drive.File
	file, erro := dataStore.UploadOneDrive(token, contents, filename, "SEAL")
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

	token, err := checkOneDriveTokenExpiry(dto.OneDriveToken)
	if err != nil {
		return
	}

	file, erro := getOneDriveItem(token, filename)
	log.Println(file.StatusCode)
	if erro != nil || file.StatusCode != 200 {
		err = &model.HTMLResponse{
			Code:    404,
			Message: "Couldn't Get One Drive Item",
		}
		return
	}
	log.Println(file)
	return
}

func updateNewOneDriveTokenFromCode(id string, code string) (oauthToken *oauth2.Token, err *model.HTMLResponse) {

	creds := setOneDriveCreds()

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

func getOneDriveItems(token *oauth2.Token, folder string) (folderchildren *externaldrive.FolderChildren, err error) {

	folderId := "5C07F9D77D4396CC!106"

	var url string
	if model.Test {
		url = "https://graph.microsoft.com/v1.0/me/drive/items/" + folderId + "/children"
	} else {
		url = os.Getenv("GET_ITEMS_URL") + folderId + "/children"
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	log.Println("Token ", token.AccessToken)
	auth := "Bearer " + token.AccessToken
	req.Header.Add("Authorization", auth)

	client := &http.Client{}
	resp, err := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)

	var v interface{}
	json.Unmarshal([]byte(body), &v)
	jsonM, _ := json.Marshal(v)

	json.Unmarshal(jsonM, &folderchildren)

	log.Println(v)
	return
}

func getOneDriveItem(token *oauth2.Token, item string) (resp *http.Response, err error) {

	var url string
	url = "https://graph.microsoft.com/v1.0/me/drive/root:/SEAL/" + item + ":/content"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	auth := "Bearer " + token.AccessToken
	req.Header.Add("Authorization", auth)

	client := &http.Client{}
	resp, err = client.Do(req)
	return
}

// Returns Oauth Token used for authorization of OneDrive requests
func checkOneDriveTokenExpiry(token oauth2.Token) (rtoken *oauth2.Token, err *model.HTMLResponse) {
	creds := setOneDriveCreds()

	now := time.Now()
	end := token.Expiry

	//if the access token hasn't expired yet
	if end.Sub(now) > 10 {
		rtoken = &token
		return
	}

	rtoken, erro := externaldrive.RequestRefreshToken(creds.OneDriveClientID, &token)

	if erro != nil {
		err = &model.HTMLResponse{
			Code:         404,
			Message:      "Error in Request to Refresh Token",
			ErrorMessage: erro.Error(),
		}
		return
	}
	//if the access token has expired. Makes a refresh token request
	return
}

// Requests a token from the web, then returns the retrieved token.
// Makes a GET request to retrive a Code. The user needs to copy paste the code on the console
// Afterwards, makes a POST request to retrive the new access_token, given necessary parameters
// In order to use the One Drive API, the client needs the clientID, the redirect_uri and the scopes of the application in the Microsfot Graph
// For more information, follow this link: https://docs.microsoft.com/en-us/onedrive/developer/rest-api/getting-started/graph-oauth?view=odsp-graph-online
func getOneDriveRedirectURL(id string) (link string) {
	creds := setOneDriveCreds()
	var u *url.URL
	//Retrieve the code
	if model.Test {
		u, _ = url.ParseRequestURI("https://login.live.com/oauth20_authorize.srf")
	} else {
		u, _ = url.ParseRequestURI(os.Getenv("AUTH_URL"))
	}
	urlStr := u.String()

	req, _ := http.NewRequest("GET", urlStr, nil)

	q := req.URL.Query()
	q.Add("client_id", creds.OneDriveClientID)
	q.Add("scope", creds.OneDriveScopes)
	if model.Test {
		q.Add("redirect_uri", "http://localhost:8082/per/code")
	} else {
		q.Add("redirect_uri", os.Getenv("REDIRECT_URL_HTTPS"))
	}
	q.Add("response_type", "code")
	q.Add("state", id)
	req.URL.RawQuery = q.Encode()

	link = req.URL.String()

	return link
}

func setOneDriveCreds() (creds *model.OneDriveCreds) {
	creds = &model.OneDriveCreds{}

	if model.Test {
		creds.OneDriveClientID = "fff1cba9-7597-479d-b653-fd96c5d56b43"
		creds.OneDriveScopes = "offline_access files.read files.read.all files.readwrite files.readwrite.all"
	} else {
		creds.OneDriveClientID = os.Getenv("ONE_DRIVE_CLIENT_ID")
		creds.OneDriveScopes = os.Getenv("ONE_DRIVE_SCOPES")
	}

	return
}

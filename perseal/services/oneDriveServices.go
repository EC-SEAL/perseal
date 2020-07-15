package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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
func storeSessionDataOneDrive(dto dto.PersistenceDTO) (dataStore *externaldrive.DataStore, err *model.HTMLResponse) {

	token, err := checkOneDriveTokenExpiry(dto.OneDriveToken)
	if err != nil {
		return
	}
	dataStore, erro := externaldrive.StoreSessionData(dto)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         500,
			Message:      "Encryption Failed",
			ErrorMessage: erro.Error(),
		}
		return
	}

	var contents []byte
	contents, _ = dataStore.UploadingBlob()

	var file *drive.File
	file, erro = dataStore.UploadOneDrive(token, contents)
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
	creds := model.EnvVariables.OneDriveCreds

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

	_, err = sm.UpdateSessionData(id, string(jsonM), model.EnvVariables.SessionVariables.OneDriveToken)
	return
}

func getOneDriveItems(token *oauth2.Token) (folderchildren *externaldrive.FolderChildren, err error) {

	id, err := getOneDriveId(token)

	url := model.EnvVariables.OneDriveURLs.Get_Items + id.ID + "/children"

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
	url = model.EnvVariables.OneDriveURLs.Get_Item + model.EnvVariables.DataStore_Folder_Name + "/" + item + "/:/content"
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

func getOneDriveId(token *oauth2.Token) (id *externaldrive.FolderProps, err error) {

	var url string
	url = model.EnvVariables.OneDriveURLs.Get_Item + model.EnvVariables.DataStore_Folder_Name
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	auth := "Bearer " + token.AccessToken
	req.Header.Add("Authorization", auth)

	client := &http.Client{}
	resp, err := client.Do(req)

	body, _ := ioutil.ReadAll(resp.Body)
	var v interface{}
	json.Unmarshal([]byte(body), &v)
	jsonM, _ := json.Marshal(v)

	json.Unmarshal(jsonM, &id)
	return
}

// Returns Oauth Token used for authorization of OneDrive requests
func checkOneDriveTokenExpiry(token oauth2.Token) (rtoken *oauth2.Token, err *model.HTMLResponse) {
	creds := model.EnvVariables.OneDriveCreds
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
	creds := model.EnvVariables.OneDriveCreds
	req, _ := http.NewRequest("GET", model.EnvVariables.OneDriveURLs.Auth, nil)

	q := req.URL.Query()
	q.Add("client_id", creds.OneDriveClientID)
	q.Add("scope", creds.OneDriveScopes)
	q.Add("redirect_uri", model.EnvVariables.Redirect_URL)
	q.Add("response_type", "code")
	q.Add("state", id)
	req.URL.RawQuery = q.Encode()

	link = req.URL.String()

	return link
}

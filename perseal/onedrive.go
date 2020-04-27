// Package onedrive implements the OAuth2 protocol for authenticating users through onedrive.
// This package can be used as a reference implementation of an OAuth2 provider for Goth.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/EC-SEAL/perseal/sm"
	"golang.org/x/oauth2"
)

const (
	dataStoreFile string = "datastore.txt"
	folderName    string = "SEAL"
)

// TokenRequestResponse - The http response after token request to One Drive API
type TokenRequestResponse struct {
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	ExtExpiresIn int    `json:"ext_expires_in"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type OneDriveCreds struct {
	OneDriveClientID     string `json:"oneDriveClient"`
	OneDriveScopes       string `json:"oneDriveScopes"`
	OneDriveAccessToken  string `json:"oneDrivetAccessToken"`
	OneDriveRefreshToken string `json:"oneDrivetRefreshToken"`
}

// FolderProps - The properties of the One Drive folder
type FolderProps struct {
	ID string `json:"id"`
}

// Used to control token expiration
var currentOneDriveToken oauth2.Token

var creds *OneDriveCreds

// POST request to create a folder in the root
func createOneDriveFolder(token *oauth2.Token) string {
	createfolderjson := []byte(`{"name":"` + folderName + `","folder": {},"@microsoft.graph.conflictBehavior": "rename"}`)
	//	req, _ := http.NewRequest("POST", os.Getenv("CREATE_FOLDER_URL"), bytes.NewBuffer(createfolderjson))
	req, _ := http.NewRequest("POST", "https://graph.microsoft.com/v1.0/me/drive/root/children", bytes.NewBuffer(createfolderjson))
	auth := "Bearer " + token.AccessToken
	req.Header.Add("Authorization", auth)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, _ := client.Do(req)
	folderID := getOneDriveFolderID(resp)
	return folderID
}

// PUT request to create a file in a given folder
func createOneDriveFile(token *oauth2.Token, folderID string, blob []byte) {
	//	url := os.Getenv("CREATE_FILE_URL") + folderID + ":/" + dataStoreFile + ":/content"
	url := "https://graph.microsoft.com/v1.0/me/drive/items/" + folderID + ":/" + dataStoreFile + ":/content"
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(blob))
	auth := "Bearer " + token.AccessToken
	req.Header.Add("Authorization", auth)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, _ := client.Do(req)

	fmt.Println(resp)
}

// GET request to fetch information of a given folder
func getOneDriveFolder(token *oauth2.Token, folder string) *http.Response {
	//url := os.Getenv("GET_FOLDER_URL") + ":/" + folder
	url := "https://graph.microsoft.com/v1.0/me/drive/root" + ":/" + folder
	req, _ := http.NewRequest("GET", url, nil)
	auth := "Bearer " + token.AccessToken
	req.Header.Add("Authorization", auth)

	client := &http.Client{}
	resp, _ := client.Do(req)

	return resp
}

// Auxiliary method: returns ID of a given folder (from previous http response)
func getOneDriveFolderID(resp *http.Response) string {
	var folderprops FolderProps
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal([]byte(body), &folderprops)
	return folderprops.ID
}

// Returns Oauth Token used for authorization of OneDrive requests
func getOneDriveToken() (redirect *Redirect, token *oauth2.Token) {

	currentOneDriveToken := oauth2.Token{
		AccessToken:  creds.OneDriveAccessToken,
		RefreshToken: creds.OneDriveRefreshToken,
		TokenType:    "Bearer",
		Expiry:       time.Now().Local().Add(time.Second * time.Duration(3600)),
	}

	if currentOneDriveToken.AccessToken == "" {
		url := getCodeFromWeb(creds.OneDriveClientID, creds.OneDriveScopes)
		desc := `Go to the following link ` + url + `"and login to your Account"`

		redirect = &Redirect{
			Description: desc,
			Link:        url,
			Module:      "oneDrive",
		}

		return redirect, nil
	}

	now := time.Now()
	end := currentOneDriveToken.Expiry

	//if the access token hasn't expired yet
	if end.Sub(now) > 10 {
		return nil, &currentOneDriveToken
	}

	//if the access token has expired. Makes a refresh token request
	return nil, requestRefreshToken(creds.OneDriveClientID, &currentOneDriveToken)
}

// Requests a token from the web, then returns the retrieved token.
// Makes a GET request to retrive a Code. The user needs to copy paste the code on the console
// Afterwards, makes a POST request to retrive the new access_token, given necessary parameters
// In order to use the One Drive API, the client needs the clientID, the redirect_uri and the scopes of the application in the Microsfot Graph
// For more information, follow this link: https://docs.microsoft.com/en-us/onedrive/developer/rest-api/getting-started/graph-oauth?view=odsp-graph-online
func getCodeFromWeb(clientID string, scopes string) string {

	//Retrieve the code
	//u, err := url.ParseRequestURI(os.Getenv("AUTH_URL"))
	u, err := url.ParseRequestURI("https://login.live.com/oauth20_authorize.srf")
	if err != nil {
		log.Fatalf("Unable to read url: %v", err)
	}
	urlStr := u.String()

	req, _ := http.NewRequest("GET", urlStr, nil)
	q := req.URL.Query()
	q.Add("client_id", "fff1cba9-7597-479d-b653-fd96c5d56b43")
	q.Add("scope", scopes)
	q.Add("redirect_uri", "https://localhost:8082/per/code")
	q.Add("response_type", "code")
	req.URL.RawQuery = q.Encode()

	log.Println("fff1cba9-7597-479d-b653-fd96c5d56b43")

	return req.URL.String()
}
func requestToken(code string, clientID string) *oauth2.Token {

	//Retrieve the access token
	values := url.Values{}
	values.Add("client_id", "fff1cba9-7597-479d-b653-fd96c5d56b43")
	values.Add("code", code)
	values.Add("grant_type", "authorization_code")
	//	values.Add("redirect_uri", os.Getenv("REDIRECT_URL"))
	values.Add("redirect_uri", "https://localhost:8082/per/code")

	//	u, _ := url.ParseRequestURI(os.Getenv("FETCH_TOKEN_URL"))
	u, _ := url.ParseRequestURI("https://login.microsoftonline.com/common/oauth2/v2.0/token")
	urlStr := u.String()
	req, _ := http.NewRequest("POST", urlStr, strings.NewReader(values.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	log.Println(req.Body)
	return tokenRequest(req)
}

// POST request to retrieve a new access and refresh tokens
func requestRefreshToken(clientID string, token *oauth2.Token) *oauth2.Token {
	values := url.Values{}
	values.Add("client_id", "fff1cba9-7597-479d-b653-fd96c5d56b43")
	values.Add("refresh_token", token.RefreshToken)
	values.Add("grant_type", "refresh_token")
	values.Add("redirect_uri", os.Getenv("REDIRECT_URL"))

	u, _ := url.ParseRequestURI(os.Getenv("FETCH_TOKEN_URL"))
	urlStr := u.String()
	req, _ := http.NewRequest("POST", urlStr, strings.NewReader(values.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return tokenRequest(req)
}

//Auxiliary method: performs a token-related http request
func tokenRequest(req *http.Request) *oauth2.Token {

	client := &http.Client{}
	var respo TokenRequestResponse

	resp, err := client.Do(req)

	if err != nil {
		log.Fatalf("Request Failed")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Couldn't read response body")
	}
	err = json.Unmarshal([]byte(body), &respo)
	if err != nil {
		log.Fatalf("Unable to unmarshall JSON content")
	}

	tok := &oauth2.Token{
		AccessToken:  respo.AccessToken,
		RefreshToken: respo.RefreshToken,
		TokenType:    respo.TokenType,
		Expiry:       time.Now().Local().Add(time.Second * time.Duration(respo.ExpiresIn)),
	}

	currentOneDriveToken = *tok

	return tok
}

func setOneDriveCreds(data interface{}) {
	creds = &OneDriveCreds{}
	smr := &sm.SessionMngrResponse{}
	jsonM, _ := json.Marshal(data)
	json.Unmarshal(jsonM, smr)
	creds.OneDriveClientID = smr.SessionData.SessionVariables["OneDriveClientID"]
	creds.OneDriveScopes = smr.SessionData.SessionVariables["OneDriveScopes"]
	creds.OneDriveAccessToken = smr.SessionData.SessionVariables["OneDriveAccessToken"]
	creds.OneDriveRefreshToken = smr.SessionData.SessionVariables["OneDriveRefreshToken"]
}

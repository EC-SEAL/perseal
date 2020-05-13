// Package onedrive implements the OAuth2 protocol for authenticating users through onedrive.
// This package can be used as a reference implementation of an OAuth2 provider for Goth.
package externaldrive

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

	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
	"golang.org/x/oauth2"
)

const (
	folderName string = "SEAL"
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
func CreateOneDriveFolder(token *oauth2.Token) (folderID string, err error) {
	createfolderjson := []byte(`{"name":"` + folderName + `","folder": {},"@microsoft.graph.conflictBehavior": "rename"}`)
	var req *http.Request
	if model.Local {
		req, err = http.NewRequest("POST", "https://graph.microsoft.com/v1.0/me/drive/root/children", bytes.NewBuffer(createfolderjson))
	} else {
		req, _ = http.NewRequest("POST", os.Getenv("CREATE_FOLDER_URL"), bytes.NewBuffer(createfolderjson))
	}
	if err != nil {
		return
	}

	auth := "Bearer " + token.AccessToken
	req.Header.Add("Authorization", auth)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	folderID, err = GetOneDriveFolderID(resp)
	return
}

// PUT request to create a file in a given folder
func CreateOneDriveFile(token *oauth2.Token, folderID string, blob []byte) (err error) {
	var url string
	if model.Local {
		url = "https://graph.microsoft.com/v1.0/me/drive/items/" + folderID + ":/datastore.txt:/content"
	} else {
		url = os.Getenv("CREATE_FILE_URL") + folderID + ":/" + os.Getenv("DATA_STORE_FILENAME") + ":/content"
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(blob))
	if err != nil {
		return
	}
	auth := "Bearer " + token.AccessToken
	req.Header.Add("Authorization", auth)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	fmt.Println(resp)
	return
}

// GET request to fetch information of a given folder
func GetOneDriveFolder(token *oauth2.Token, folder string) (resp *http.Response, err error) {

	var url string
	if model.Local {
		url = "https://graph.microsoft.com/v1.0/me/drive/root" + ":/" + folder
	} else {
		url = os.Getenv("GET_FOLDER_URL") + ":/" + folder
	}

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

func GetOneDriveItem(token *oauth2.Token, item string) (resp *http.Response, err error) {

	var url string
	if model.Local {
		url = "https://graph.microsoft.com/v1.0/me/drive/root:/SEAL/" + item + ":/content"
	} else {
		url = os.Getenv("GET_ITEM_URL") + ":/" + folderName + "/" + item + ":/content"
	}
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

// Auxiliary method: returns ID of a given folder (from previous http response)
func GetOneDriveFolderID(resp *http.Response) (id string, err error) {
	var folderprops FolderProps
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	json.Unmarshal([]byte(body), &folderprops)
	id = folderprops.ID
	return
}

// Returns Oauth Token used for authorization of OneDrive requests
func GetOneDriveToken(creds *OneDriveCreds) (redirect *model.Redirect, token *oauth2.Token, err error) {

	log.Println("entrando")
	currentOneDriveToken := oauth2.Token{
		AccessToken:  creds.OneDriveAccessToken,
		RefreshToken: creds.OneDriveRefreshToken,
		TokenType:    "Bearer",
		Expiry:       time.Now().Local().Add(time.Second * time.Duration(3600)),
	}

	log.Println("yes")
	if currentOneDriveToken.AccessToken == "" {

		log.Println("the spot")
		var url string
		url, err = getCodeFromWeb(creds.OneDriveClientID, creds.OneDriveScopes)
		if err != nil {
			return
		}

		desc := `Go to the following link ` + url + `"and login to your Account"`

		redirect = &model.Redirect{
			Description: desc,
			Link:        url,
			Module:      "oneDrive",
		}
		return
	}

	now := time.Now()
	end := currentOneDriveToken.Expiry

	//if the access token hasn't expired yet
	if end.Sub(now) > 10 {
		token = &currentOneDriveToken
		return
	}

	token, err = requestRefreshToken(creds.OneDriveClientID, &currentOneDriveToken)
	//if the access token has expired. Makes a refresh token request
	return
}

// Requests a token from the web, then returns the retrieved token.
// Makes a GET request to retrive a Code. The user needs to copy paste the code on the console
// Afterwards, makes a POST request to retrive the new access_token, given necessary parameters
// In order to use the One Drive API, the client needs the clientID, the redirect_uri and the scopes of the application in the Microsfot Graph
// For more information, follow this link: https://docs.microsoft.com/en-us/onedrive/developer/rest-api/getting-started/graph-oauth?view=odsp-graph-online
func getCodeFromWeb(clientID string, scopes string) (code string, err error) {

	var u *url.URL
	//Retrieve the code
	if model.Local {
		u, err = url.ParseRequestURI("https://login.live.com/oauth20_authorize.srf")
	} else {
		u, err = url.ParseRequestURI(os.Getenv("AUTH_URL"))
	}
	if err != nil {
		return
	}
	urlStr := u.String()

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return
	}

	q := req.URL.Query()
	q.Add("client_id", clientID)
	q.Add("scope", scopes)
	if model.Local {
		q.Add("redirect_uri", "https://localhost:8082/per/code")
	} else {
		q.Add("redirect_uri", os.Getenv("REDIRECT_URL_HTTPS"))
	}
	q.Add("response_type", "code")
	req.URL.RawQuery = q.Encode()

	code = req.URL.String()

	return code, nil
}
func RequestToken(code string, clientID string) (token *oauth2.Token, err error) {

	//Retrieve the access token
	values := url.Values{}
	values.Add("client_id", clientID)
	values.Add("code", code)
	values.Add("grant_type", "authorization_code")
	var u *url.URL
	if model.Local {
		values.Add("redirect_uri", "https://localhost:8082/per/code")
		u, err = url.ParseRequestURI("https://login.microsoftonline.com/common/oauth2/v2.0/token")
	} else {
		values.Add("redirect_uri", os.Getenv("REDIRECT_URL_HTTPS"))
		u, err = url.ParseRequestURI(os.Getenv("FETCH_TOKEN_URL"))
	}
	if err != nil {
		return
	}

	urlStr := u.String()
	req, err := http.NewRequest("POST", urlStr, strings.NewReader(values.Encode()))
	if err != nil {
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	log.Println(req.Body)
	token, err = tokenRequest(req)
	return
}

// POST request to retrieve a new access and refresh tokens
func requestRefreshToken(clientID string, token *oauth2.Token) (tokne *oauth2.Token, err error) {
	values := url.Values{}
	values.Add("client_id", clientID)
	values.Add("refresh_token", token.RefreshToken)
	values.Add("grant_type", "refresh_token")

	var u *url.URL
	if model.Local {
		values.Add("redirect_uri", "https://localhost:8082/per/code")
		u, err = url.ParseRequestURI("https://login.microsoftonline.com/common/oauth2/v2.0/token")
	} else {
		values.Add("redirect_uri", os.Getenv("REDIRECT_URL_HTTPS"))
		u, err = url.ParseRequestURI(os.Getenv("FETCH_TOKEN_URL"))
	}
	if err != nil {
		return
	}

	urlStr := u.String()
	req, err := http.NewRequest("POST", urlStr, strings.NewReader(values.Encode()))
	if err != nil {
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	token, err = tokenRequest(req)
	return
}

//Auxiliary method: performs a token-related http request
func tokenRequest(req *http.Request) (tok *oauth2.Token, err error) {

	client := &http.Client{}
	var respo TokenRequestResponse

	resp, err := client.Do(req)

	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal([]byte(body), &respo)
	if err != nil {
		return
	}

	tok = &oauth2.Token{
		AccessToken:  respo.AccessToken,
		RefreshToken: respo.RefreshToken,
		TokenType:    respo.TokenType,
		Expiry:       time.Now().Local().Add(time.Second * time.Duration(respo.ExpiresIn)),
	}

	currentOneDriveToken = *tok

	return
}

func SetOneDriveCreds(data interface{}) (creds *OneDriveCreds, err error) {
	smr := &sm.SessionMngrResponse{}
	jsonM, err := json.Marshal(data)
	if err != nil {
		return
	}
	creds = &OneDriveCreds{}

	json.Unmarshal(jsonM, smr)
	creds.OneDriveClientID = smr.SessionData.SessionVariables["OneDriveClientID"]
	creds.OneDriveScopes = smr.SessionData.SessionVariables["OneDriveScopes"]
	creds.OneDriveAccessToken = smr.SessionData.SessionVariables["OneDriveAccessToken"]
	creds.OneDriveRefreshToken = smr.SessionData.SessionVariables["OneDriveRefreshToken"]
	return
}

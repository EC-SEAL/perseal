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
	"golang.org/x/oauth2"
)

const (
	folderName string = "SEAL"
	folderId   string = "5C07F9D77D4396CC!106"
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

// FolderProps - The properties of the One Drive folder
type FolderProps struct {
	ID string `json:"id"`
}

type FolderChildren struct {
	Values []struct {
		Name string `json:"name"`
	} `json:"value"`
}

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
func CreateOneDriveFile(token *oauth2.Token, folderID string, filename string, blob []byte) (err error) {
	var url string
	if model.Local {
		url = "https://graph.microsoft.com/v1.0/me/drive/items/" + folderID + ":/" + filename + ":/content"
	} else {
		url = os.Getenv("CREATE_FILE_URL") + folderID + ":/" + filename + ":/content"
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

func GetOneDriveItems(token *oauth2.Token, folder string) (folderchildren *FolderChildren, err error) {
	var url string
	if model.Local {
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

func GetOneDriveItem(token *oauth2.Token, item string) (resp *http.Response, err error) {

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
func CheckOneDriveTokenExpiry(token oauth2.Token, creds *model.OneDriveCreds) (rtoken *oauth2.Token, err error) {

	now := time.Now()
	end := token.Expiry

	//if the access token hasn't expired yet
	if end.Sub(now) > 10 {
		rtoken = &token
		return
	}

	rtoken, err = requestRefreshToken(creds.OneDriveClientID, &token)
	//if the access token has expired. Makes a refresh token request
	return
}

// Requests a token from the web, then returns the retrieved token.
// Makes a GET request to retrive a Code. The user needs to copy paste the code on the console
// Afterwards, makes a POST request to retrive the new access_token, given necessary parameters
// In order to use the One Drive API, the client needs the clientID, the redirect_uri and the scopes of the application in the Microsfot Graph
// For more information, follow this link: https://docs.microsoft.com/en-us/onedrive/developer/rest-api/getting-started/graph-oauth?view=odsp-graph-online
func GetOneDriveRedirectURL(id string, creds *model.OneDriveCreds) (link string, err error) {

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
	q.Add("client_id", creds.OneDriveClientID)
	q.Add("scope", creds.OneDriveScopes)
	if model.Local {
		q.Add("redirect_uri", "http://localhost:8082/per/code")
	} else {
		q.Add("redirect_uri", os.Getenv("REDIRECT_URL_HTTPS"))
	}
	q.Add("response_type", "code")
	q.Add("state", id)
	req.URL.RawQuery = q.Encode()

	link = req.URL.String()

	return link, nil
}
func RequestToken(code string, clientID string) (token *oauth2.Token, err error) {

	//Retrieve the access token
	values := url.Values{}
	values.Add("client_id", clientID)
	values.Add("code", code)
	values.Add("grant_type", "authorization_code")
	var u *url.URL
	if model.Local {
		values.Add("redirect_uri", "http://localhost:8082/per/code")
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
		values.Add("redirect_uri", "http://localhost:8082/per/code")
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
	return
}

func SetOneDriveCreds() (creds *model.OneDriveCreds, err error) {
	creds = &model.OneDriveCreds{}

	if model.Local {
		creds.OneDriveClientID = "fff1cba9-7597-479d-b653-fd96c5d56b43"
		creds.OneDriveScopes = "offline_access files.read files.read.all files.readwrite files.readwrite.all"
	} else {
		creds.OneDriveClientID = os.Getenv("ONE_DRIVE_CLIENT_ID")
		creds.OneDriveScopes = os.Getenv("ONE_DRIVE_SCOPES")
	}

	return
}

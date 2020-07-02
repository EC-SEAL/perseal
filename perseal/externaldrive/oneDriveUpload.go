// Package onedrive implements the OAuth2 protocol for authenticating users through onedrive.
// This package can be used as a reference implementation of an OAuth2 provider for Goth.
package externaldrive

import (
	"bytes"
	"encoding/json"
	"errors"
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
	"google.golang.org/api/drive/v3"
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

// FolderProps - The properties of the One Drive folder
type FolderProps struct {
	ID string `json:"id"`
}

type FolderChildren struct {
	Values []struct {
		Name string `json:"name"`
	} `json:"value"`
}

// UploadOneDrive - Uploads file to One Drive
func (ds *DataStore) UploadOneDrive(oauthToken *oauth2.Token, data []byte, filename string, folderName string) (file *drive.File, err error) {

	//if the folder exists, only creats the datastore file
	fileExists, err := GetOneDriveFolder(oauthToken, folderName)
	log.Println(fileExists)
	if err != nil {
		return
	}
	if fileExists.StatusCode == 401 {
		err = errors.New("Unauthorized Request")
		return
	}

	var folderID string
	if fileExists.StatusCode == 404 {
		folderID, err = CreateOneDriveFolder(oauthToken)
		if err != nil {
			return
		}
		err = CreateOneDriveFile(oauthToken, folderID, filename, data)
		if err != nil {
			return
		}
	} else {
		folderID, err = getOneDriveFolderID(fileExists)
		if err != nil {
			return
		}
		err = CreateOneDriveFile(oauthToken, folderID, filename, data)
	}
	return
}

// POST request to create a folder in the root
func CreateOneDriveFolder(token *oauth2.Token) (folderID string, err error) {
	createfolderjson := []byte(`{"name":"` + folderName + `","folder": {},"@microsoft.graph.conflictBehavior": "rename"}`)
	var req *http.Request
	if model.Test {
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

	folderID, err = getOneDriveFolderID(resp)
	return
}

// PUT request to create a file in a given folder
func CreateOneDriveFile(token *oauth2.Token, folderID string, filename string, blob []byte) (err error) {
	var url string
	if model.Test {
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
	if model.Test {
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

// Auxiliary method: returns ID of a given folder (from previous http response)
func getOneDriveFolderID(resp *http.Response) (id string, err error) {
	var folderprops FolderProps
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	json.Unmarshal([]byte(body), &folderprops)
	id = folderprops.ID
	return
}

func RequestToken(code string, clientID string) (token *oauth2.Token, err error) {

	//Retrieve the access token
	values := url.Values{}
	values.Add("client_id", clientID)
	values.Add("code", code)
	values.Add("grant_type", "authorization_code")
	var u *url.URL
	if model.Test {
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
func RequestRefreshToken(clientID string, token *oauth2.Token) (tokne *oauth2.Token, err error) {
	values := url.Values{}
	values.Add("client_id", clientID)
	values.Add("refresh_token", token.RefreshToken)
	values.Add("grant_type", "refresh_token")

	var u *url.URL
	if model.Test {
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

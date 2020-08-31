package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/externaldrive"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
	"golang.org/x/oauth2"
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
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedEncryption, erro.Error())
		return
	}

	contents, _ := dataStore.UploadingBlob()

	erro = uploadOneDrive(dataStore, token, contents)

	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedDataStoreStoringInFile, erro.Error())
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
	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedExecuteRequest, erro.Error())
		return
	}
	if file.StatusCode != 200 {
		err = model.BuildResponse(http.StatusNotFound, model.Messages.FailedGetCloudFile+model.EnvVariables.One_Drive_PDS)
		return
	}
	return
}

func updateNewOneDriveTokenFromCode(id string, code string) (oauthToken *oauth2.Token, err *model.HTMLResponse) {
	creds := model.EnvVariables.OneDriveCreds

	var erro error
	oauthToken, erro = requestToken(code, creds.OneDriveClientID)
	if erro != nil {
		err = model.BuildResponse(http.StatusNotFound, model.Messages.FailedGetToken+model.EnvVariables.One_Drive_PDS, erro.Error())
		return
	}

	jsonM, erro := json.Marshal(oauthToken)
	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedParseToken+model.EnvVariables.One_Drive_PDS, erro.Error())
		return
	}

	err = sm.UpdateSessionData(id, string(jsonM), model.EnvVariables.SessionVariables.OneDriveToken)
	return
}

func getOneDriveItems(token *oauth2.Token) (folderchildren *FolderChildren, err error) {

	id, err := getOneDriveId(token)

	url := model.EnvVariables.OneDriveURLs.Get_Items + id.ID + "/children"

	req, err := http.NewRequest(http.MethodGet, url, nil)
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

	json.Unmarshal(jsonM, &folderchildren)

	return
}

func getOneDriveItem(token *oauth2.Token, item string) (resp *http.Response, err error) {

	var url string
	url = model.EnvVariables.OneDriveURLs.Get_Item + model.EnvVariables.DataStore_Folder_Name + "/" + item + "/:/content"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}
	auth := "Bearer " + token.AccessToken
	req.Header.Add("Authorization", auth)

	client := &http.Client{}
	resp, err = client.Do(req)

	return
}

func getOneDriveId(token *oauth2.Token) (id *FolderProps, err error) {

	var url string
	url = model.EnvVariables.OneDriveURLs.Get_Item + model.EnvVariables.DataStore_Folder_Name
	req, err := http.NewRequest(http.MethodGet, url, nil)
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

	rtoken, erro := requestRefreshToken(creds.OneDriveClientID, &token)

	if erro != nil {
		err = model.BuildResponse(http.StatusNotFound, model.Messages.FailedRefreshToken, erro.Error())
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
	req, _ := http.NewRequest(http.MethodGet, model.EnvVariables.OneDriveURLs.Auth, nil)

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

// One Drive Upload Methods

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
func uploadOneDrive(dataStore *externaldrive.DataStore, oauthToken *oauth2.Token, data []byte) (err error) {
	//if the folder exists, only creats the datastore file
	fileExists, err := getOneDriveFolder(oauthToken)
	log.Println(fileExists)
	if err != nil {
		return
	}
	if fileExists.StatusCode == 401 {
		err = errors.New(model.Messages.UnauthorizedRequest)
		return
	}

	var folderID string
	if fileExists.StatusCode == 404 {
		folderID, err = createOneDriveFolder(oauthToken)
		if err != nil {
			return
		}
		err = createOneDriveFile(oauthToken, folderID, data)
		if err != nil {
			return
		}
	} else {
		folderID, err = getOneDriveFolderID(fileExists)
		if err != nil {
			return
		}
		err = createOneDriveFile(oauthToken, folderID, data)
	}
	return
}

// POST request to create a folder in the root
func createOneDriveFolder(token *oauth2.Token) (folderID string, err error) {
	createfolderjson := []byte(`{"name":"` + model.EnvVariables.DataStore_Folder_Name + `","folder": {},"@microsoft.graph.conflictBehavior": "rename"}`)
	req, err := http.NewRequest(http.MethodPost, model.EnvVariables.OneDriveURLs.Create_Folder, bytes.NewBuffer(createfolderjson))

	if err != nil {
		return
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := oneDriveStorageOperationsRequest(req, *token)
	folderID, err = getOneDriveFolderID(resp)
	return
}

// PUT request to create a file in a given folder
func createOneDriveFile(token *oauth2.Token, folderID string, blob []byte) (err error) {
	url := model.EnvVariables.OneDriveURLs.Create_File + folderID + ":/" + model.EnvVariables.DataStore_File_Name + ":/content"
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(blob))
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")
	_, err = oneDriveStorageOperationsRequest(req, *token)

	return
}

// GET request to fetch information of a given folder
func getOneDriveFolder(token *oauth2.Token) (resp *http.Response, err error) {

	url := model.EnvVariables.OneDriveURLs.Get_Folder + ":/" + model.EnvVariables.DataStore_Folder_Name
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}
	return oneDriveStorageOperationsRequest(req, *token)
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

func requestToken(code string, clientID string) (token *oauth2.Token, err error) {

	//Retrieve the access token
	values := url.Values{}
	values.Add("client_id", clientID)
	values.Add("code", code)
	values.Add("grant_type", "authorization_code")
	values.Add("redirect_uri", model.EnvVariables.Redirect_URL)

	u, err := url.ParseRequestURI(model.EnvVariables.OneDriveURLs.Fetch_Token)
	if err != nil {
		return
	}

	urlStr := u.String()
	req, err := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(values.Encode()))
	if err != nil {
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	token, err = tokenRequest(req)
	return
}

// POST request to retrieve a new access and refresh tokens
func requestRefreshToken(clientID string, token *oauth2.Token) (tokne *oauth2.Token, err error) {
	values := url.Values{}
	values.Add("client_id", clientID)
	values.Add("refresh_token", token.RefreshToken)
	values.Add("grant_type", "refresh_token")
	values.Add("redirect_uri", model.EnvVariables.Redirect_URL)

	u, err := url.ParseRequestURI(model.EnvVariables.OneDriveURLs.Fetch_Token)
	if err != nil {
		return
	}

	urlStr := u.String()
	req, err := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(values.Encode()))
	if err != nil {
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	token, err = tokenRequest(req)
	return
}

func oneDriveStorageOperationsRequest(req *http.Request, token oauth2.Token) (resp *http.Response, err error) {
	auth := "Bearer " + token.AccessToken
	req.Header.Add("Authorization", auth)

	client := &http.Client{}
	return client.Do(req)
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

package model

import (
	"os"
	"strings"
)

type HTMLResponse struct {
	Code               int    `json:"code"`
	Message            string `json:"message"`
	ErrorMessage       string `json:"error"`
	ClientCallbackAddr string
	DataStore          string
}

type TokenResponse struct {
	Payload string `json:"payload"`
	Status  struct {
		Message string `json:"message"`
	}
}

type GoogleDriveCreds struct {
	Web struct {
		ClientId                string   `json:"client_id"`
		ProjectId               string   `json:"project_id"`
		AuthURI                 string   `json:"auth_uri"`
		TokenURI                string   `json:"token_uri"`
		AuthProviderx509CertUrl string   `json:"auth_provider_x509_cert_url"`
		ClientSecret            string   `json:"client_secret"`
		RedirectURIS            []string `json:"redirect_uris"`
	} `json:"web"`
}

type OneDriveCreds struct {
	OneDriveClientID     string `json:"oneDriveClient"`
	OneDriveScopes       string `json:"oneDriveScopes"`
	OneDriveAccessToken  string `json:"oneDrivetAccessToken"`
	OneDriveRefreshToken string `json:"oneDrivetRefreshToken"`
}

var Test = false

var EnvVariables struct {
	Store_Method      string
	Load_Method       string
	Store_Load_Method string

	Google_Drive_PDS string
	One_Drive_PDS    string
	Mobile_PDS       string
	Browser_PDS      string

	DataStore_Folder_Name string
	DataStore_Folder_ID   string
	DataStore_File_Name   string

	Redirect_URL string
	Host         string

	Project_SEAL_Email string

	GoogleDriveCreds GoogleDriveCreds
	OneDriveCreds    OneDriveCreds

	OneDriveURLs struct {
		Auth          string
		Create_Folder string
		Create_File   string
		Get_Folder    string
		Fetch_Token   string
		Get_Items     string
		Get_Item      string
	}

	SMURLs struct {
		EndPoint            string
		Validate_Token      string
		Get_Session_Data    string
		Update_Session_Data string
	}
}

func TestEnvVariables() {
	EnvVariables.Google_Drive_PDS = "googleDrive"
	EnvVariables.One_Drive_PDS = "oneDrive"
	EnvVariables.Browser_PDS = "Browser"
	EnvVariables.Mobile_PDS = "Mobile"

	EnvVariables.Store_Method = "store"
	EnvVariables.Load_Method = "load"
	EnvVariables.Store_Load_Method = "storeload"

	EnvVariables.DataStore_Folder_Name = "SEAL"
	EnvVariables.DataStore_File_Name = "datastore.seal"

	EnvVariables.Redirect_URL = "http://localhost:8082/per/code"
	EnvVariables.Host = "localhost:8082"

	EnvVariables.Project_SEAL_Email = "info@project-seal.eu"

	EnvVariables.OneDriveURLs.Auth = "https://login.live.com/oauth20_authorize.srf"
	EnvVariables.OneDriveURLs.Create_Folder = "https://graph.microsoft.com/v1.0/me/drive/root/children"
	EnvVariables.OneDriveURLs.Create_File = "https://graph.microsoft.com/v1.0/me/drive/items/"
	EnvVariables.OneDriveURLs.Get_Folder = "https://graph.microsoft.com/v1.0/me/drive/root"
	EnvVariables.OneDriveURLs.Fetch_Token = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
	EnvVariables.OneDriveURLs.Get_Items = "https://graph.microsoft.com/v1.0/me/drive/items/"
	EnvVariables.OneDriveURLs.Get_Item = "https://graph.microsoft.com/v1.0/me/drive/root:/SEAL/"

	EnvVariables.GoogleDriveCreds.Web.ClientId = "425112724933-9o8u2rk49pfurq9qo49903lukp53tbi5.apps.googleusercontent.com"
	EnvVariables.GoogleDriveCreds.Web.ProjectId = "seal-274215"
	EnvVariables.GoogleDriveCreds.Web.AuthURI = "https://accounts.google.com/o/oauth2/auth"
	EnvVariables.GoogleDriveCreds.Web.TokenURI = "https://oauth2.googleapis.com/token"
	EnvVariables.GoogleDriveCreds.Web.AuthProviderx509CertUrl = "https://www.googleapis.com/oauth2/v1/certs"
	EnvVariables.GoogleDriveCreds.Web.ClientSecret = "0b3WtqfasYfWDmk31xa8UAht"
	EnvVariables.GoogleDriveCreds.Web.RedirectURIS = []string{"http://localhost:8082/per/code"}

	EnvVariables.OneDriveCreds.OneDriveClientID = "fff1cba9-7597-479d-b653-fd96c5d56b43"
	EnvVariables.OneDriveCreds.OneDriveScopes = "offline_access files.read files.read.all files.readwrite files.readwrite.all"

	EnvVariables.SMURLs.EndPoint = "http://vm.project-seal.eu:9090/sm"
	EnvVariables.SMURLs.Validate_Token = "/validateToken?token="
	EnvVariables.SMURLs.Get_Session_Data = "/getSessionData?sessionId="
	EnvVariables.SMURLs.Update_Session_Data = "/updateSessionData"
}

func ProductionEnvVariables() {

	EnvVariables.Google_Drive_PDS = os.Getenv("GOOGLE_DRIVE_PDS")
	EnvVariables.One_Drive_PDS = os.Getenv("ONE_DRIVE_PDS")
	EnvVariables.Browser_PDS = os.Getenv("BROWSER_PDS")
	EnvVariables.Mobile_PDS = os.Getenv("MOBILE_PDS")

	EnvVariables.Store_Method = os.Getenv("STORE_METHOD")
	EnvVariables.Load_Method = os.Getenv("LOAD_METHOD")
	EnvVariables.Store_Load_Method = os.Getenv("STORE_LOAD_METHOD")

	EnvVariables.DataStore_Folder_Name = os.Getenv("DATASTORE_FOLDER_NAME")
	EnvVariables.DataStore_File_Name = os.Getenv("DATASTORE_FILE_NAME")

	EnvVariables.Redirect_URL = os.Getenv("REDIRECT_URL")
	EnvVariables.Host = os.Getenv("HOST")

	EnvVariables.Project_SEAL_Email = os.Getenv("PROJECT_SEAL_EMAIL")

	EnvVariables.OneDriveURLs.Auth = os.Getenv("AUTH_URL")
	EnvVariables.OneDriveURLs.Create_Folder = os.Getenv("CREATE_FOLDER_URL")
	EnvVariables.OneDriveURLs.Create_File = os.Getenv("CREATE_FILE_URL")
	EnvVariables.OneDriveURLs.Get_Folder = os.Getenv("GET_FOLDER_URL")
	EnvVariables.OneDriveURLs.Fetch_Token = os.Getenv("FETCH_TOKEN_URL")
	EnvVariables.OneDriveURLs.Get_Items = os.Getenv("GET_ITEMS_URL")

	EnvVariables.GoogleDriveCreds.Web.ClientId = os.Getenv("GOOGLE_DRIVE_CLIENT_ID")
	EnvVariables.GoogleDriveCreds.Web.ProjectId = os.Getenv("GOOGLE_DRIVE_CLIENT_PROJECT")
	EnvVariables.GoogleDriveCreds.Web.AuthURI = os.Getenv("GOOGLE_DRIVE_AUTH_URI")
	EnvVariables.GoogleDriveCreds.Web.TokenURI = os.Getenv("GOOGLE_DRIVE_TOKEN_URI")
	EnvVariables.GoogleDriveCreds.Web.AuthProviderx509CertUrl = os.Getenv("GOOGLE_DRIVE_AUTH_PROVIDER")
	EnvVariables.GoogleDriveCreds.Web.ClientSecret = os.Getenv("GOOGLE_DRIVE_CLIENT_SECRET")
	EnvVariables.GoogleDriveCreds.Web.RedirectURIS = strings.Split([]string{os.Getenv("GOOGLE_DRIVE_REDIRECT_URIS")}[0], ",")

	EnvVariables.OneDriveCreds.OneDriveClientID = os.Getenv("ONE_DRIVE_CLIENT_ID")
	EnvVariables.OneDriveCreds.OneDriveScopes = os.Getenv("ONE_DRIVE_SCOPES")

	EnvVariables.SMURLs.EndPoint = os.Getenv("SM_ENDPOINT")
	EnvVariables.SMURLs.Validate_Token = os.Getenv("VALIDATE_TOKEN")
	EnvVariables.SMURLs.Get_Session_Data = os.Getenv("GET_SESSION_DATA")
	EnvVariables.SMURLs.Update_Session_Data = os.Getenv("UPDATE_SESSION_DATA")
}

var TestUser string
var MSToken string

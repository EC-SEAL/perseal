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
	MSToken            string
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

//Review Env Variables Before Deploy
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
	DataStore_File_Name   string

	Redirect_URL         string
	Dashboard_Custom_URL string
	Host                 string

	Project_SEAL_Email string

	GoogleDriveCreds GoogleDriveCreds
	OneDriveCreds    OneDriveCreds

	SessionVariables struct {
		CurrentMethod      string
		ClientCallbackAddr string
		PDS                string
		OneDriveToken      string
		GoogleDriveToken   string
		DataStore          string
		UserDevice         string
		MockUserData       string
		SessionId          string
	}

	TestURLs struct {
		MockRedirectDashboard string
	}

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
		New_Add             string
		New_Search          string
		New_Delete          string
		APIGW_Endpoint      string
	}
}

func SetEnvVariables() {

	EnvVariables.Google_Drive_PDS = os.Getenv("GOOGLE_DRIVE_PDS")
	EnvVariables.One_Drive_PDS = os.Getenv("ONE_DRIVE_PDS")
	EnvVariables.Browser_PDS = os.Getenv("BROWSER_PDS")
	EnvVariables.Mobile_PDS = os.Getenv("MOBILE_PDS")

	EnvVariables.Store_Method = os.Getenv("STORE_METHOD")
	EnvVariables.Load_Method = os.Getenv("LOAD_METHOD")
	EnvVariables.Store_Load_Method = os.Getenv("STORE_LOAD_METHOD")

	EnvVariables.DataStore_Folder_Name = os.Getenv("DATASTORE_FOLDER_NAME")
	EnvVariables.Dashboard_Custom_URL = os.Getenv("DASHBOARD_CUSTOM_URL")
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
	EnvVariables.OneDriveURLs.Get_Item = os.Getenv("GET_ITEM_URL")

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
	EnvVariables.SMURLs.New_Add = os.Getenv("NEW_ADD")
	EnvVariables.SMURLs.New_Search = os.Getenv("NEW_SEARCH")
	EnvVariables.SMURLs.New_Delete = os.Getenv("NEW_DELETE")
	EnvVariables.SMURLs.Validate_Token = os.Getenv("VALIDATE_TOKEN")
	EnvVariables.SMURLs.Get_Session_Data = os.Getenv("GET_SESSION_DATA")
	EnvVariables.SMURLs.Update_Session_Data = os.Getenv("UPDATE_SESSION_DATA")
	EnvVariables.SMURLs.APIGW_Endpoint = os.Getenv("APIGW_ENDPOINT")

	EnvVariables.TestURLs.MockRedirectDashboard = os.Getenv("MOCK_REDIRECT_DASHBOARD")

	EnvVariables.SessionVariables.CurrentMethod = os.Getenv("CURRENT_METHOD_VARIABLE")
	EnvVariables.SessionVariables.ClientCallbackAddr = os.Getenv("CLIENT_CALLBACK_ADDR_VARIABLE")
	EnvVariables.SessionVariables.MockUserData = os.Getenv("MOCK_USER_DATA_VARIABLE")
	EnvVariables.SessionVariables.PDS = os.Getenv("PDS_VARIABLE")
	EnvVariables.SessionVariables.OneDriveToken = os.Getenv("ONE_DRIVE_TOKEN_VARIABLE")
	EnvVariables.SessionVariables.GoogleDriveToken = os.Getenv("GOOGLE_DRIVE_TOKEN_VARIABLE")
	EnvVariables.SessionVariables.DataStore = os.Getenv("DATASTORE_VARIABLE")
	EnvVariables.SessionVariables.UserDevice = os.Getenv("USER_DEVICE_VARIABLE")
	EnvVariables.SessionVariables.SessionId = os.Getenv("SESSION_ID")
}

var TestUser string
var MSToken string

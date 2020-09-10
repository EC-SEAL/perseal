package model

import (
	"encoding/json"
	"os"
	"strings"
)

type HTMLResponse struct {
	Code               int    `json:"code"`
	Message            string `json:"message"`
	ErrorMessage       string `json:"error"`
	DataStore          string
	MSToken            string
	MSTokenRedirect    string
	MSTokenDownload    string
	FailedInput        string
	ClientCallbackAddr string
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
//IMPORTANT: urls in qr.html
var Test = true

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

	Redirect_URL            string
	CustomURL               string
	Host                    string
	Perseal_Sender_Receiver string
	APGW_ID                 string
	RM_ID                   string

	Project_SEAL_Email string

	GoogleDriveCreds GoogleDriveCreds
	OneDriveCreds    OneDriveCreds

	SessionVariables struct {
		FinishedPersealBackChannel string
		CurrentMethod              string
		ClientCallbackAddr         string
		PDS                        string
		OneDriveToken              string
		GoogleDriveToken           string
		DataStore                  string
		MockUserData               string
		SessionId                  string
	}

	TestURLs struct {
		APIGW_Endpoint string
	}
	Perseal_RM_UCs_Callback string
	RM_Endpoint             string
	Perseal_QRCode_Endpoint string

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
		Generate_Token      string
		Validate_Token      string
		Get_Session_Data    string
		Update_Session_Data string
		New_Add             string
		New_Search          string
		New_Delete          string
	}
}

var Messages struct {
	NoPassword           string
	FileContentsNotFound string
	FailedOpenFile       string
	FailedReadFile       string
	NoMSTokenErrorMsg    string

	FailedFoundDataStore string

	FailedMarshall string

	FailedGenerateURL    string
	FailedSignRequest    string
	FailedExecuteRequest string
	FailedReadResponse   string
	FailedParseResponse  string
	FailedGenerateSMResp string
	ISE                  string
	SMRespError          string

	StoredDataStore  string
	FailedEncryption string

	LoadedDataStore string

	InvalidSignature      string
	InvalidPassword       string
	InvalidDataStore      string
	FileContainsDataStore string
	FailedGetCloudFiles   string

	UnauthorizedRequest          string
	FailedDataStoreStoringInFile string
	FailedGetCloudFile           string
	FailedGetToken               string
	FailedParseToken             string
	FailedRefreshToken           string

	IncompleteQRCode string
	PrintedQRCode    string
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
	EnvVariables.DataStore_File_Name = os.Getenv("DATASTORE_FILE_NAME")

	EnvVariables.CustomURL = os.Getenv("CUSTOM_URL")
	EnvVariables.Redirect_URL = os.Getenv("REDIRECT_URL")
	EnvVariables.Host = os.Getenv("HOST")
	EnvVariables.Perseal_Sender_Receiver = os.Getenv("PERSEAL_SENDER_RECEIVER")

	EnvVariables.APGW_ID = os.Getenv("APGW_ID")
	EnvVariables.RM_ID = os.Getenv("RM_ID")

	EnvVariables.Project_SEAL_Email = os.Getenv("PROJECT_SEAL_EMAIL")
	EnvVariables.Perseal_RM_UCs_Callback = os.Getenv("PERSEAL_RM_UCs_CALLBACK")
	EnvVariables.RM_Endpoint = os.Getenv("RM_ENDPOINT")
	EnvVariables.Perseal_QRCode_Endpoint = os.Getenv("PERSEAL_QR_CODE_ENDPOINT")

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
	EnvVariables.SMURLs.Generate_Token = os.Getenv("GENERATE_TOKEN")
	EnvVariables.SMURLs.New_Add = os.Getenv("NEW_ADD")
	EnvVariables.SMURLs.New_Search = os.Getenv("NEW_SEARCH")
	EnvVariables.SMURLs.New_Delete = os.Getenv("NEW_DELETE")
	EnvVariables.SMURLs.Validate_Token = os.Getenv("VALIDATE_TOKEN")
	EnvVariables.SMURLs.Get_Session_Data = os.Getenv("GET_SESSION_DATA")
	EnvVariables.SMURLs.Update_Session_Data = os.Getenv("UPDATE_SESSION_DATA")

	EnvVariables.TestURLs.APIGW_Endpoint = "https://vm.project-seal.eu:9154"

	EnvVariables.SessionVariables.FinishedPersealBackChannel = os.Getenv("FINISHED_PERSEAL_BACK_CHANNEL")
	EnvVariables.SessionVariables.CurrentMethod = os.Getenv("CURRENT_METHOD_VARIABLE")
	EnvVariables.SessionVariables.ClientCallbackAddr = os.Getenv("CLIENT_CALLBACK_ADDR_VARIABLE")
	EnvVariables.SessionVariables.MockUserData = os.Getenv("MOCK_USER_DATA_VARIABLE")
	EnvVariables.SessionVariables.PDS = os.Getenv("PDS_VARIABLE")
	EnvVariables.SessionVariables.OneDriveToken = os.Getenv("ONE_DRIVE_TOKEN_VARIABLE")
	EnvVariables.SessionVariables.GoogleDriveToken = os.Getenv("GOOGLE_DRIVE_TOKEN_VARIABLE")
	EnvVariables.SessionVariables.DataStore = os.Getenv("DATASTORE_VARIABLE")
	EnvVariables.SessionVariables.SessionId = os.Getenv("SESSION_ID")

	Messages.NoPassword = "No Password Found"
	Messages.FileContentsNotFound = "Could not find file contents"
	Messages.FailedOpenFile = "Could not open file"
	Messages.FailedReadFile = "Could not read file contents"
	Messages.NoMSTokenErrorMsg = "JWT is blacklisted"
	Messages.FailedFoundDataStore = "Couldn't find DataStore"
	Messages.FailedMarshall = "Could Not Marshall "
	Messages.FailedGenerateURL = "Couldn't Generate URL to "
	Messages.FailedSignRequest = "Could Not Sign Request"
	Messages.FailedExecuteRequest = "Couldn't Execute Request"
	Messages.FailedReadResponse = "Couldn't Read Response from "
	Messages.FailedParseResponse = "Couldn't Parse Response Body to Object - "
	Messages.FailedGenerateSMResp = "Couldn't Generate SessionManagerResponse - "
	Messages.ISE = "Internal Server Error"
	Messages.SMRespError = `ERROR" code in the received SessionData`
	Messages.FailedEncryption = "Encryption Failed"
	Messages.InvalidSignature = "Invalid Signature"
	Messages.InvalidPassword = "Couldn't Decrypt DataStore. Check your password"
	Messages.InvalidDataStore = "Bad Structure of DataStore"
	Messages.FileContainsDataStore = "The File must contain only a valid DataStore"
	Messages.FailedGetCloudFiles = "Could not Get Cloud Files - "
	Messages.FailedGetCloudFiles = "Could not Get Cloud File - "
	Messages.UnauthorizedRequest = "Unauthorized Request"
	Messages.FailedDataStoreStoringInFile = "Couldn't Store DataStore in File"
	Messages.FailedGetToken = "Could not Fetch Access Token - "
	Messages.FailedParseToken = "Couldn't Parse the Access Token to byte array - "
	Messages.FailedRefreshToken = "Error in Request to Refresh Token"
	Messages.IncompleteQRCode = "Incomplete QR Code"

	Messages.StoredDataStore = "Stored DataStore"
	Messages.LoadedDataStore = "Loaded DataStore "
	Messages.PrintedQRCode = "Printed Qr Code"

}

var TestUser string
var MSToken string

func BuildResponse(code int, message string, erro ...string) *HTMLResponse {
	resp := &HTMLResponse{
		Code:    code,
		Message: message,
	}
	if len(erro) > 0 || erro != nil {
		resp.ErrorMessage = erro[0]
	}
	return resp
}

func MarshallResponseToPrint(response HTMLResponse) string {

	type HTMLResponseToPrint struct {
		Code               int    `json:"code"`
		Message            string `json:"message"`
		ErrorMessage       string `json:"error"`
		ClientCallbackAddr string
	}

	resBytes, _ := json.MarshalIndent(&HTMLResponseToPrint{
		Code:               response.Code,
		ClientCallbackAddr: response.ClientCallbackAddr,
		Message:            response.Message,
		ErrorMessage:       response.ErrorMessage,
	}, "", "\t")
	return string(resBytes)
}

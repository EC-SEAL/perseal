package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/EC-SEAL/perseal/sm"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
)

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

var AccessCreds string
var googleCreds *GoogleDriveCreds

// Requests a token from the web, then returns the retrieved token.
func getGoogleLinkForDashboardRedirect(config *oauth2.Config) (string, string) {
	//	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("redirect_uri", os.Getenv("REDIRECT_URL")))
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("redirect_uri", "https://localhost:8082/per/code"))

	desc := `Go to the following link ` + authURL + `"and login to your Account"`

	return desc, authURL
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {

}

func getService(client *http.Client) (*drive.Service, error) {
	service, err := drive.New(client)
	if err != nil {
		fmt.Printf("Cannot create the Google Drive service: %v\n", err)
		return nil, err
	}

	return service, err
}

func isFolder(file *drive.File) bool {
	return file.MimeType == "application/vnd.google-apps.folder"
}

func createGoogleDriveDir(service *drive.Service, name string, parentId string) (*drive.File, error) {
	files, err := service.Files.List().Do()
	if err != nil {
		return nil, err
	}
	for _, f := range files.Files {
		// service.Files.Delete(f.Id).Do()
		if f.Name == name && isFolder(f) {
			log.Println(f.Id)
			return f, nil
		}
	}
	d := &drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{parentId},
	}

	file, err := service.Files.Create(d).Do()
	if err != nil {
		log.Println("Could not create dir: " + err.Error())
		return nil, err
	}

	return file, nil
}

func createGoogleDriveFile(service *drive.Service, name string, mimeType string, content io.Reader, parentId string) (*drive.File, error) {
	files, err := service.Files.List().Do()
	if err != nil {
		return nil, errors.New("Could not list")
	}
	log.Println(files)
	// TODO check if already exists and if so - update file
	f := &drive.File{
		MimeType: mimeType,
		Name:     name,
		Parents:  []string{parentId},
	}
	for _, v := range files.Files {
		log.Println(v.Name)
		if v.Name == name && !isFolder(v) {
			service.Files.Delete(v.Id).Do()
			// Update does not seem to be working - so It deletes the file before writing it again instead
			// service.Files.Update(v.Id, f).Do()
			// return v, nil
		}
	}

	file, err := service.Files.Create(f).Media(content).Do()

	if err != nil {
		log.Println("Could not create file: " + err.Error())
		return nil, err
	}

	return file, nil
}

type FileProps struct {
	Id          string
	Name        string
	Path        string
	Blob        []byte
	Md5sum      string
	ContentType string
}

func TokenFromSessionData() (outhToken *oauth2.Token, err error) {
	tok := &oauth2.Token{}
	err = json.NewDecoder(strings.NewReader(AccessCreds)).Decode(tok)
	return tok, err
}

//SendFile gdrive file given encrypted blob and oauth token
func SendFile(fileProps *FileProps, client *http.Client) (file *drive.File, err error) {

	service, err := getService(client)
	if err != nil {
		return
	}
	// Creates dir if it doesnt already exist
	dir, err := createGoogleDriveDir(service, fileProps.Path, "root")

	if err != nil {
		// return nil, errors.New("Could not create folder")
		return nil, err
	}
	file, err = createGoogleDriveFile(service, fileProps.Name, fileProps.ContentType, bytes.NewReader(fileProps.Blob), dir.Id)
	if err != nil {
		// return nil, errors.New("Could not create file")
		return nil, err
	}
	// TODO check md5sum of data with CreatedFile
	// md5sum := md5.Sum(fileProps.Blob)
	return file, err
}

func GetGoogleDriveFile(fileName string, oauthToken *http.Client) (file *drive.File, err error) {
	service, err := getService(oauthToken)
	file, err = service.Files.Get(fileName).Do()
	return file, err
}

func setGoogleDriveCreds(data interface{}) {
	googleCreds = &GoogleDriveCreds{}
	jsonM, _ := json.Marshal(data)
	smr := &sm.SessionMngrResponse{}
	json.Unmarshal(jsonM, smr)

	googleCreds.Web.ClientId = smr.SessionData.SessionVariables["GoogleDriveClientID"]
	googleCreds.Web.ProjectId = smr.SessionData.SessionVariables["GoogleDriveProject"]
	googleCreds.Web.AuthURI = smr.SessionData.SessionVariables["GoogleDriveAuthURI"]
	googleCreds.Web.TokenURI = smr.SessionData.SessionVariables["GoogleDriveTokenURI"]
	googleCreds.Web.AuthProviderx509CertUrl = smr.SessionData.SessionVariables["GoogleDriveAuthProviderx509CertUrl"]
	googleCreds.Web.ClientSecret = smr.SessionData.SessionVariables["GoogleDriveClientSecret"]
	googleCreds.Web.RedirectURIS = strings.Split([]string{smr.SessionData.SessionVariables["GoogleDriveRedirectUris"]}[0], ",")
	AccessCreds = smr.SessionData.SessionVariables["GoogleDriveAccessCreds"]
}

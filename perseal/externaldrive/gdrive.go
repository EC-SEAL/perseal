package externaldrive

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/EC-SEAL/perseal/model"
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
func GetGoogleLinkForDashboardRedirect(config *oauth2.Config) string {
	var authURL string
	if model.Local {
		authURL = config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("redirect_uri", "http://localhost:4200/code"))
	} else {
		authURL = config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("redirect_uri", os.Getenv("REDIRECT_URL")))
	}
	return authURL
}

func isFolder(file *drive.File) bool {
	return file.MimeType == "application/vnd.google-apps.folder"
}

func createGoogleDriveDir(service *drive.Service, name string, parentId string) (file *drive.File, err error) {
	files, err := service.Files.List().Do()
	if err != nil {
		return
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

	file, err = service.Files.Create(d).Do()
	return
}

func GetGoogleDriveFile(filename string, client *http.Client) (file *http.Response, err error) {
	service, err := drive.New(client)
	if err != nil {
		return
	}

	list, err := service.Files.List().Do()
	if err != nil {
		return
	}
	var fileId string
	for _, v := range list.Files {
		if v.Name == filename {
			fileId = v.Id
		}
	}
	file, err = service.Files.Get(fileId).Download()
	return
}

func GetGoogleDriveFiles(client *http.Client) (fileList []string, err error) {
	service, err := drive.New(client)
	if err != nil {
		return
	}

	list, err := service.Files.List().Do()
	if err != nil {
		return
	}
	fileList = make([]string, 0)
	for _, v := range list.Files {
		fileList = append(fileList, v.Name)
	}
	return
}

func DeleteGoogleDriveFiles(client *http.Client) (fileList []string, err error) {
	service, err := drive.New(client)
	if err != nil {
		return
	}

	list, err := service.Files.List().Do()
	if err != nil {
		return
	}
	fileList = make([]string, 0)
	for _, v := range list.Files {
		service.Files.Delete(v.Id)
	}
	fileList = fileList[:len(fileList)-1]
	return
}

func createGoogleDriveFile(service *drive.Service, name string, mimeType string, content io.Reader, parentId string) (file *drive.File, err error) {
	files, err := service.Files.List().Do()
	if err != nil {
		return
	}
	log.Println(files)
	// TODO check if already exists and if so - update file
	f := &drive.File{
		MimeType: mimeType,
		Name:     name,
		Parents:  []string{parentId},
	}
	for _, v := range files.Files {
		if v.Name == name && !isFolder(v) {
			service.Files.Delete(v.Id).Do()
			// Update does not seem to be working - so It deletes the file before writing it again instead
			// service.Files.Update(v.Id, f).Do()
			// return v, nil
		}
	}

	file, err = service.Files.Create(f).Media(content).Do()
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

//SendFile gdrive file given encrypted blob and oauth token
func SendFile(fileProps *FileProps, client *http.Client) (file *drive.File, err error) {

	service, err := drive.New(client)
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

func SetGoogleDriveCreds(data sm.SessionMngrResponse) GoogleDriveCreds {
	googleCreds = &GoogleDriveCreds{}

	log.Println("the data ", data)
	googleCreds.Web.ClientId = data.SessionData.SessionVariables["GoogleDriveClientID"]
	googleCreds.Web.ProjectId = data.SessionData.SessionVariables["GoogleDriveProject"]
	googleCreds.Web.AuthURI = data.SessionData.SessionVariables["GoogleDriveAuthURI"]
	googleCreds.Web.TokenURI = data.SessionData.SessionVariables["GoogleDriveTokenURI"]
	googleCreds.Web.AuthProviderx509CertUrl = data.SessionData.SessionVariables["GoogleDriveAuthProviderx509CertUrl"]
	googleCreds.Web.ClientSecret = data.SessionData.SessionVariables["GoogleDriveClientSecret"]
	googleCreds.Web.RedirectURIS = strings.Split([]string{data.SessionData.SessionVariables["GoogleDriveRedirectUris"]}[0], ",")
	AccessCreds = data.SessionData.SessionVariables["GoogleDriveAccessCreds"]
	return *googleCreds
}

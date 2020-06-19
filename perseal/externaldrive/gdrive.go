package externaldrive

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/EC-SEAL/perseal/model"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
)

// Requests a token from the web, then returns the retrieved token.
func GetGoogleLinkForDashboardRedirect(id string, config *oauth2.Config) string {
	var authURL string
	if model.Local {
		authURL = config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("redirect_uri", "http://localhost:8082/per/code"), oauth2.SetAuthURLParam("state", id))
	} else {
		authURL = config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("redirect_uri", os.Getenv("REDIRECT_URL")), oauth2.SetAuthURLParam("state", id))
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
	log.Println(service)
	log.Println(fileProps.Path)
	dir, err := createGoogleDriveDir(service, fileProps.Path, "root")

	if err != nil {
		return nil, err
	}
	file, err = createGoogleDriveFile(service, fileProps.Name, fileProps.ContentType, bytes.NewReader(fileProps.Blob), dir.Id)
	if err != nil {
		return nil, err
	}
	// TODO check md5sum of data with CreatedFile
	// md5sum := md5.Sum(fileProps.Blob)
	return file, err
}

func SetGoogleDriveCreds() model.GoogleDriveCreds {
	googleCreds := &model.GoogleDriveCreds{}
	if model.Local {
		googleCreds.Web.ClientId = "425112724933-9o8u2rk49pfurq9qo49903lukp53tbi5.apps.googleusercontent.com"
		googleCreds.Web.ProjectId = "seal-274215"
		googleCreds.Web.AuthURI = "https://accounts.google.com/o/oauth2/auth"
		googleCreds.Web.TokenURI = "https://oauth2.googleapis.com/token"
		googleCreds.Web.AuthProviderx509CertUrl = "https://www.googleapis.com/oauth2/v1/certs"
		googleCreds.Web.ClientSecret = "0b3WtqfasYfWDmk31xa8UAht"
		googleCreds.Web.RedirectURIS = []string{"http://localhost:8082/per/code"}
	} else {
		googleCreds.Web.ClientId = os.Getenv("GOOGLE_DRIVE_CLIENT_ID")
		googleCreds.Web.ProjectId = os.Getenv("GOOGLE_DRIVE_CLIENT_PROJECT")
		googleCreds.Web.AuthURI = os.Getenv("GOOGLE_DRIVE_AUTH_URI")
		googleCreds.Web.TokenURI = os.Getenv("GOOGLE_DRIVE_TOKEN_URI")
		googleCreds.Web.AuthProviderx509CertUrl = os.Getenv("GOOGLE_DRIVE_AUTH_PROVIDER")
		googleCreds.Web.ClientSecret = os.Getenv("GOOGLE_DRIVE_CLIENT_SECRET")
		googleCreds.Web.RedirectURIS = strings.Split([]string{os.Getenv("GOOGLE_DRIVE_REDIRECT_URIS")}[0], ",")
	}
	return *googleCreds
}

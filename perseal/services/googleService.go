package services

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/externaldrive"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

// GOOGLE DRIVE SERVICE METHODS

//Attempts to Store the Session Data On Google Drive
func storeSessionDataGoogleDrive(dto dto.PersistenceDTO) (dataStore *externaldrive.DataStore, err *model.HTMLResponse) {
	client := getGoogleDriveClient(dto.GoogleAccessCreds)
	dataStore, erro := externaldrive.StoreSessionData(dto)
	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedEncryption, erro.Error())
		return
	}
	erro = uploadGoogleDrive(dataStore, client)
	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedDataStoreStoringInFile, erro.Error())
	}
	return
}

//Attempts to Load a Datastore from GoogleDrive into Session
func loadSessionDataGoogleDrive(dto dto.PersistenceDTO, filename string) (file *http.Response, err *model.HTMLResponse) {

	client := getGoogleDriveClient(dto.GoogleAccessCreds)
	file, erro := getGoogleDriveFile(filename, client)
	if erro != nil {
		err = model.BuildResponse(http.StatusNotFound, model.Messages.FailedGetCloudFile+model.EnvVariables.Google_Drive_PDS, erro.Error())
	}
	return
}

func getGoogleRedirectURL(id string) (url string) {
	config := establishGoogleDriveCreds()
	url = getGoogleLinkForDashboardRedirect(id, config)
	return
}

func getGoogleDriveClient(accessCreds oauth2.Token) (client *http.Client) {
	googleCreds := establishGoogleDriveCreds()

	b2, _ := json.Marshal(googleCreds)
	config, _ := google.ConfigFromJSON([]byte(b2), drive.DriveFileScope)
	client = config.Client(context.Background(), &accessCreds)
	return
}

// Uploads Google Drive Token to SessionVariables
func updateNewGoogleDriveTokenFromCode(id string, code string) (tok *oauth2.Token, err *model.HTMLResponse) {

	config := establishGoogleDriveCreds()

	tok, erro := config.Exchange(oauth2.NoContext, code)
	if erro != nil {
		err = model.BuildResponse(http.StatusNotFound, model.Messages.FailedGetToken+model.EnvVariables.Google_Drive_PDS, erro.Error())
		return
	}

	b, erro := json.Marshal(tok)
	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedParseToken+model.EnvVariables.Google_Drive_PDS, erro.Error())
		return
	}

	err = sm.UpdateSessionData(id, string(b), model.EnvVariables.SessionVariables.GoogleDriveToken)
	return
}

// Uploads new GoogleDrive data
func establishGoogleDriveCreds() (config *oauth2.Config) {
	googleCreds := model.EnvVariables.GoogleDriveCreds
	b2, _ := json.Marshal(googleCreds)
	config, _ = google.ConfigFromJSON([]byte(b2), drive.DriveFileScope)
	return

}

func getGoogleDriveFile(filename string, client *http.Client) (file *http.Response, err error) {
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

// Requests a token from the web, then returns the retrieved token.
func getGoogleLinkForDashboardRedirect(id string, config *oauth2.Config) string {
	var authURL string
	if model.Test {
		authURL = config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("redirect_uri", model.EnvVariables.Redirect_URL), oauth2.SetAuthURLParam("state", id), oauth2.SetAuthURLParam("user_id", model.EnvVariables.Project_SEAL_Email))
	} else {
		authURL = config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("redirect_uri", model.EnvVariables.Redirect_URL), oauth2.SetAuthURLParam("state", id))
	}
	return authURL
}

func getGoogleDriveFiles(client *http.Client) (fileList []string, err error) {
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

// Google Drive Upload Methods

type FileProps struct {
	Id          string
	Name        string
	Path        string
	Blob        []byte
	Md5sum      string
	ContentType string
}

// UploadGoogleDrive - Uploads file to Google Drive
func uploadGoogleDrive(ds *externaldrive.DataStore, client *http.Client) (err error) {
	data, err := ds.UploadingBlob()
	if err != nil {
		return
	}

	fp := &FileProps{
		Id:          ds.ID,
		Name:        model.EnvVariables.DataStore_File_Name, //TODO what should the name of the Blob be in Gdrive???
		Path:        model.EnvVariables.DataStore_Folder_Name,
		Blob:        data,
		ContentType: "application/octet-stream",
	}
	err = sendFile(fp, client)
	return
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

func createGoogleDriveFile(service *drive.Service, name string, mimeType string, content io.Reader, parentId string) (err error) {
	files, err := service.Files.List().Do()
	if err != nil {
		return
	}
	f := &drive.File{
		MimeType: mimeType,
		Name:     name,
		Parents:  []string{parentId},
	}
	for _, v := range files.Files {
		if v.Name == name && !isFolder(v) {
			service.Files.Delete(v.Id).Do()
		}
	}

	_, err = service.Files.Create(f).Media(content).Do()
	return
}

//SendFile gdrive file given encrypted blob and oauth token
func sendFile(fileProps *FileProps, client *http.Client) (err error) {

	service, err := drive.New(client)
	if err != nil {
		return
	}
	// Creates dir if it doesnt already exist
	dir, err := createGoogleDriveDir(service, fileProps.Path, "root")

	if err != nil {
		return
	}
	err = createGoogleDriveFile(service, fileProps.Name, fileProps.ContentType, bytes.NewReader(fileProps.Blob), dir.Id)
	if err != nil {
		return
	}
	// md5sum := md5.Sum(fileProps.Blob)
	return
}

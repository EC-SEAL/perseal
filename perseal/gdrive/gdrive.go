package gdrive

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

type GoogleDrive struct {
}

// Retrieves a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	tok, err := TokenFromSessionData()
	log.Println(err)
	if err != nil {
		tok = getTokenFromWeb(config)
	}
	return config.Client(context.Background(), tok)
}

// Requests a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("redirect_uri", "http://seal.me:8082"))
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	defer f.Close()
	if err != nil {
		log.Fatalf("Unable to cache OAuth token: %v", err)
	}
	json.NewEncoder(f).Encode(token)
}

func getService(oauthToken *oauth2.Token) (*drive.Service, error) {
	b := os.Getenv("GOOGLE_DRIVE_CREDENTIALS")
	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON([]byte(b), drive.DriveFileScope)

	// Testing
	// x := getTokenFromWeb(config)
	// log.Println(x)

	if err != nil {
		return nil, err
	}
	client := getClient(config)
	//client := config.Client(context.Background(), oauthToken)

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

func createDir(service *drive.Service, name string, parentId string) (*drive.File, error) {
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

func createFile(service *drive.Service, name string, mimeType string, content io.Reader, parentId string) (*drive.File, error) {
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
	err = json.NewDecoder(strings.NewReader(os.Getenv("GOOGLE_DRIVE_CLIENT"))).Decode(tok)
	return tok, err
}

//SendFile gdrive file given encrypted blob and oauth token
func SendFile(fileProps *FileProps, oauthToken *oauth2.Token) (file *drive.File, err error) {

	service, err := getService(oauthToken)
	if err != nil {
		return
	}
	// Creates dir if it doesnt already exist
	dir, err := createDir(service, fileProps.Path, "root")

	if err != nil {
		// return nil, errors.New("Could not create folder")
		return nil, err
	}
	file, err = createFile(service, fileProps.Name, fileProps.ContentType, bytes.NewReader(fileProps.Blob), dir.Id)
	if err != nil {
		// return nil, errors.New("Could not create file")
		return nil, err
	}
	// TODO check md5sum of data with CreatedFile
	// md5sum := md5.Sum(fileProps.Blob)
	return file, err
}

func GetFile(fileName string, oauthToken *oauth2.Token) (file *drive.File, err error) {
	service, err := getService(oauthToken)
	file, err = service.Files.Get(fileName).Do()
	return file, err
}

package externaldrive

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/model"
	"google.golang.org/api/drive/v3"
)

// UploadGoogleDrive - Uploads file to Google Drive
func (ds DataStore) UploadGoogleDrive(client *http.Client) (err error) {
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
	log.Println(err)
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

func createGoogleDriveFile(service *drive.Service, name string, mimeType string, content io.Reader, parentId string) (err error) {
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

	_, err = service.Files.Create(f).Media(content).Do()
	return
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
func sendFile(fileProps *FileProps, client *http.Client) (err error) {

	service, err := drive.New(client)
	if err != nil {
		return
	}
	// Creates dir if it doesnt already exist
	log.Println(service)
	log.Println(fileProps.Path)
	dir, err := createGoogleDriveDir(service, fileProps.Path, "root")

	if err != nil {
		return
	}
	err = createGoogleDriveFile(service, fileProps.Name, fileProps.ContentType, bytes.NewReader(fileProps.Blob), dir.Id)
	if err != nil {
		return
	}
	// TODO check md5sum of data with CreatedFile
	// md5sum := md5.Sum(fileProps.Blob)
	return
}

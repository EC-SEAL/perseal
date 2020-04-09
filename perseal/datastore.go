package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"regexp"

	"github.com/EC-SEAL/perseal/gdrive"
	"github.com/EC-SEAL/perseal/sm"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
)

// // DataStore sent in POST /per/load/{sessionToken} and received in POST /per/store/{sessionToken}
// type DataStore struct {
// 	ID                  string `json:"id"`
// 	EncryptedData       string `json:"encryptedData"`
// 	Signature           string `json:"signature"`
// 	SignatureAlgorithm  string `json:"signatureAlgorithm"`
// 	EncryptionAlgorithm string `json:"encryptionAlgorithm"`
// 	ClearData           []struct {
// 		ID         string    `json:"id"`
// 		Type       string    `json:"type"`
// 		Categories []string  `json:"categories"`
// 		IssuerID   string    `json:"issuerId"`
// 		SubjectID  string    `json:"subjectId"`
// 		Loa        string    `json:"loa"`
// 		Issued     time.Time `json:"issued"`
// 		Expiration time.Time `json:"expiration"`
// 		Attributes []struct {
// 			Name         string   `json:"name"`
// 			FriendlyName string   `json:"friendlyName"`
// 			Encoding     string   `json:"encoding"`
// 			Language     string   `json:"language"`
// 			IsMandatory  bool     `json:"isMandatory"`
// 			Values       []string `json:"values"`
// 		} `json:"attributes"`
// 		Properties struct {
// 			AdditionalProp1 string `json:"additionalProp1"`
// 			AdditionalProp2 string `json:"additionalProp2"`
// 			AdditionalProp3 string `json:"additionalProp3"`
// 		} `json:"properties"`
// 	} `json:"clearData"`
// }

// DataStore sent in POST /per/load/{sessionToken} and received in POST /per/store/{sessionToken}
type DataStore struct {
	ID                  string      `json:"id"`
	EncryptedData       string      `json:"encryptedData"`
	Signature           string      `json:"signature"`
	SignatureAlgorithm  string      `json:"signatureAlgorithm"`
	EncryptionAlgorithm string      `json:"encryptionAlgorithm"`
	ClearData           interface{} `json:"clearData,omitempty"`
}

// NewDataStore creates a new DataStore object
func NewDataStore(ID string, data interface{}) (ds *DataStore, err error) {
	if data == nil {
		return nil, errors.New("Cannot store empty data")
	}
	//Validates UUID
	uuid := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
	if !uuid.MatchString(ID) {
		return nil, errors.New("Invalid UUID")
	}

	ds = &DataStore{
		ID:        ID,
		ClearData: data,
	}
	return
}

// Encrypt encrypts `ClearData` stored in the DataStore into `EncryptedData`
func (ds *DataStore) Encrypt(cipherPassword string) (err error) {
	data, err := json.MarshalIndent(ds.ClearData, "", "\t")
	if err != nil {
		return
	}
	// blob, err = AESEncrypt(Pbkdf2([]byte(cipherPassword)), data)
	blob, err := AESEncrypt(Padding([]byte(cipherPassword), 16), data)
	// blob, err = AESEncrypt([]byte(cipherPassword), data)
	if err != nil {
		return
	}
	ds.EncryptedData = base64.URLEncoding.EncodeToString(blob)
	ds.EncryptionAlgorithm = "aes-cfb"
	return
}

// Sign signs the DataStore with rsa-sha256
func (ds *DataStore) Sign(privateKey []byte) (err error) { //TODO
	//TODO
	ds.SignatureAlgorithm = "rsa-sha256"
	return
}

// Decrypt decrypts `EncryptedData` stored in the DataStore into `ClearData`
func (ds *DataStore) Decrypt(cipherPassword string) (err error) {
	data, err := AESDecrypt(Padding([]byte(cipherPassword), 16), ds.EncryptedData)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &ds.ClearData)
	return
}

var (
	gdriveRootFolder string = "SEAL"
)

func (ds *DataStore) marshalWithoutClearText() (res []byte, err error) {
	return json.MarshalIndent(&DataStore{
		ID:                  ds.ID,
		EncryptedData:       ds.EncryptedData,
		EncryptionAlgorithm: ds.EncryptionAlgorithm,
		Signature:           ds.Signature,
		SignatureAlgorithm:  ds.SignatureAlgorithm,
	}, "", "\t")
}

// Upload the DataStore given the user's oauthToken
func (ds *DataStore) Upload(oauthToken *oauth2.Token) (data []byte, err error) {
	log.Println("Uploading Blob ", ds.ID)
	if ds.EncryptedData != "" {
		data, err = ds.marshalWithoutClearText()
		log.Println(string(data))
	} else {
		log.Println("No Encryption data for this DataStore - storing as plaintext")
		data, err = json.MarshalIndent(ds, "", "\t")
	}
	if err != nil {
		return nil, err
	}
	return
}

// UploadGoogleDrive - Uploads file to Google Drive
func (ds *DataStore) UploadGoogleDrive(oauthToken *oauth2.Token) (file *drive.File, err error) {
	data, _ := ds.Upload(oauthToken)
	fp := &gdrive.FileProps{
		Id:          ds.ID,
		Name:        ds.ID, //TODO what should the name of the Blob be in Gdrive???
		Path:        gdriveRootFolder,
		Blob:        data,
		ContentType: "application/octet-stream",
	}
	file, err = gdrive.SendFile(fp, oauthToken)
	return
}

// UploadOneDrive - Uploads file to One Drive
func (ds *DataStore) UploadOneDrive(oauthToken *oauth2.Token, data []byte) (file *drive.File, err error) {

	//if the folder exists, only creats the datastore file
	fileExists := getFolder(oauthToken, folderName)

	if fileExists.StatusCode == 404 {
		log.Println("eieieiie")
		folderID := createFolder(oauthToken)
		createFile(oauthToken, folderID, data)
	} else {
		log.Println("eieieiie!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!1")
		folderID := getFolderID(fileExists)
		createFile(oauthToken, folderID, data)
		log.Println("yes")
	}
	return
}

// data = sessionData
func storeSessionData(data interface{}, uuid, cipherPassword string) (dataStore *DataStore, err error) {
	dataStore, err = NewDataStore(uuid, data)
	if err != nil {
		return
	}
	// Encrypt blob if cipherPassword param is set
	if cipherPassword != "" {
		err = dataStore.Encrypt(cipherPassword)
		if err != nil {
			return
		}
		log.Println("Encrypted blob: ", dataStore.EncryptedData)
	}
	return
}

func storeSessionDataGoogleDrive(data interface{}, uuid, cipherPassword string) (datastore *DataStore, err error) {
	datastore, err = storeSessionData(data, uuid, cipherPassword)
	oauthToken, _ := gdrive.TokenFromSessionData() //TODO
	_, err = datastore.UploadGoogleDrive(oauthToken)
	return
}

func storeSessionDataOneDrive(data sm.SessionMngrResponse, uuid, cipherPassword string) (datastore *DataStore, err error) {
	datastore, _ = storeSessionData(data, uuid, cipherPassword)
	oauthToken := getToken(data.SessionData.SessionVariables.OneDriveClient, data.SessionData.SessionVariables.OneDriveScopes)
	contents, _ := datastore.Upload(oauthToken)
	_, err = datastore.UploadOneDrive(oauthToken, contents)
	return
}

func storeFileOneDriveClearText(data sm.SessionMngrResponse, uuid, cipherPassword string, contents interface{}) (datastore *DataStore, err error) {
	datastore, _ = storeSessionData(data, uuid, cipherPassword)
	oauthToken := getToken(data.SessionData.SessionVariables.OneDriveClient, data.SessionData.SessionVariables.OneDriveScopes)
	b, _ := json.Marshal(contents)
	_, err = datastore.UploadOneDrive(oauthToken, b)
	return
}

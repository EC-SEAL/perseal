package externaldrive

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"

	"github.com/EC-SEAL/perseal/utils"
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

/*  Mock dataStore value
{
  "id": "6c0f70a8-f32b-4535-b5f6-0d596c52813a",
  "encryptedData": "string",
  "signature": "string",
  "signatureAlgorithm": "string",
  "encryptionAlgorithm": "string",
  "clearData": [
    {
      "id": "6c0f70a8-f32b-4535-b5f6-0d596c52813a",
      "type": "string",
      "categories": [
        "string"
      ],
      "issuerId": "string",
      "subjectId": "string",
      "loa": "string",
      "issued": "2018-12-06T19:40:16Z",
      "expiration": "2018-12-06T19:45:16Z",
      "attributes": [
        {
          "name": "http://eidas.europa.eu/attributes/naturalperson/CurrentGivenName",
          "friendlyName": "CurrentGivenName",
          "encoding": "plain",
          "language": "ES_es",
          "isMandatory": true,
          "values": [
            "JOHN"
          ]
        }
      ],
      "properties": {
        "additionalProp1": "string",
        "additionalProp2": "string",
        "additionalProp3": "string"
      }
    }
  ]
}
*/

type DataStore struct {
	ID                  string      `json:"id"`
	EncryptedData       string      `json:"encryptedData"`
	Signature           string      `json:"signature"`
	SignatureAlgorithm  string      `json:"signatureAlgorithm"`
	EncryptionAlgorithm string      `json:"encryptionAlgorithm"`
	ClearData           interface{} `json:"clearData,omitempty"`
}

// DataStore sent in POST /per/load/{sessionToken} and received in POST /per/store/{sessionToken}

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
	blob, err := utils.AESEncrypt(utils.Padding([]byte(cipherPassword), 16), data)
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
	data, err := utils.AESDecrypt(utils.Padding([]byte(cipherPassword), 16), ds.EncryptedData)
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
func (ds *DataStore) UploadingBlob(oauthToken *oauth2.Token) (data []byte, err error) {
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
func (ds DataStore) UploadGoogleDrive(oauthToken *oauth2.Token, client *http.Client) (file *drive.File, err error) {
	data, err := ds.UploadingBlob(oauthToken)

	fp := &FileProps{
		Id:          ds.ID,
		Name:        "datastore.txt", //TODO what should the name of the Blob be in Gdrive???
		Path:        gdriveRootFolder,
		Blob:        data,
		ContentType: "application/octet-stream",
	}
	file, err = SendFile(fp, client)
	return
}

// UploadOneDrive - Uploads file to One Drive
func (ds *DataStore) UploadOneDrive(oauthToken *oauth2.Token, data []byte, folderName string) (file *drive.File, err error) {

	//if the folder exists, only creats the datastore file
	fileExists, err := GetOneDriveFolder(oauthToken, folderName)
	if err != nil {
		return
	}

	var folderID string
	if fileExists.StatusCode == 404 {
		folderID, err = CreateOneDriveFolder(oauthToken)
		if err != nil {
			return
		}
		err = CreateOneDriveFile(oauthToken, folderID, data)
		if err != nil {
			return
		}
	} else {
		folderID, err = GetOneDriveFolderID(fileExists)
		if err != nil {
			return
		}
		err = CreateOneDriveFile(oauthToken, folderID, data)
	}
	return
}

// data = sessionData
func StoreSessionData(data interface{}, uuid, cipherPassword string) (dataStore *DataStore, err error) {
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

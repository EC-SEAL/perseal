package externaldrive

import (
	"encoding/base64"
	"encoding/json"
	"log"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/EC-SEAL/perseal/utils"
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

// NewDataStore creates a new DataStore object
func NewDataStore(data sm.SessionMngrResponse) (ds *DataStore, err error) {

	currentDs := &DataStore{}
	var inter interface{}
	json.Unmarshal([]byte(data.SessionData.SessionVariables[model.EnvVariables.SessionVariables.DataStore]), &inter)

	jsonM, err := json.Marshal(inter)
	if err != nil {
		return
	}
	json.Unmarshal(jsonM, &currentDs)

	log.Println(data)
	log.Println(currentDs)

	sessionWithoutDataStore := data.SessionData.SessionVariables
	delete(sessionWithoutDataStore, model.EnvVariables.SessionVariables.DataStore)

	ds = &DataStore{
		ID:        currentDs.ID,
		ClearData: sessionWithoutDataStore,
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

// Signs DataStore using the certificate key
func (ds *DataStore) SignDataStore() (err error) {
	b64dec, err := utils.GetSignature(ds.EncryptedData)
	if err != nil {
		return
	}
	ds.Signature = string(b64dec)
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

func (ds *DataStore) MarshalWithoutClearText() (res []byte, err error) {
	return json.MarshalIndent(&DataStore{
		ID:                  ds.ID,
		EncryptedData:       ds.EncryptedData,
		EncryptionAlgorithm: ds.EncryptionAlgorithm,
		Signature:           ds.Signature,
		SignatureAlgorithm:  ds.SignatureAlgorithm,
	}, "", "\t")
}

// Upload the DataStore given the user's oauthToken
func (ds *DataStore) UploadingBlob() (data []byte, err error) {
	log.Println("Uploading Blob ", ds.ID)
	if ds.EncryptedData != "" {
		data, err = ds.MarshalWithoutClearText()
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

func StoreSessionData(dto dto.PersistenceDTO) (dataStore *DataStore, err error) {
	tmpDataStore, err := NewDataStore(dto.SMResp)
	if err != nil {
		return
	}
	// Encrypt blob if cipherPassword param is set
	err = tmpDataStore.Encrypt(dto.Password)
	if err != nil {
		return
	}

	log.Println("Encrypted blob: ", tmpDataStore.EncryptedData)
	err = tmpDataStore.SignDataStore()
	if err != nil {
		return
	}

	log.Println("DataStore Signed: ", tmpDataStore)
	dataStore = &DataStore{
		ID:                  tmpDataStore.ID,
		EncryptedData:       tmpDataStore.EncryptedData,
		EncryptionAlgorithm: tmpDataStore.EncryptionAlgorithm,
		Signature:           tmpDataStore.Signature,
		SignatureAlgorithm:  tmpDataStore.SignatureAlgorithm,
	}
	return
}

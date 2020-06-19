package services

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/externaldrive"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/EC-SEAL/perseal/utils"
	"golang.org/x/oauth2"
)

func GetRedirectURL(dto dto.PersistenceDTO) (url string, err *model.HTMLResponse) {

	if dto.PDS == "googleDrive" && dto.GoogleAccessCreds == "" {
		url, err = getGoogleRedirectURL(dto)
	} else if dto.PDS == "oneDrive" && dto.OneDriveToken.AccessToken == "" {
		url, err = getOneDriveRedirectURL(dto)
	}

	return
}

// Decrypts dataStore and loads it into session
func DecryptAndMarshallDataStore(dataStore *externaldrive.DataStore, dto dto.PersistenceDTO) (err *model.HTMLResponse) {

	erro := dataStore.Decrypt(dto.Password)
	dataStore.EncryptedData = ""
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         500,
			Message:      "Couldn't Decrypt DataStore",
			ErrorMessage: erro.Error(),
		}
		json.MarshalIndent(err, "", "\t")
		return
	}

	jsonM, erro := json.Marshal(dataStore)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         500,
			Message:      "Couldn't Parse Response Body from Get Session Data to Object",
			ErrorMessage: erro.Error(),
		}
		json.MarshalIndent(err, "", "\t")
		return
	}

	_, err = sm.UpdateSessionData(dto.ID, string(jsonM), "dataStore")

	return
}

func ValidateSignature(encrypted string, sigToValidate string) bool {
	sig, err := utils.GetSignature(encrypted)
	if err != nil {
		return false
	}

	log.Println("sig: ", sig)
	log.Println("toValidate: ", sigToValidate)
	if sig != sigToValidate {
		return false
	}
	log.Println("Validated")
	return true
}

func GetCloudFileNames(dto dto.PersistenceDTO) (files []string, err *model.HTMLResponse) {

	if dto.PDS == "googleDrive" {
		var client *http.Client
		_, client, err = getGoogleDriveClient(dto.GoogleAccessCreds)

		if err != nil {
			return
		}
		var erro error
		files, erro = externaldrive.GetGoogleDriveFiles(client)
		if erro != nil {
			err = &model.HTMLResponse{
				Code:         404,
				Message:      "Could not Get GoogleDrive Files",
				ErrorMessage: erro.Error(),
			}
			return
		}

	} else if dto.PDS == "oneDrive" {
		var token *oauth2.Token
		token, err = checkOneDriveTokenExpiry(dto)
		if err != nil {
			return
		}
		resp, erro := externaldrive.GetOneDriveItems(token, "SEAL")
		if erro != nil {
			err = &model.HTMLResponse{
				Code:         404,
				Message:      "Couldn't Get One Drive Items",
				ErrorMessage: erro.Error(),
			}
			return
		}
		for _, v := range resp.Values {
			files = append(files, v.Name)
		}
		log.Println(resp.Values)
	}
	return
}

// Reads the Body of a Given Datastore Response
func readBody(file *http.Response, id string) (ds *externaldrive.DataStore, err *model.HTMLResponse) {
	body, erro := ioutil.ReadAll(file.Body)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         400,
			Message:      "Couldn't Read Body From Response of Google Drive File",
			ErrorMessage: erro.Error(),
		}
		return
	}

	var v interface{}
	json.Unmarshal([]byte(body), &v)

	log.Println(v)
	jsonM, erro := json.Marshal(v)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         400,
			Message:      "Couldn't Parse the Body From Response of Google Drive File to JSON",
			ErrorMessage: erro.Error(),
		}
		return
	}

	json.Unmarshal(jsonM, &ds)

	_, err = sm.UpdateSessionData(id, string(jsonM), "dataStore")
	return
}

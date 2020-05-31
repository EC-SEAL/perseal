package services

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/externaldrive"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/EC-SEAL/perseal/utils"
	"golang.org/x/oauth2"
)

// if no files to load were found

func checkClientId(smResp sm.SessionMngrResponse, pds string) (returningSM sm.SessionMngrResponse, err *model.DashboardResponse) {

	id := smResp.SessionData.SessionID
	returningSM = smResp
	if pds == "googleDrive" {
		clientID := smResp.SessionData.SessionVariables["GoogleDriveAccessCreds"]
		// Validates if the session data contains the google drive authentication token
		if clientID == "" {
			returningSM.Error = "Session Data Not Correctly Set - Google Drive Client Missing"
			establishGoogleCredentials(id)
			returningSM, err = sm.GetSessionData(id, "")
			if err != nil {
				return
			}
		}
	} else if pds == "oneDrive" {
		clientID := smResp.SessionData.SessionVariables["OneDriveAccessToken"]
		log.Println(clientID)
		if clientID == "" {
			returningSM.Error = "Session Data Not Correctly Set - One Drive Client Missing"
			establishOneDriveCredentials(id)
			returningSM, err = sm.GetSessionData(id, "")

			if err != nil {
				return
			}
		}
	}
	return
}

// Decrypts dataStore and loads it into session
func DecryptAndMarshallDataStore(dataStore *externaldrive.DataStore, sessionToken string, cipherPassword string) (err *model.DashboardResponse) {

	erro := dataStore.Decrypt(cipherPassword)
	dataStore.EncryptedData = ""
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         404,
			Message:      "Couldn't Decrypt DataStore",
			ErrorMessage: erro.Error(),
		}
		json.MarshalIndent(err, "", "\t")
		return
	}

	jsonM, erro := json.Marshal(dataStore)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         404,
			Message:      "Couldn't Parse Response Body from Get Session Data to Object",
			ErrorMessage: erro.Error(),
		}
		json.MarshalIndent(err, "", "\t")
		return
	}

	_, err = sm.UpdateSessionData(sessionToken, string(jsonM), "dataStore")

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

func GetCloudFileNames(pds string, smResp sm.SessionMngrResponse) (files []string, err *model.DashboardResponse) {

	if pds == "googleDrive" {
		var client *http.Client
		client, err = getGoogleDriveClient(smResp)

		if err != nil {
			return
		}
		var erro error
		files, erro = externaldrive.GetGoogleDriveFiles(client)
		if erro != nil {
			err = &model.DashboardResponse{
				Code:         500,
				Message:      "Could not Get GoogleDrive Files",
				ErrorMessage: erro.Error(),
			}
			return
		}

	} else if pds == "oneDrive" {
		var oauthToken *oauth2.Token
		oauthToken, err = getOneDriveToken(smResp, smResp.SessionData.SessionID, "load&store")
		if err != nil {
			return
		}
		resp, erro := externaldrive.GetOneDriveItems(oauthToken, "SEAL")
		if erro != nil {
			err = &model.DashboardResponse{
				Code:         500,
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
func readBody(file *http.Response, id string) (ds *externaldrive.DataStore, err *model.DashboardResponse) {
	body, erro := ioutil.ReadAll(file.Body)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
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
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Couldn't Parse the Body From Response of Google Drive File to JSON",
			ErrorMessage: erro.Error(),
		}
		return
	}

	json.Unmarshal(jsonM, &ds)

	_, err = sm.UpdateSessionData(id, string(jsonM), "dataStore")
	return
}
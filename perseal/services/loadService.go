package services

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/EC-SEAL/perseal/externaldrive"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

// Service Method to Fetch the DataStore according to the PDS variable
func FetchDataStore(pds string, smResp sm.SessionMngrResponse) (dataStore *externaldrive.DataStore, err *model.DashboardResponse) {

	var file *http.Response
	if pds == "googleDrive" {
		googleCreds := externaldrive.SetGoogleDriveCreds(smResp)
		var token *oauth2.Token
		erro := json.NewDecoder(strings.NewReader(externaldrive.AccessCreds)).Decode(token)
		if err != nil {
			return
		}

		b2, erro := json.Marshal(googleCreds)
		if erro != nil {
			err = &model.DashboardResponse{
				Code:         500,
				Message:      "Google Creds JSON Malformed",
				ErrorMessage: erro.Error(),
			}
			return
		}

		config, erro := google.ConfigFromJSON([]byte(b2), drive.DriveFileScope)
		if err != nil {
			if erro != nil {
				err = &model.DashboardResponse{
					Code:         500,
					Message:      "Couldn't retrieve config from Google Creds JSON",
					ErrorMessage: erro.Error(),
				}
				return
			}

			client := config.Client(context.Background(), token)

			file, erro = externaldrive.GetGoogleDriveFile("datastore.txt", client)
			if erro != nil {
				err = &model.DashboardResponse{
					Code:         500,
					Message:      "Couldn't Get Google Drive File",
					ErrorMessage: erro.Error(),
				}
				return
			}
			dataStore, err = readBody(file, smResp.SessionData.SessionID)

			return
		}
	}

	if pds == "oneDrive" {
		log.Println("yapppppp")
		creds, erro := externaldrive.SetOneDriveCreds(smResp)
		if erro != nil {
			err = &model.DashboardResponse{
				Code:         500,
				Message:      "Couldn't Set One Drive Credentials",
				ErrorMessage: erro.Error(),
			}
			return
		}
		var oauthToken *oauth2.Token
		_, oauthToken, erro = externaldrive.GetOneDriveToken(creds)
		if erro != nil {
			err = &model.DashboardResponse{
				Code:         500,
				Message:      "Couldn't Get One Drive Token",
				ErrorMessage: erro.Error(),
			}
			return
		}
		log.Println(oauthToken)
		file, erro = externaldrive.GetOneDriveItem(oauthToken, "datastore.txt")
		if erro != nil {
			err = &model.DashboardResponse{
				Code:         500,
				Message:      "Couldn't Get One Drive Item",
				ErrorMessage: erro.Error(),
			}
			return
		}

		log.Println(file)
		dataStore, err = readBody(file, smResp.SessionData.SessionID)
		log.Println(dataStore)

		return
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

	_, err = sm.UpdateSessionData(id, string(jsonM), "DataStore")
	return
}

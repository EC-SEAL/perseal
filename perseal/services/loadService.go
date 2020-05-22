package services

import (
	"context"
	"encoding/json"
	"fmt"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/EC-SEAL/perseal/externaldrive"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/skip2/go-qrcode"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

// Service Method to Fetch the DataStore according to the PDS variable
func FetchCloudDataStore(pds string, smResp sm.SessionMngrResponse, filename *model.File) (dataStore *externaldrive.DataStore, err *model.DashboardResponse) {

	var file *http.Response
	if pds == "googleDrive" {
		var client *http.Client
		client, err = getGoogleDriveClient(smResp)
		log.Println(filename.Method)
		if filename.Method == "store" {
			err = &model.DashboardResponse{
				Code:    302,
				Message: "New Store Method",
			}
			log.Println(filename.Method)
			return
		}
		file, erro := externaldrive.GetGoogleDriveFile(filename.Filename, client)
		log.Println(file)
		if erro != nil {
			err = &model.DashboardResponse{
				Code:         500,
				Message:      "Couldn't Get Google Drive File",
				ErrorMessage: erro.Error(),
			}
			return
		}
		dataStore, err = readBody(file, smResp.SessionData.SessionID)
	} else if pds == "oneDrive" {
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
		model.Filename = make(chan model.File)
		filename := <-model.Filename
		log.Println(filename)
		close(model.Filename)
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
	}
	return
}

func FetchLocalDataStore(pds string, clientCallback string, smResp sm.SessionMngrResponse) bool {
	if pds == "Mobile" || pds == "Browser" {
		qr, _ := qrcode.New(clientCallback+"/cl/persistence/"+pds+"/load?sessionID="+smResp.SessionData.SessionID, qrcode.Medium)
		im := qr.Image(256)
		out, _ := os.Create("./QRImg.png")
		_ = png.Encode(out, im)
	} else if pds == "Browser" {
		//Asks SEAL ENCRYPTED PERSONAL DATASTORE
	}
	return true
}

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
	}
	return
}

func getGoogleDriveClient(smResp sm.SessionMngrResponse) (client *http.Client, err *model.DashboardResponse) {
	googleCreds := externaldrive.SetGoogleDriveCreds(smResp)
	var token *oauth2.Token = &oauth2.Token{}
	erro := json.NewDecoder(strings.NewReader(externaldrive.AccessCreds)).Decode(token)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Could not Decode Credentials to Token",
			ErrorMessage: erro.Error(),
		}
		return
	}

	fmt.Println(googleCreds)
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
	}

	client = config.Client(context.Background(), token)
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

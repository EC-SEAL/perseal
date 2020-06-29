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
	"golang.org/x/oauth2"
)

// Service Method to Fetch the DataStore according to the PDS variable
func FetchCloudDataStore(dto dto.PersistenceDTO, filename string) (dataStore *externaldrive.DataStore, err *model.HTMLResponse) {
	id := dto.ID
	var file *http.Response

	if dto.PDS == "googleDrive" {
		file, err = loadSessionDataGoogleDrive(dto, filename)
		if err != nil {
			return
		}
		dataStore, err = readBody(file, id)

	} else if dto.PDS == "oneDrive" {
		file, err = loadSessionDataOneDrive(dto, filename)
		if err != nil {
			return
		}
		dataStore, err = readBody(file, id)
		log.Println(dataStore)
	}
	return
}

func FetchLocalDataStore(r *http.Request) (ds *externaldrive.DataStore) {
	file, handler, _ := r.FormFile("file")
	defer file.Close()
	f, _ := handler.Open()
	body, erro := ioutil.ReadAll(f)
	if erro != nil {
		return
	}

	var v string
	str := string(body)
	log.Println("string", str)
	json.Unmarshal([]byte(str), &v)

	log.Println(v)
	if erro != nil {
		return
	}

	err := json.Unmarshal([]byte(v), &ds)
	log.Println(err)
	return
}

func GetCloudFileNames(dto dto.PersistenceDTO) (files []string, err *model.HTMLResponse) {

	if dto.PDS == "googleDrive" {
		var client *http.Client
		_, client = getGoogleDriveClient(dto.GoogleAccessCreds)
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
	body, _ := ioutil.ReadAll(file.Body)

	var v interface{}
	json.Unmarshal([]byte(body), &v)

	log.Println(v)
	jsonM, _ := json.Marshal(v)

	json.Unmarshal(jsonM, &ds)

	_, err = sm.UpdateSessionData(id, string(jsonM), "dataStore")
	return
}

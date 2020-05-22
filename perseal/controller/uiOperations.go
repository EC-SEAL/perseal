package controller

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/services"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/EC-SEAL/perseal/utils"
)

func showCloudFiles(w http.ResponseWriter, r *http.Request) {

}

func RecieveDataStoreFile(w http.ResponseWriter, r *http.Request) {
	log.Println("recieveDataStoreFile")
	log.Println(r.Body)

	bodybytes, erro := ioutil.ReadAll(r.Body)
	if erro != nil {
		err := &model.DashboardResponse{
			Code:         404,
			Message:      "Couldn't Read Body from Request",
			ErrorMessage: erro.Error(),
		}
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}
	filename := string(bodybytes)
	log.Println(filename)
	model.Filename <- filename
}

func RecievePassword(w http.ResponseWriter, r *http.Request) {
	log.Println("recievePassword")
	log.Println(r.Body)

	bodybytes, erro := ioutil.ReadAll(r.Body)
	methods := r.URL.Query()["method"]
	method := methods[0]
	if erro != nil {
		err := &model.DashboardResponse{
			Code:         404,
			Message:      "Couldn't Read Body from Request",
			ErrorMessage: erro.Error(),
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}
	password := string(bodybytes)
	log.Println(password)
	model.Password <- password

	log.Println("store")
	if method == "store" {
		model.Redirect = make(chan model.RedirectStruct)
		redirect := <-model.Redirect
		close(model.Redirect)
		if redirect.Redirect {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w = utils.WriteResponseMessage(w, redirect.URL, 302)
			return
		}
	}
}

func FetchCloudFiles(w http.ResponseWriter, r *http.Request) {
	log.Println("fetchCloudFiles")
	sessionTokens := r.URL.Query()["sessionToken"]
	sessionToken := sessionTokens[0]
	log.Println(sessionToken)
	smResp, err := sm.GetSessionData(sessionToken, "")

	if err != nil {
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}

	log.Println(smResp)
	pds := smResp.SessionData.SessionVariables["PDS"]

	if pds == "googleDrive" {
		clientId := smResp.SessionData.SessionVariables["GoogleDriveAccessCreds"]
		log.Println(clientId)
	} else if pds == "oneDrive" {
		clientId := smResp.SessionData.SessionVariables["OneDriveAccessToken"]
		log.Println(clientId)
	}
	files, _ := services.GetCloudFileNames(pds, smResp)
	log.Println(files)
	w.Header().Set("content-type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w = utils.WriteResponseMessage(w, files, 200)
	return
}

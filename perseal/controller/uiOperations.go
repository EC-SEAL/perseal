package controller

import (
	"fmt"
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/services"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/EC-SEAL/perseal/utils"
)

func GetSessionId(w http.ResponseWriter, r *http.Request) {
	log.Println("getSessionId")

	id, _, err := getSessionDataFromMSToken(r)

	log.Println(err)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("content-type", "application/json")

	if err != nil {
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}
	w = utils.WriteResponseMessage(w, id, 200)
	return

}
func ShowCloudFiles(w http.ResponseWriter, r *http.Request) {
	log.Println("showCloudFiles")

	//Recieves Current User
	log.Println(sm.CurrentUser)
	if sm.CurrentUser == nil {
		sm.CurrentUser = make(chan sm.SessionMngrResponse)
	}
	smResp := <-sm.CurrentUser

	log.Println("smResp in showCloudFiles ", smResp)
	pds := smResp.SessionData.SessionVariables["PDS"]
	var clientId string

	if pds == "googleDrive" {
		clientId = smResp.SessionData.SessionVariables["GoogleDriveAccessCreds"]
		log.Println(clientId)
	} else if pds == "oneDrive" {
		clientId = smResp.SessionData.SessionVariables["OneDriveAccessToken"]
		log.Println(clientId)
	}

	// If the Token doesn't exist, then the files don't exist as well
	if clientId == "" {
		var files []string
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("content-type", "application/json")
		w = utils.WriteResponseMessage(w, files, 200)
		return
	}

	files, err := services.GetCloudFileNames(pds, smResp)

	log.Println(files)
	log.Println(err)

	if err != nil {
		log.Println(err)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w = utils.WriteResponseMessage(w, err, err.Code)
	} else {
		w.Header().Set("content-type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w = utils.WriteResponseMessage(w, files, 200)
		log.Println("Correctly Fetched Cloud Files")
	}
	return
}

func CheckFirstAccess(w http.ResponseWriter, r *http.Request) {
	log.Println("CheckFirstAccess")

	if model.CheckFirstAccess == nil {
		model.CheckFirstAccess = make(chan bool)
	}

	toStore, err := utils.ReadRequestBody(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}

	if toStore == "true" {
		model.CheckFirstAccess <- true
	} else {
		model.CheckFirstAccess <- false
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w = utils.WriteResponseMessage(w, "", 200)
	return
}

func RedirectRequest(w http.ResponseWriter, r *http.Request) {
	log.Println("redirectRequest")

	log.Println(model.Redirect)
	if model.Redirect == nil {
		model.Redirect = make(chan model.RedirectStruct)
	}

	var redirect model.RedirectStruct
	redirect = <-model.Redirect

	log.Println("tenho")
	// If true, redirects to Login Page of Cloud
	if redirect.Redirect {
		fmt.Println(redirect)
		model.Redirect = nil
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w = utils.WriteResponseMessage(w, redirect.URL, 302)
		return
	} else {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w = utils.WriteResponseMessage(w, "", 200)
		return
	}

}

func RecievePassword(w http.ResponseWriter, r *http.Request) {
	log.Println("recievePassword")

	password, err := utils.ReadRequestBody(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}

	sha := utils.HashSUM256(password)

	model.Password <- sha
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w = utils.WriteResponseMessage(w, "", 200)
	return

}

func RetrieveCode(w http.ResponseWriter, r *http.Request) {
	log.Println("recieveCode")

	code, err := utils.ReadRequestBody(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w = utils.WriteResponseMessage(w, err, err.Code)
		return

	}

	log.Println(model.Code)
	if model.Code == nil {
		model.Code = make(chan string)
	}

	model.Code <- code
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w = utils.WriteResponseMessage(w, "", 200)
	return
}

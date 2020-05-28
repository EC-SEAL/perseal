package controller

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/services"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/EC-SEAL/perseal/utils"
)

func ShowCloudFiles(w http.ResponseWriter, r *http.Request) {
	log.Println("showCloudFiles")

	//Recieves Current User
	smResp := <-sm.CurrentUser
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

		w.Header().Set("content-type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w = utils.WriteResponseMessage(w, files, 200)
		return
	}

	files, err := services.GetCloudFileNames(pds, smResp)
	if err != nil {
		log.Println(err)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w = utils.WriteResponseMessage(w, err, err.Code)

	} else {
		w.Header().Set("content-type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w = utils.WriteResponseMessage(w, files, 200)
	}
	return
}

func RecieveDataStoreFile(w http.ResponseWriter, r *http.Request) {
	log.Println("recieveDataStoreFile")
	log.Println(r.Body)

	model.Redirect = make(chan model.RedirectStruct)
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
	methods := r.URL.Query()["method"]
	method := methods[0]
	filename := string(bodybytes)

	file := model.File{
		Filename: filename,
		Method:   method,
	}
	model.Filename <- file
	close(model.Filename)

	if method == "store" {
		var redirect model.RedirectStruct
		log.Println(model.Redirect)
		redirect = <-model.Redirect
		if redirect.Redirect {
			fmt.Println(redirect)
			close(model.Redirect)
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w = utils.WriteResponseMessage(w, redirect.URL, 302)
			return
		} else {

			w.Header().Set("Access-Control-Allow-Origin", "*")
			w = utils.WriteResponseMessage(w, "", 200)
			return
		}
	} else if method == "load" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w = utils.WriteResponseMessage(w, "", 200)
		return
	}

}
func RecievePassword(w http.ResponseWriter, r *http.Request) {
	log.Println("recievePassword")
	log.Println(r.Body)

	bodybytes, erro := ioutil.ReadAll(r.Body)
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
	hasher := sha1.New()
	hasher.Write([]byte(password))
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	log.Println(sha)
	model.Password <- sha
}

func RetrieveCode(w http.ResponseWriter, r *http.Request) {

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
	code := string(bodybytes)

	model.Password = make(chan string)
	model.Code <- code
}

func getSessionDataFromMSToken(msToken string) (id string, smResp sm.SessionMngrResponse, err *model.DashboardResponse) {
	id, err = sm.ValidateToken(msToken)
	if err != nil {
		return
	}
	smResp, err = sm.GetSessionData(id, "")

	if err != nil {
		return
	}

	log.Println(smResp)

	if err = sm.ValidateSessionMngrResponse(smResp); err != nil {
		return
	}

	return
}

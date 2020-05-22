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

func ShowCloudFiles(w http.ResponseWriter, r *http.Request) {
	log.Println("showCloudFiles")
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
	log.Println(clientId)
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

	model.Redirect = make(chan model.RedirectStruct)

	filename := string(bodybytes)
	log.Println(filename)
	log.Println(method)
	file := model.File{
		Filename: filename,
		Method:   method,
	}
	log.Println(file)
	model.Filename <- file
	close(model.Filename)

	log.Println("eu on redirect")
	log.Println("a ir ")
	if method == "store" {

		var redirect model.RedirectStruct
		log.Println(model.Redirect)
		redirect = <-model.Redirect

		log.Println("oeste")
		close(model.Redirect)
		log.Println("este")
		log.Println(&redirect)
		if redirect.Redirect {
			if redirect.Redirect {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w = utils.WriteResponseMessage(w, redirect.URL, 302)
				return
			}

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
	log.Println(password)
	model.Password <- password
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

	log.Println("e agr pass")
	model.Password = make(chan string)
	model.Code <- code
}

package controller

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/skip2/go-qrcode"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/services"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/EC-SEAL/perseal/utils"
)

//Opens HTML to display event message and redirect to ClientCallbackAddr
func writeResponseMessage(w http.ResponseWriter, dto dto.PersistenceDTO, response model.HTMLResponse) {
	dto.Response = response
	log.Println(dto.Response.ErrorMessage)
	log.Println(dto.Response.Message)
	if dto.Response.ClientCallbackAddr == "" {
		dto.Response.ClientCallbackAddr = dto.ClientCallbackAddr
	}
	log.Println(dto.Response.ClientCallbackAddr)
	t, _ := template.ParseFiles("ui/message.html")
	t.Execute(w, dto.Response)
}

//Opens HTML of corresponding operation (store or load)
func redirectToOperation(dto dto.PersistenceDTO, w http.ResponseWriter) {
	if dto.PDS == "Browser" {
		dto.IsLocal = true
		dto.StoreAndLoad = false
		log.Println(dto.IsLocal)
		t, _ := template.ParseFiles("ui/insertPassword.html")
		t.Execute(w, dto)
	} else if dto.PDS == "googleDrive" || dto.PDS == "oneDrive" || dto.PDS == "Mobile" {
		dto.IsLocal = false
		if dto.Method == "load" {
			if dto.PDS == "Mobile" {
				img, err := qrcode.Encode(dto.ClientCallbackAddr+"/cl/persistence/"+dto.PDS+"/load?sessionID="+dto.ID, qrcode.Medium, 256)
				dto.Image = base64.StdEncoding.EncodeToString(img)
				if err != nil {
					fmt.Print(err)
				}
				dto.StoreAndLoad = false
				t, _ := template.ParseFiles("ui/qr.html")
				t.Execute(w, dto)
				return
			} else {
				files, _ := services.GetCloudFileNames(dto)
				log.Println(files)

				if files == nil || len(files) == 0 {
					dto.DoesNotHaveFiles = true
					t, _ := template.ParseFiles("ui/noFilesFound.html")
					t.Execute(w, dto)
				} else {
					dto.StoreAndLoad = false
					t, _ := template.ParseFiles("ui/insertPassword.html")
					t.Execute(w, dto)
				}
			}
		} else {
			dto.StoreAndLoad = false
			t, _ := template.ParseFiles("ui/insertPassword.html")
			t.Execute(w, dto)
		}
	}
}

// Retrieves Password and SessionID from recieving request
func recieveSessionIdAndPassword(r *http.Request) (obj dto.PersistenceDTO, err *model.HTMLResponse) {
	password := r.FormValue("password")
	sha := utils.HashSUM256(password)
	id := r.FormValue("sessionId")
	sessionData, err := sm.GetSessionData(id, "")
	if err != nil {
		return
	}

	obj, err = dto.PersistenceWithPasswordBuilder(id, sessionData, sha)
	return
}

func initialConfig(id, method string, w http.ResponseWriter, r *http.Request) {
	smResp, err := sm.GetSessionData(id, "")
	if err != nil {
		obj, err := dto.PersistenceBuilder(id, sm.SessionMngrResponse{}, "")
		writeResponseMessage(w, obj, *err)
		return
	}

	obj, err := dto.PersistenceBuilder(id, smResp, method)
	log.Println(obj.Method)
	if err != nil {
		writeResponseMessage(w, obj, *err)
		return
	}

	log.Println(obj)
	url, err := services.GetRedirectURL(obj)

	log.Println(url)
	if url != "" {
		http.Redirect(w, r, url, 302)
	} else {
		redirectToOperation(obj, w)
	}
}

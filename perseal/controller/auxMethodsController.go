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

//Opens HTML of corresponding operation (store or load | local or cloud)
func redirectToOperation(dto dto.PersistenceDTO, w http.ResponseWriter) (url string) {

	if dto.PDS == "Mobile" {
		if dto.Method == "load" {
			mobileLoad(dto, w)
		} else if dto.Method == "store" {
			insertPassword(dto, w)
		}
	} else if dto.PDS == "Browser" {
		if dto.Method == "load" {
			dto.IsLocalLoad = true
		}
		insertPassword(dto, w)
	} else if dto.PDS == "googleDrive" || dto.PDS == "oneDrive" {
		url = services.GetRedirectURL(dto)

		log.Println(url)
		if url != "" {
			return
		}
		if dto.Method == "load" {
			files, _ := services.GetCloudFileNames(dto)
			log.Println(files)

			if files == nil || len(files) == 0 {
				noFilesFound(dto, w)
			} else {

				insertPassword(dto, w)
			}
		} else if dto.Method == "store" {
			insertPassword(dto, w)
		}
	}
	return
}

//Opens HTML to display event message and redirect to ClientCallbackAddr
func writeResponseMessage(w http.ResponseWriter, dto dto.PersistenceDTO, response model.HTMLResponse) {
	dto.Response = response
	if dto.Response.ClientCallbackAddr == "" {
		dto.Response.ClientCallbackAddr = dto.ClientCallbackAddr
	}
	t, _ := template.ParseFiles("ui/message.html")
	t.Execute(w, dto.Response)
}

func getQueryParameter(r *http.Request, paramName string) string {
	var param string
	if keys, ok := r.URL.Query()[paramName]; ok {
		param = keys[0]
	}
	log.Println(param)
	return param
}

func getSessionData(id string, w http.ResponseWriter) (smResp sm.SessionMngrResponse) {
	smResp, err := sm.GetSessionData(id, "")
	if err != nil {
		var obj dto.PersistenceDTO
		obj, err = dto.PersistenceBuilder(id, sm.SessionMngrResponse{})
		writeResponseMessage(w, obj, *err)
		return
	}
	return
}

func validateToken(token string, w http.ResponseWriter) (id string) {
	id, err := sm.ValidateToken(token)
	if err != nil {
		dto, _ := dto.PersistenceBuilder(id, sm.SessionMngrResponse{})
		writeResponseMessage(w, dto, *err)
	}
	return
}

func mobileLoad(dto dto.PersistenceDTO, w http.ResponseWriter) {
	img, err := qrcode.Encode(dto.ClientCallbackAddr+"/cl/persistence/"+dto.PDS+"/load?sessionID="+dto.ID, qrcode.Medium, 256)
	dto.Image = base64.StdEncoding.EncodeToString(img)
	if err != nil {
		fmt.Print(err)
	}
	t, _ := template.ParseFiles("ui/qr.html")
	t.Execute(w, dto)
	return
}

func noFilesFound(dto dto.PersistenceDTO, w http.ResponseWriter) {
	sm.UpdateSessionData(dto.ID, "storeload", "CurrentMethod")
	t, _ := template.ParseFiles("ui/noFilesFound.html")
	t.Execute(w, dto)
}

func insertPassword(dto dto.PersistenceDTO, w http.ResponseWriter) {
	t, _ := template.ParseFiles("ui/insertPassword.html")
	t.Execute(w, dto)
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

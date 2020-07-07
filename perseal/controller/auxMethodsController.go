package controller

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"io/ioutil"
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

	if dto.PDS == model.EnvVariables.Mobile_PDS {
		if dto.Method == model.EnvVariables.Load_Method {
			mobileQRCode(dto, model.EnvVariables.Load_Method, w)
		} else if dto.Method == model.EnvVariables.Store_Method {
			mobileQRCode(dto, model.EnvVariables.Store_Method, w)
		}
	} else if dto.PDS == model.EnvVariables.Browser_PDS {
		if dto.Method == model.EnvVariables.Load_Method {
			dto.IsLocalLoad = true
		}
		insertPassword(dto, w)
	} else if dto.PDS == model.EnvVariables.Google_Drive_PDS || dto.PDS == model.EnvVariables.One_Drive_PDS {
		url = services.GetRedirectURL(dto)

		log.Println(url)
		// Will redirect user to login on cloud account and send the Code to /per/code
		if url != "" {
			return
		}
		if dto.Method == model.EnvVariables.Load_Method {
			files, _ := services.GetCloudFileNames(dto)
			log.Println(files)

			if files == nil || len(files) == 0 {
				noFilesFound(dto, w)
			} else {

				insertPassword(dto, w)
			}
		} else if dto.Method == model.EnvVariables.Store_Method {
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
	if response.ErrorMessage != "" {
		log.Println(response.ErrorMessage)
	}
	t, _ := template.ParseFiles("ui/message.html")
	t.Execute(w, dto.Response)
}

func getQueryParameter(r *http.Request, paramName string) string {
	var param string
	if keys, ok := r.URL.Query()[paramName]; ok {
		param = keys[0]
	}
	return param
}

func getSessionData(id string, w http.ResponseWriter) (smResp sm.SessionMngrResponse) {
	smResp, err := sm.GetSessionData(id)
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

func mobileQRCode(dto dto.PersistenceDTO, method string, w http.ResponseWriter) {
	// TODO: check if this variable is indeed the custom URL. Needs confirmation
	customURL := dto.ClientCallbackAddr + "/cl/persistence/" + dto.PDS + "/" + method + "?sessionID=" + dto.ID
	img, err := qrcode.Encode(customURL, qrcode.Medium, 256)
	dto.Image = base64.StdEncoding.EncodeToString(img)
	if err != nil {
		fmt.Print(err)
		return
	}
	t, _ := template.ParseFiles("ui/qr.html")
	t.Execute(w, dto)
	return
}

func noFilesFound(dto dto.PersistenceDTO, w http.ResponseWriter) {
	sm.UpdateSessionData(dto.ID, model.EnvVariables.Store_Load_Method, "CurrentMethod")
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
	sessionData, err := sm.GetSessionData(id)
	if err != nil {
		return
	}

	obj, err = dto.PersistenceWithPasswordBuilder(id, sessionData, sha)
	return
}

func fetchLocalDataStore(r *http.Request) (body []byte, err *model.HTMLResponse) {
	var erro error

	file, handler, erro := r.FormFile("file")
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         404,
			Message:      "Could not find file contents",
			ErrorMessage: erro.Error(),
		}
		return
	}

	defer file.Close()
	f, erro := handler.Open()
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         404,
			Message:      "Could not open file",
			ErrorMessage: erro.Error(),
		}
		return
	}

	body, erro = ioutil.ReadAll(f)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         404,
			Message:      "Could not read file contents",
			ErrorMessage: erro.Error(),
		}
	}
	return
}

package controller

import (
	"encoding/base64"
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

var menuHTML = "ui/menu.html"
var insertPasswordHTML = "ui/insertPassword.html"

//Opens HTML of corresponding operation (store or load | local or cloud)
func redirectToOperation(dto dto.PersistenceDTO, w http.ResponseWriter) (url string) {
	if dto.PDS == model.EnvVariables.Mobile_PDS {
		if dto.Method == model.EnvVariables.Store_Method {
			mobileQRCode(dto, w)
		} else if dto.Method == model.EnvVariables.Load_Method {
			dto.MenuOption = "GenerateQRCode"
			openHTML(dto, w, menuHTML)
		}
	} else if dto.PDS == model.EnvVariables.Browser_PDS {
		if dto.Method == model.EnvVariables.Load_Method {
			dto.MenuOption = "BrowserOption"
			openHTML(dto, w, menuHTML)
		} else if dto.Method == model.EnvVariables.Store_Method {
			openHTML(dto, w, insertPasswordHTML)
		}
	} else if dto.PDS == model.EnvVariables.Google_Drive_PDS || dto.PDS == model.EnvVariables.One_Drive_PDS {
		url = services.GetRedirectURL(dto)
		// Will redirect user to login on cloud account and send the Code to /per/code
		if url != "" {
			log.Println("Redirecting to: " + url)
			sm.UpdateSessionData(dto.ID, dto.Method, model.EnvVariables.SessionVariables.CurrentMethod)
			return
		}
		if dto.Method == model.EnvVariables.Load_Method {
			files, _ := services.GetCloudFileNames(dto)
			log.Println(files)

			if files == nil || len(files) == 0 {
				dto.MenuOption = "NoFilesFound"
				openHTML(dto, w, menuHTML)
			} else {
				openHTML(dto, w, insertPasswordHTML)
			}
		} else if dto.Method == model.EnvVariables.Store_Method {
			openHTML(dto, w, insertPasswordHTML)
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
	log.Println(dto.Response)
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

func validateToken(token string, w http.ResponseWriter) (id string) {
	id, err := sm.ValidateToken(token)
	if err != nil {
		dto, _ := dto.PersistenceBuilder(id, sm.SessionMngrResponse{})
		writeResponseMessage(w, dto, *err)
	}
	return
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

func mobileQRCode(obj dto.PersistenceDTO, w http.ResponseWriter) {
	token, _ := utils.GenerateTokenAPI(obj.PDS, obj.ID)
	obj.CustomURL = model.EnvVariables.Dashboard_Custom_URL + obj.Method + "/" + token

	img, _ := qrcode.Encode(obj.CustomURL, qrcode.Medium, 380)
	obj.Image = base64.StdEncoding.EncodeToString(img)
	t, _ := template.ParseFiles("ui/qr.html")
	log.Println(obj.UserDevice)
	log.Println(obj.CustomURL)
	t.Execute(w, obj)
	return
}

func openHTML(obj dto.PersistenceDTO, w http.ResponseWriter, filename string) {
	var err *model.HTMLResponse
	obj.MSToken, err = utils.GenerateTokenAPI(obj.PDS, obj.ID)
	if err != nil {
		writeResponseMessage(w, obj, *err)
	}
	t, _ := template.ParseFiles(filename)
	t.Execute(w, obj)
}

// Retrieves Password and SessionID from recieving request
func recieveSessionIdAndPassword(w http.ResponseWriter, r *http.Request, method string) (obj dto.PersistenceDTO, err *model.HTMLResponse) {
	msToken := r.FormValue("msToken")
	id := validateToken(msToken, w)
	smResp := getSessionData(id, w)
	log.Println(id)
	password := r.FormValue("password")
	if password == "" {
		err = &model.HTMLResponse{
			Code:               400,
			Message:            "No Password Found",
			ClientCallbackAddr: smResp.SessionData.SessionVariables[model.EnvVariables.SessionVariables.ClientCallbackAddr],
		}

		if err.ClientCallbackAddr == "" && model.Test {
			err.ClientCallbackAddr = model.EnvVariables.TestURLs.MockRedirectDashboard
		}
		return
	}
	sha := utils.HashSUM256(password)

	obj, err = dto.PersistenceWithPasswordBuilder(id, sha, smResp, method)
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

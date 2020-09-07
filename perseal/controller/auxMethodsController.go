package controller

import (
	"encoding/base64"
	"encoding/json"
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

var (
	menuHTML           = "ui/menu.html"
	insertPasswordHTML = "ui/insertPassword.html"
)

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

func openMessageHTML(dto dto.PersistenceDTO, w http.ResponseWriter) {
	if dto.PDS == model.EnvVariables.Browser_PDS {
		tok, _ := sm.GenerateToken(model.EnvVariables.Perseal_Sender_Receiver, model.EnvVariables.Perseal_Sender_Receiver, dto.ID)
		dto.Response.MSTokenDownload = tok.AdditionalData
	}
	res := model.MarshallResponseToPrint(dto.Response)
	log.Println("Response Object: ", res)
	t, _ := template.ParseFiles("ui/message.html")
	w.WriteHeader(dto.Response.Code)
	t.Execute(w, dto.Response)

}

func openHTML(obj dto.PersistenceDTO, w http.ResponseWriter, filename string) {
	token, err := sm.GenerateToken(model.EnvVariables.Perseal_Sender_Receiver, model.EnvVariables.Perseal_Sender_Receiver, obj.ID)
	if err != nil {
		writeResponseMessage(w, obj, *err)
		return
	}

	obj.MSToken = token.AdditionalData
	t, _ := template.ParseFiles(filename)
	t.Execute(w, obj)
}

//Opens HTML to display event message and makes post request to ClientCallbackAddr
func writeResponseMessage(w http.ResponseWriter, dto dto.PersistenceDTO, response model.HTMLResponse) {
	dto.Response = response
	dto.MenuOption = response.FailedInput

	dto.Response.ClientCallbackAddr = dto.ClientCallbackAddr

	if dto.MenuOption != "" {
		openHTML(dto, w, menuHTML)
	} else {
		var tok1, tok2 string
		if dto.Response.Code == http.StatusOK {
			tok1, tok2 = services.BuildDataOfMSToken(dto.ID, "OK", dto.Response.ClientCallbackAddr)
			log.Println("Token contains OK message")
		} else {
			if dto.Response.ErrorMessage == model.Messages.NoMSTokenErrorMsg {
				dto.Response.MSToken = ""
			} else {
				tok1, tok2 = services.BuildDataOfMSToken(dto.ID, "ERROR", dto.Response.ClientCallbackAddr, "Failure! "+"\n"+dto.Response.Message+"\n"+dto.Response.ErrorMessage)
				log.Println("Token contains ERROR message")
			}
		}
		dto.Response.MSTokenRedirect = tok1
		dto.Response.MSToken = tok2
		if tok1 != "" && tok2 != "" {
			log.Println("Generated both tokens")
		}
		openMessageHTML(dto, w)
	}
}

func writeBackChannelResponse(dto dto.PersistenceDTO, w http.ResponseWriter) {
	if dto.MenuOption != "BadQR" {
		w.WriteHeader(dto.Response.Code)
		w.Write([]byte(dto.Response.Message))
	}

	var tok string
	if dto.Response.Code == http.StatusOK {
		tok, _ = services.BuildDataOfMSToken(dto.ID, "OK", dto.Response.ClientCallbackAddr)
		log.Println("Token contains OK message")
	} else {
		tok, _ = services.BuildDataOfMSToken(dto.ID, "ERROR", dto.Response.ClientCallbackAddr, "Failure! "+"\n"+dto.Response.Message+"\n"+dto.Response.ErrorMessage)
		log.Println("Token contains ERROR message")
	}
	services.ClientCallbackAddrPost(tok, dto.ClientCallbackAddr)
}

//Gets Query Parameter with a specific paramName from a Request (r)
func getQueryParameter(r *http.Request, paramName string) (param string) {
	if keys, ok := r.URL.Query()[paramName]; ok {
		param = keys[0]
	}
	return param
}

func getSessionData(id string, w http.ResponseWriter) (smResp sm.SessionMngrResponse) {
	smResp, err := sm.GetSessionData(id)
	if err != nil {
		var obj dto.PersistenceDTO
		obj, err = dto.PersistenceFactory(id, sm.SessionMngrResponse{})
		writeResponseMessage(w, obj, *err)
	}
	return
}

// Generates QR code and presents it in HTML
func mobileQRCode(obj dto.PersistenceDTO, w http.ResponseWriter) {

	type QRVariables struct {
		SessionId       string `json:"sessionId"`
		Method          string `json:"method"`
		PersealCallback string `json:"persealCallback"`
	}

	variables := QRVariables{
		SessionId:       obj.ID,
		Method:          obj.Method,
		PersealCallback: model.EnvVariables.Perseal_RM_UCs_Callback,
	}
	qrCodeContents, _ := json.Marshal(variables)
	img, _ := qrcode.Encode(string(qrCodeContents), qrcode.Medium, 380)
	obj.Image = base64.StdEncoding.EncodeToString(img)

	if containsEmpty(variables.SessionId, variables.Method, variables.PersealCallback) {
		resp := model.BuildResponse(http.StatusInternalServerError, model.Messages.IncompleteQRCode)
		writeResponseMessage(w, obj, *resp)
		return
	} else {
		resp := model.BuildResponse(http.StatusOK, model.Messages.PrintedQRCode)
		obj.Response = *resp
		obj.MenuOption = "BadQR"
		writeBackChannelResponse(obj, w)
	}

	t, _ := template.ParseFiles("ui/qr.html")
	t.Execute(w, obj)
	return
}

func containsEmpty(stringArray ...string) bool {
	for _, s := range stringArray {
		if s == "" {
			return true
		}
	}
	return false
}

// Retrieves Password and SessionID from recieving request
func recieveSessionIdAndPassword(w http.ResponseWriter, r *http.Request, method string) (obj dto.PersistenceDTO, err *model.HTMLResponse) {
	msToken := r.FormValue("msToken")
	smResp, err := sm.ValidateToken(msToken)
	id := smResp.SessionData.SessionID

	smResp = getSessionData(id, w)

	obj, err = dto.PersistenceFactory(id, smResp, method)
	password := r.FormValue("password")
	if password == "" {
		err = model.BuildResponse(http.StatusBadRequest, model.Messages.NoPassword)
		err.FailedInput = "Password"
		return
	}
	sha := utils.HashSUM256(password)
	obj.Password = sha
	return
}

func fetchLocalDataStore(r *http.Request) (body []byte, err *model.HTMLResponse) {
	file, handler, erro := r.FormFile("file")
	if erro != nil {
		err = model.BuildResponse(http.StatusNotFound, model.Messages.FileContentsNotFound, erro.Error())
		return
	}

	defer file.Close()
	f, erro := handler.Open()
	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedOpenFile, erro.Error())
		return
	}

	body, erro = ioutil.ReadAll(f)
	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedReadFile, erro.Error())
	}
	return
}

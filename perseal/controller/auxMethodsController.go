package controller

import (
	"encoding/base64"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

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

func openResponse(dto dto.PersistenceDTO, w http.ResponseWriter) {
	log.Println("ref: ", dto.Response)
	if dto.PDS == model.EnvVariables.Browser_PDS {
		tok, _ := sm.GenerateToken(model.EnvVariables.Perseal_Sender_Receiver, model.EnvVariables.Perseal_Sender_Receiver, dto.ID)
		dto.Response.MSTokenDownload = tok.AdditionalData
	}
	t, _ := template.ParseFiles("ui/message.html")
	w.WriteHeader(dto.Response.Code)
	t.Execute(w, dto.Response)

}

func buildDataOfMSToken(id, code, clientCallbackAddr string, message ...string) (string, string) {
	dash := &sm.SessionMngrResponse{
		SessionData: sm.SessionData{
			SessionID: id,
		},
		Code: code,
	}

	if len(message) > 0 || message != nil {
		dash.AdditionalData = message[0]
	}
	b, _ := json.Marshal(dash)
	tok1, err := sm.GenerateToken(model.EnvVariables.Perseal_Sender_Receiver, model.EnvVariables.Perseal_Sender_Receiver, id, string(b))
	if err != nil {
		log.Println(err)
		return "", ""
	}
	tok2, err := sm.GenerateToken(model.EnvVariables.Perseal_Sender_Receiver, model.EnvVariables.Perseal_Sender_Receiver, id)
	if err != nil {
		log.Println(err)
		return "", ""
	}
	return tok1.AdditionalData, tok2.AdditionalData
}

//Opens HTML to display event message and redirect to ClientCallbackAddr
func writeResponseMessage(w http.ResponseWriter, dto dto.PersistenceDTO, response model.HTMLResponse) {
	dto.Response = response
	dto.MenuOption = response.FailedInput

	dto.Response.ClientCallbackAddr = dto.ClientCallbackAddr

	if dto.MenuOption != "" {
		log.Println(dto.MenuOption)
		log.Println(dto.PDS)
		openHTML(dto, w, menuHTML)
	} else {
		var tok1, tok2 string
		if dto.Response.Code == http.StatusOK {
			tok1, tok2 = buildDataOfMSToken(dto.ID, "OK", dto.Response.ClientCallbackAddr)
		} else {
			if dto.Response.ErrorMessage == model.Messages.NoMSTokenErrorMsg {
				dto.Response.MSToken = ""
			} else {
				tok1, tok2 = buildDataOfMSToken(dto.ID, "ERROR", dto.Response.ClientCallbackAddr, "Failure! "+"\n"+dto.Response.Message+"\n"+dto.Response.ErrorMessage)

			}
		}
		dto.Response.MSTokenRedirect = tok1
		dto.Response.MSToken = tok2
		if tok1 != "" && tok2 != "" {
			log.Println("Generated both tokens")
		}
		openResponse(dto, w)
	}
}

func writeBackChannelResponse(dto dto.PersistenceDTO, w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	w.Write([]byte(message))

	var tok string
	if dto.Response.Code == http.StatusOK {
		tok, _ = buildDataOfMSToken(dto.ID, "OK", dto.Response.ClientCallbackAddr)
	} else {
		tok, _ = buildDataOfMSToken(dto.ID, "ERROR", dto.Response.ClientCallbackAddr, "Failure! "+"\n"+dto.Response.Message+"\n"+dto.Response.ErrorMessage)
	}
	clientCallbackAddrPost(tok, dto.ClientCallbackAddr)
}

func clientCallbackAddrPost(token, clientCallbackAddr string) {
	hc := http.Client{}
	form := url.Values{}
	form.Add("msToken", token)

	log.Println(clientCallbackAddr)
	req, _ := http.NewRequest(http.MethodPost, clientCallbackAddr, strings.NewReader(form.Encode()))
	log.Println(req)
	log.Println(hc.Do(req))
}

func getQueryParameter(r *http.Request, paramName string) string {
	var param string
	if keys, ok := r.URL.Query()[paramName]; ok {
		param = keys[0]
	}
	return param
}

func validateToken(token string, w http.ResponseWriter) (id string) {
	smResp, err := sm.ValidateToken(token)
	id = smResp.SessionData.SessionID

	log.Println(id)
	log.Println(err)
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
	}
	return
}

func mobileQRCode(obj dto.PersistenceDTO, w http.ResponseWriter) {

	type QRVariables struct {
		SessionId string `json:"sessionId"`
		Method    string `json:"method"`
	}

	variables := QRVariables{
		SessionId: obj.ID,
		Method:    obj.Method,
	}
	qrCodeContents, _ := json.Marshal(variables)

	img, _ := qrcode.Encode(string(qrCodeContents), qrcode.Medium, 380)
	obj.Image = base64.StdEncoding.EncodeToString(img)
	t, _ := template.ParseFiles("ui/qr.html")
	t.Execute(w, obj)
	return
}

func openHTML(obj dto.PersistenceDTO, w http.ResponseWriter, filename string) {
	var err *model.HTMLResponse
	token, err := sm.GenerateToken(model.EnvVariables.Perseal_Sender_Receiver, model.EnvVariables.Perseal_Sender_Receiver, obj.ID)
	if err != nil {
		writeResponseMessage(w, obj, *err)
	}
	obj.MSToken = token.AdditionalData
	t, _ := template.ParseFiles(filename)
	t.Execute(w, obj)
}

// Retrieves Password and SessionID from recieving request
func recieveSessionIdAndPassword(w http.ResponseWriter, r *http.Request, method string) (obj dto.PersistenceDTO, err *model.HTMLResponse) {
	msToken := r.FormValue("msToken")
	id := validateToken(msToken, w)
	log.Println("Current Session Id: " + id)
	smResp := getSessionData(id, w)

	obj, err = dto.PersistenceBuilder(id, smResp, method)
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
	var erro error

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

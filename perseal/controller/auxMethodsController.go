package controller

import (
	"encoding/base64"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/skip2/go-qrcode"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/services"
	"github.com/EC-SEAL/perseal/sm"
)

var (
	menuHTML           = "ui/menu.html"
	insertPasswordHTML = "ui/insertPassword.html"
)

type QRVariables struct {
	SessionId string `json:"sessionId"`
	Method    string `json:"method"`
}

// ================================== METHODS CALLED AT THE BEGINNING OF THE OPERATIONS ==============================

//Opens HTML of corresponding operation (store or load | local or cloud)
func redirectToOperation(dto dto.PersistenceDTO, w http.ResponseWriter) (url string) {

	//Mobile UC
	if dto.PDS == model.EnvVariables.Mobile_PDS {
		//Defines Contents of QRCode/msToken
		contents := QRVariables{
			Method:    dto.Method,
			SessionId: dto.ID,
		}
		log.Println("Contents of QRCode/msToken: ", contents)

		// Generate msToken with the variables
		b, _ := json.Marshal(contents)
		token, _ := services.BuildDataOfMSToken(dto.ID, "OK", dto.ClientCallbackAddr, string(b))

		// Makes request to Custom URL to attempt to open the mobile app.
		// If unreachable, redirect to EP to generate QRCode
		timeout := time.Duration(1 * time.Second)
		client := http.Client{
			Timeout: timeout,
		}
		url = model.EnvVariables.CustomURL + "?msToken=" + token
		_, err := client.Get(url)
		if err != nil {
			log.Println("Custom URL unreachable")
			url = model.EnvVariables.Perseal_QRCode_Endpoint + "?msToken=" + token
		}

		// Sets session flag to signify back-channel hasn't finished yet
		sm.UpdateSessionData(dto.ID, "not finished", model.EnvVariables.SessionVariables.FinishedPersealBackChannel)

		log.Println("Redirecting to: " + url)
		return

		//Local File System UC
	} else if dto.PDS == model.EnvVariables.Browser_PDS {

		if dto.Method == model.EnvVariables.Load_Method {
			dto.MenuOption = "BrowserOption"
			openInternalHTML(dto, w, menuHTML)
		} else if dto.Method == model.EnvVariables.Store_Method {
			openInternalHTML(dto, w, insertPasswordHTML)
		}

		//Cloud UC's
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
				openInternalHTML(dto, w, menuHTML)
			} else {
				openInternalHTML(dto, w, insertPasswordHTML)
			}
		} else if dto.Method == model.EnvVariables.Store_Method {
			openInternalHTML(dto, w, insertPasswordHTML)
		}
	}
	return
}

// Validates msToken, gets Session Data and Builds Persistence DTO (as well as other minor operations)
func initialEPSetup(w http.ResponseWriter, token, method string, backChannel bool, cipherPassword ...string) (obj dto.PersistenceDTO, cnt string, err *model.HTMLResponse) {

	smResp, err := sm.ValidateToken(token)
	id := smResp.SessionData.SessionID
	if err != nil {
		if len(cipherPassword) > 0 || cipherPassword != nil {
			if cipherPassword[0] != "" {
				w.WriteHeader(err.Code)
				w.Write([]byte(err.Message))
				return
			}

		} else {
			dto, _ := dto.PersistenceFactory(id, sm.SessionMngrResponse{})
			writeResponseMessage(w, dto, *err)
			return
		}
	}
	log.Println("MSToken Contents: " + smResp.AdditionalData)
	cnt = smResp.AdditionalData

	// EXCEPTION: Mobile Storage can be enable if cipherPassword is sent immediatly in the GET request
	if len(cipherPassword) > 0 || cipherPassword != nil {
		if cipherPassword[0] != "" {
			backChannelStoring(w, id, cipherPassword[0], method, smResp)
			return
		}
	}

	smResp = getSessionData(id, w)

	obj, err = dto.PersistenceFactory(id, smResp, method)
	if err != nil {
		if backChannel {
			sm.UpdateSessionData(obj.ID, err.Message+"!\n"+err.ErrorMessage, model.EnvVariables.SessionVariables.FinishedPersealBackChannel)
		}
		writeResponseMessage(w, obj, *err)
		return
	}

	log.Println("Current Persistence Object: ", obj)
	return
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

// ================================== METHODS CALLED AT THE END OF THE OPERATIONS ==============================

// Opens HTML that displays the message in the screen
func openExternalHTML(dto dto.PersistenceDTO, w http.ResponseWriter) {
	if dto.PDS == model.EnvVariables.Browser_PDS {
		// msToken for the perseal DataStore File Download EP, as a security measure
		tok, _ := sm.GenerateToken(model.EnvVariables.Perseal_Sender_Receiver, model.EnvVariables.Perseal_Sender_Receiver, dto.ID)
		dto.Response.MSTokenDownload = tok.AdditionalData
	}
	res := model.MarshallResponseToPrint(dto.Response)
	log.Println("Response Object: ", res)
	t, _ := template.ParseFiles("ui/message.html")
	w.WriteHeader(dto.Response.Code)
	t.Execute(w, dto.Response)

}

// Opens HTML that displays the following steps in the persistence methods (insert password, choose option, etc)
// They may redirect to other perseal EP's
func openInternalHTML(obj dto.PersistenceDTO, w http.ResponseWriter, filename string) {
	// msToken for the perseal ClientCallbackAddr redirection EP, as a security measure
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

	// If the error that caused the operation can be corrected by the user (wrong password, wrong file, etc), it will reload the HTML
	if dto.MenuOption != "" {
		openInternalHTML(dto, w, menuHTML)
	} else {
		// tok1 contains msToken with info about the perseal operation
		// tok2 is msToken for the perseal ClientCallbackAddr redirection EP, as a security measure
		var tok1, tok2 string

		// If the operation was successful
		if dto.Response.Code == http.StatusOK {
			tok1, tok2 = services.BuildDataOfMSToken(dto.ID, "OK", dto.ClientCallbackAddr)
			log.Println("Token contains OK message")
		} else {

			// If the failure was due to the msToken invalidation
			if dto.Response.ErrorMessage == model.Messages.NoMSTokenErrorMsg {
				dto.Response.MSToken = ""
			} else {

				// Any other failure in the perseal operation
				tok1, tok2 = services.BuildDataOfMSToken(dto.ID, "ERROR", dto.ClientCallbackAddr, "Failure! "+"\n"+dto.Response.Message+"\n"+dto.Response.ErrorMessage)
				log.Println("Token contains ERROR message")
			}
		}

		dto.Response.MSTokenRedirect = tok1
		dto.Response.MSToken = tok2
		if tok1 != "" && tok2 != "" {
			log.Println("Generated both tokens")
		}
		openExternalHTML(dto, w)
	}
}

// Response for the Back-Channel Operations. Generates Token with information of the operation and polls to ClientCallbackAddr
func writeBackChannelResponse(dto dto.PersistenceDTO, w http.ResponseWriter) {
	if dto.MenuOption != "BadQR" {
		w.WriteHeader(dto.Response.Code)
		w.Write([]byte(dto.Response.Message))
	}

	var tok string
	if dto.Response.Code == http.StatusOK {
		tok, _ = services.BuildDataOfMSToken(dto.ID, "OK", dto.ClientCallbackAddr)
		log.Println("Token contains OK message")
	} else {
		tok, _ = services.BuildDataOfMSToken(dto.ID, "ERROR", dto.ClientCallbackAddr, "Failure! "+"\n"+dto.Response.Message+"\n"+dto.Response.ErrorMessage)
		log.Println("Token contains ERROR message")
	}
	services.ClientCallbackAddrPost(tok, dto.ClientCallbackAddr)
}

// ================================== OTHER METHODS ==============================

// Generates QR code and presents it in HTML
func mobileQRCode(obj dto.PersistenceDTO, variables QRVariables, w http.ResponseWriter) {
	b, _ := json.Marshal(variables)
	var receiver string
	if strings.Contains(obj.ClientCallbackAddr, "/rm/response") {
		receiver = model.EnvVariables.RM_ID
	} else {
		receiver = model.EnvVariables.APGW_ID
	}

	tok1, _ := sm.GenerateToken(model.EnvVariables.Perseal_Sender_Receiver, receiver, obj.ID, string(b))

	//TODO: Contents should have URL, not just the token???
	qrCodeContents, _ := json.Marshal(tok1.AdditionalData)
	img, _ := qrcode.Encode(string(qrCodeContents), qrcode.Medium, 380)
	obj.Image = base64.StdEncoding.EncodeToString(img)

	if containsEmpty(variables.SessionId, variables.Method) {
		resp := model.BuildResponse(http.StatusInternalServerError, model.Messages.IncompleteQRCode)
		writeResponseMessage(w, obj, *resp)
		return
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

package sm

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/utils"
)

// SessionMngrResponse is the response to /sm/ValidateToken - it represents the current Users's session
type SessionMngrResponse struct {
	AdditionalData string      `json:"additionalData"`
	Code           string      `json:"code"` //OK, ERROR, NEW
	Error          string      `json:"error"`
	SessionData    SessionData `json:"sessionData"`
}

type SessionData struct {
	SessionID        string            `json:"sessionId"`
	SessionVariables map[string]string `json:"sessionVariables"`
}
type UpdateDataRequest struct {
	DataObject   string `json:"dataObject"`
	SessionId    string `json:"sessionId"`
	VariableName string `json:"variableName"`
}

type RequestParameters struct {
	Data      string `json:"data"`
	SessionId string `json:"sessionId"`
	Type      string `json:"type"`
	ID        string `json:"id"`
}

var (
	client http.Client
)

// OLD API

func GenerateToken(receiver, sender, sessionId string, data ...string) (tokenResp SessionMngrResponse, err *model.HTMLResponse) {
	u, _ := url.ParseRequestURI(model.EnvVariables.SMURLs.EndPoint)
	u.Path = model.EnvVariables.SMURLs.Generate_Token
	url := u.String()
	req, erro := http.NewRequest(http.MethodGet, url, nil)
	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedGenerateURL+u.Path, erro.Error())
		return
	}

	q := req.URL.Query()
	q.Add("sender", sender)
	q.Add("receiver", receiver)
	q.Add("sessionId", sessionId)
	if len(data) > 0 || data != nil {
		q.Add("data", data[0])
	}
	req.URL.RawQuery = q.Encode()

	tokenResp, err = smRequest(req, url)
	if err == nil {
		log.Println("MSToken Generated")
	}
	return
}

// ValidateToken - SessionManager function where the passed security token’s signature will be validated, as well as the validity as well as other validation measuresResponds by code: OK,
// sessionData.sessionId the sessionId used to gen. the jwt, and additionalData: extraData that were used to generate the jwt
func ValidateToken(token string) (smResp SessionMngrResponse, err *model.HTMLResponse) {
	u, _ := url.ParseRequestURI(model.EnvVariables.SMURLs.EndPoint)
	u.Path = model.EnvVariables.SMURLs.Validate_Token
	url := u.String()
	req, erro := http.NewRequest(http.MethodGet, url, nil)
	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedGenerateURL+u.Path, erro.Error())
		return
	}

	q := req.URL.Query()
	q.Add("token", token)
	req.URL.RawQuery = q.Encode()

	smResp, err = smRequest(req, url)
	if err == nil {
		log.Println("Valid MSToken")
	} else {
		log.Println("Invalid MSToken")
	}
	return smResp, err
}

// GetSessionData - SessionManager function where a variable or the whole session object is retrieved. Responds by code:OK, sessionData:{sessionId: the session, sessioVarialbes: map of variables,values}
func GetSessionData(sessionID string) (smResp SessionMngrResponse, err *model.HTMLResponse) {
	u, _ := url.ParseRequestURI(model.EnvVariables.SMURLs.EndPoint)
	u.Path = model.EnvVariables.SMURLs.Get_Session_Data
	url := u.String()
	req, erro := http.NewRequest(http.MethodGet, url, nil)

	q := req.URL.Query()
	q.Add(model.EnvVariables.SessionVariables.SessionId, sessionID)
	req.URL.RawQuery = q.Encode()

	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedGenerateURL+u.Path, erro.Error())
		return
	}

	smResp, err = smRequest(req, url)
	if err != nil {
		log.Println("Invalid ID. Could not fetch SessionData")
	}
	return
}

//Updates a Session Variable, by providind the sessionID, the new value of the variable and the the variable name
func UpdateSessionData(sessionId string, dataObject string, variableName string) (err *model.HTMLResponse) {
	u, _ := url.ParseRequestURI(model.EnvVariables.SMURLs.EndPoint)
	u.Path = model.EnvVariables.SMURLs.Update_Session_Data
	url := u.String()

	up := &UpdateDataRequest{
		SessionId:    sessionId,
		DataObject:   dataObject,
		VariableName: variableName,
	}
	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(up)

	req, erro := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBodyBytes.Bytes()))
	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedGenerateURL+u.Path, erro.Error())
		return
	}

	req, erro = utils.PrepareRequestHeaders(req, url)
	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedSignRequest, erro.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, erro := client.Do(req)
	if erro != nil {
		err = model.BuildResponse(http.StatusNotFound, model.Messages.FailedExecuteRequest, erro.Error())
	} else {
		log.Println("Updated Session Variable ", up.VariableName)
	}

	err = retryIfInternalServerError(req, resp)
	return
}

// ValidateSessionMngrResponse valites the fields in the received data in ValidateToken/GetSessionData
func ValidateSessionMngrResponse(smResp SessionMngrResponse, olderr *model.HTMLResponse) (err *model.HTMLResponse) {
	if smResp.Code == "ERROR" {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.SMRespError, smResp.Error)
	} else {
		err = olderr
	}
	return
}

// NEW API

func GenerateTokenWithPayload(receiver, sender, sessionId, data string) (smResp SessionMngrResponse, err *model.HTMLResponse) {
	//model.EnvVariables.SMURLs.EndPoint=http://vm.project-seal.eu:9090/sm
	u, _ := url.ParseRequestURI(model.EnvVariables.SMURLs.EndPoint)
	u.Path = model.EnvVariables.SMURLs.Generate_Token_With_Payload
	url := u.String()

	up := &RequestParameters{
		SessionId: sessionId,
		Data:      data,
	}
	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(up)

	req, erro := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBodyBytes.Bytes()))
	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedGenerateURL+u.Path, erro.Error())
		return
	}
	req.Header.Add("Content-Type", "application/json")

	q := req.URL.Query()
	q.Add("sender", sender)
	q.Add("receiver", receiver)
	q.Add("sessionId", sessionId)
	req.URL.RawQuery = q.Encode()

	smResp, err = smRequest(req, url)
	if err == nil {
		log.Println("Generate Token With Payload Operation Successful")
	}
	return
}

func StartSession(sessionId string) (smResp SessionMngrResponse, err *model.HTMLResponse) {
	//model.EnvVariables.SMURLs.EndPoint=http://vm.project-seal.eu:9090/sm
	u, _ := url.ParseRequestURI(model.EnvVariables.SMURLs.EndPoint)
	u.Path = model.EnvVariables.SMURLs.Start_Session
	url := u.String()

	up := &RequestParameters{
		SessionId: sessionId,
	}

	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(up)

	var req *http.Request
	var erro error

	req, erro = http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBodyBytes.Bytes()))

	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedGenerateURL+u.Path, erro.Error())
		return
	}
	req.Header.Add("Content-Type", "application/json")

	smResp, err = smRequest(req, url)
	if err == nil {
		log.Println("New Session Operation Successful")
	}
	return
}

func NewAdd(s RequestParameters) (smResp SessionMngrResponse, err *model.HTMLResponse) {
	//model.EnvVariables.SMURLs.EndPoint=http://vm.project-seal.eu:9090/sm
	u, _ := url.ParseRequestURI(model.EnvVariables.SMURLs.EndPoint)
	u.Path = model.EnvVariables.SMURLs.New_Add
	url := u.String()

	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(s)
	req, erro := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBodyBytes.Bytes())) // URL-encoded payload
	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedGenerateURL+u.Path, erro.Error())
		return
	}

	req.Header.Add("Content-Type", "application/json")
	log.Println(s)
	smResp, err = smRequest(req, url)
	if err == nil {
		log.Println("Added Entry to DataStore")
	}
	return

}

/*

// GetSessionData - SessionManager function where a variable or the whole session object is retrieved. Responds by code:OK, sessionData:{sessionId: the session, sessioVarialbes: map of variables,values}
func NewDelete(sessionId string, id ...string) (smResp SessionMngrResponse, err *model.HTMLResponse) {
	//model.EnvVariables.SMURLs.EndPoint=http://vm.project-seal.eu:9090/sm
	u, _ := url.ParseRequestURI(model.EnvVariables.SMURLs.EndPoint)
	u.Path = model.EnvVariables.SMURLs.New_Delete
	url := u.String()

	var reqBodyBytes *bytes.Buffer
	if len(id) > 0 || id != nil {
		up := &RequestParameters{
			SessionId: sessionId,
			ID:        id[0],
		}
		reqBodyBytes = new(bytes.Buffer)
		json.NewEncoder(reqBodyBytes).Encode(up)
	} else {
		up := &OverwriteData{
			SessionId: sessionId,
		}
		reqBodyBytes = new(bytes.Buffer)
		json.NewEncoder(reqBodyBytes).Encode(up)
	}

	req, erro := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBodyBytes.Bytes())) // URL-encoded payload
	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedGenerateURL+u.Path, erro.Error())
		return
	}

	req.Header.Add("Content-Type", "application/json")
	smResp, err = smRequest(req, url)
	if err == nil {
		log.Println("Delete Operation Successful")
	}
	return
}
*/
// GetSessionData - SessionManager function where a variable or the whole session object is retrieved. Responds by code:OK, sessionData:{sessionId: the session, sessioVarialbes: map of variables,values}
func NewSearch(sessionId string, typeValue ...string) (smResp SessionMngrResponse, err *model.HTMLResponse) {
	//model.EnvVariables.SMURLs.EndPoint=http://vm.project-seal.eu:9090/sm
	u, _ := url.ParseRequestURI(model.EnvVariables.SMURLs.EndPoint)
	u.Path = model.EnvVariables.SMURLs.New_Search
	url := u.String()

	up := &RequestParameters{
		Type:      "",
		SessionId: sessionId,
	}
	if len(typeValue) > 0 || typeValue != nil {
		up.Type = typeValue[0]
	}

	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(up)

	req, erro := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBodyBytes.Bytes()))
	req.Header.Add("Content-Type", "application/json")
	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedGenerateURL+u.Path, erro.Error())
		return
	}

	q := req.URL.Query()
	q.Add(model.EnvVariables.SessionVariables.SessionId, sessionId)

	req.URL.RawQuery = q.Encode()

	smResp, err = smRequest(req, url)
	if err == nil {
		log.Println("Search Operation Successful")
	}
	return
}

func smRequest(req *http.Request, url string) (smResp SessionMngrResponse, err *model.HTMLResponse) {

	req, erro := utils.PrepareRequestHeaders(req, url)
	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedSignRequest, erro.Error())
		return
	}

	resp, erro := client.Do(req)
	if erro != nil {
		err = model.BuildResponse(http.StatusNotFound, model.Messages.FailedExecuteRequest, erro.Error())
		return
	}
	err = retryIfInternalServerError(req, resp)
	if err != nil {
		return
	}

	body, erro := ioutil.ReadAll(resp.Body)
	if erro != nil {
		err = model.BuildResponse(http.StatusBadRequest, model.Messages.FailedReadResponse+req.URL.Path, erro.Error())
		return
	}

	var result interface{}
	erro = json.Unmarshal([]byte(body), &result)
	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedParseResponse+req.URL.Path, erro.Error())
		return
	}

	jsonM, erro := json.Marshal(result)
	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedGenerateSMResp+req.URL.Path, erro.Error())
		return
	}
	json.Unmarshal(jsonM, &smResp)
	err = ValidateSessionMngrResponse(smResp, err)
	return
}

func retryIfInternalServerError(req *http.Request, resp *http.Response) (err *model.HTMLResponse) {
	var erro bool
	if resp.StatusCode == http.StatusInternalServerError {
		erro = true
		for i := 0; i < 2; i++ {
			time.Sleep(2 * time.Second)
			client.Do(req)
			if resp.StatusCode != http.StatusInternalServerError {
				erro = false
				break
			}
		}
		if erro {
			err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedExecuteRequest, model.Messages.ISE)
		}
	}
	return
}

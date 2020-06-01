package sm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/utils"
)

/* validateToken
/sm/validateToken as per IN
get:
  tags:
	- SessionManager
  summary: 'The passed security token’s signature will be validated, as well as the validity as well as other validation measuresResponds by code: OK, sessionData.sessionId the sessionId used to gen. the jwt, and additionalData: extraData that were used to generate the jwt'
  operationId: validateTokenUsingGET
  produces:
	- application/json
  parameters:
	- name: token
	  in: query
	  description: token
	  required: true
	  type: string
  responses:
	'200':
	  description: OK
	  schema:
		$ref: '#/definitions/SessionMngrResponse'
	'401':
	  description: Unauthorized
	'403':
	  description: Forbidden
	'404':
	  description: Not Found
  deprecated: false


Response schema - application/json:
SessionMngrResponse:
	type: object
	properties:
	  additionalData:
		type: string
	  code:
		type: string
		enum:
		  - OK
		  - ERROR
		  - NEW
	  error:
		type: string
	  sessionData:
		$ref: '#/definitions/MngrSessionTO'
	title: SessionMngrResponse
*/

/*
 UpdateDataRequest:
    type: object
    properties:
      dataObject:
        type: string
      sessionId:
        type: string
      variableName:
        type: string
    title: UpdateDataRequest

*/

// SessionMngrResponse is the response to /sm/ValidateToken - it represents the current Users's session
type SessionMngrResponse struct {
	AdditionalData string `json:"additionalData"`
	Code           string `json:"code"` //OK, ERROR, NEW
	Error          string `json:"error"`
	SessionData    struct {
		SessionID        string            `json:"sessionId"`
		SessionVariables map[string]string `json:"sessionVariables"`
	} `json:"sessionData"`
}

type UpdateDataRequest struct {
	DataObject   string `json:"dataObject"`
	SessionId    string `json:"sessionId"`
	VariableName string `json:"variableName"`
}

type TokenResponse struct {
	Payload string `json:"payload"`
	Status  struct {
		Message string `json:"message"`
	}
}

var (
	client      http.Client
	CurrentUser chan SessionMngrResponse
)

func GenerateToken(data string, receiver string, sender string, sessionId string) (tokenResp SessionMngrResponse, err *model.DashboardResponse) {
	var url string
	if model.Local {
		url = "http://vm.project-seal.eu:9090/sm/generateToken?receiver=" + receiver + "&sender=" + sender + "&sessionId=" + sessionId
	} else {
		url = os.Getenv("SM_ENDPOINT") + "/generateToken?receiver=" + receiver + "&sender=" + sender + "&sessionId=" + sessionId
	}
	req, erro := http.NewRequest("GET", url, nil)

	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Couldn't Generate URL to Generate Token",
			ErrorMessage: erro.Error(),
		}
		return
	}
	req, erro = utils.PrepareRequestHeaders(req, url)

	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Couldn't Sign Request",
			ErrorMessage: erro.Error(),
		}
		return
	}

	fmt.Println(req.URL)
	resp, erro := client.Do(req)
	fmt.Println("\n", resp)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         404,
			Message:      "Couldn't Execute Request to Generate Token",
			ErrorMessage: erro.Error(),
		}
		return
	}

	body, erro := ioutil.ReadAll(resp.Body)

	if err != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Couldn't Read Response from Request to  Generate Token",
			ErrorMessage: erro.Error(),
		}
		return
	}

	var dat interface{}
	json.Unmarshal([]byte(body), &dat)
	fmt.Println("\n", dat)
	jsonM, erro := json.Marshal(dat)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Couldn't Generate JSON From Response Body of Generate Token",
			ErrorMessage: erro.Error(),
		}
		return
	}

	tokenResp = SessionMngrResponse{}
	json.Unmarshal(jsonM, &tokenResp)
	fmt.Println(tokenResp)
	return
}

// ValidateToken - SessionManager function where the passed security token’s signature will be validated, as well as the validity as well as other validation measuresResponds by code: OK,
// sessionData.sessionId the sessionId used to gen. the jwt, and additionalData: extraData that were used to generate the jwt
func ValidateToken(token string) (sessionId string, err *model.DashboardResponse) {
	var url string
	if model.Local {
		url = "http://vm.project-seal.eu:9090/sm/validateToken?token=" + token
	} else {
		url = os.Getenv("SM_ENDPOINT") + "/validateToken?token=" + token
	}
	req, erro := http.NewRequest("GET", url, nil)

	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Couldn't Generate URL to Validate Token",
			ErrorMessage: erro.Error(),
		}
		return
	}

	req, erro = utils.PrepareRequestHeaders(req, url)

	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Couldn't Sign Request",
			ErrorMessage: erro.Error(),
		}
		return
	}

	fmt.Println(req.URL)
	resp, erro := client.Do(req)
	fmt.Println(resp)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         404,
			Message:      "Couldn't Execute Request to Validate Token",
			ErrorMessage: erro.Error(),
		}
		return
	}

	body, erro := ioutil.ReadAll(resp.Body)

	if err != nil {
		err = &model.DashboardResponse{
			Code:         400,
			Message:      "Couldn't Read Response from Request to Validate Token",
			ErrorMessage: erro.Error(),
		}
		return
	}

	var data interface{}
	json.Unmarshal([]byte(body), &data)
	fmt.Println(data)
	jsonM, erro := json.Marshal(data)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Couldn't Generate JSON From Response Body of Validate Token",
			ErrorMessage: erro.Error(),
		}
		return
	}

	smResp := SessionMngrResponse{}
	json.Unmarshal(jsonM, &smResp)

	return smResp.SessionData.SessionID, ValidateSessionMngrResponse(smResp)
}

// GetSessionData - SessionManager function where a variable or the whole session object is retrieved. Responds by code:OK, sessionData:{sessionId: the session, sessioVarialbes: map of variables,values}
func GetSessionData(sessionID string, variableName string) (smResp SessionMngrResponse, err *model.DashboardResponse) {
	var url string
	if model.Local {
		url = "http://vm.project-seal.eu:9090/sm/getSessionData?sessionId=" + sessionID + "&variableName=" + variableName
	} else {
		url = os.Getenv("SM_ENDPOINT") + "/getSessionData?sessionId=" + sessionID + "&variableName=" + variableName
	}
	req, erro := http.NewRequest("GET", url, nil)

	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Couldn't Generate URL to Get Session Data",
			ErrorMessage: erro.Error(),
		}
		return
	}

	req, erro = utils.PrepareRequestHeaders(req, url)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Couldn't Sign Request",
			ErrorMessage: erro.Error(),
		}
		return
	}

	resp, erro := client.Do(req)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         404,
			Message:      "Couldn't Execute Request to Get Session Data",
			ErrorMessage: erro.Error(),
		}
		return
	}

	body, erro := ioutil.ReadAll(resp.Body)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         400,
			Message:      "Couldn't Read Response from Request to Get Session Data",
			ErrorMessage: erro.Error(),
		}
		return
	}

	var result interface{}
	erro = json.Unmarshal([]byte(body), &result)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Couldn't Parse Response Body from Get Session Data to Object",
			ErrorMessage: erro.Error(),
		}
		return
	}

	jsonM, erro := json.Marshal(result)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Couldn't Generate JSON From Object to SessionManagerResponse",
			ErrorMessage: erro.Error(),
		}
		return
	}
	json.Unmarshal(jsonM, &smResp)

	return
}

// ValidateSessionMngrResponse valites the fields in the received data in ValidateToken/GetSessionData
func ValidateSessionMngrResponse(smResp SessionMngrResponse) (err *model.DashboardResponse) {
	if smResp.Code == "ERROR" {
		err = &model.DashboardResponse{
			Code:         404,
			Message:      `ERROR" code in the received SessionData`,
			ErrorMessage: smResp.Error,
		}
	}
	return
}

//Updates a Session Variable, by providind the sessionID, the new value of the variable and the the variable name
func UpdateSessionData(sessionId string, dataObject string, variableName string) (body string, err *model.DashboardResponse) {
	var url string
	if model.Local {
		url = "http://vm.project-seal.eu:9090/sm/updateSessionData"
	} else {
		url = os.Getenv("SM_ENDPOINT") + "/updateSessionData"
	}
	up := &UpdateDataRequest{
		SessionId:    sessionId,
		DataObject:   dataObject,
		VariableName: variableName,
	}
	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(up)

	req, erro := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyBytes.Bytes()))
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Couldn't Generate URL to UpdataSessionData",
			ErrorMessage: erro.Error(),
		}
		return
	}

	req, erro = utils.PrepareRequestHeaders(req, url)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         500,
			Message:      "Couldn't Sign Request",
			ErrorMessage: erro.Error(),
		}
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, erro := client.Do(req)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         404,
			Message:      "Couldn't Execute Request to UpdataSessionData",
			ErrorMessage: erro.Error(),
		}
		return
	}
	defer resp.Body.Close()

	bodybytes, erro := ioutil.ReadAll(resp.Body)
	if erro != nil {
		err = &model.DashboardResponse{
			Code:         400,
			Message:      "Couldn't Read Response from Request to UpdataSessionData",
			ErrorMessage: erro.Error(),
		}
		return
	}
	body = string(bodybytes)
	return
}

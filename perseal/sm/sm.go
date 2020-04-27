package sm

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
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
	client http.Client
)

// ValidateToken - SessionManager function where the passed security token’s signature will be validated, as well as the validity as well as other validation measuresResponds by code: OK,
// sessionData.sessionId the sessionId used to gen. the jwt, and additionalData: extraData that were used to generate the jwt
func ValidateToken(token string) (smResp SessionMngrResponse, err error) {

	url := "http://vm.project-seal.eu:9090/sm/validateToken?token=" + token
	//url := os.Getenv("SM_ENDPOINT") + "/validateToken?token=" + token
	req, _ := http.NewRequest("GET", url, nil)
	req = prepareRequestHeaders(req, url)
	resp, _ := client.Do(req)
	body, err := ioutil.ReadAll(resp.Body)

	var data interface{}
	json.Unmarshal([]byte(body), &data)
	log.Println(data)
	jsonM, _ := json.Marshal(data)
	smResp = SessionMngrResponse{}
	json.Unmarshal(jsonM, &smResp)

	return smResp, ValidateSessionMngrResponse(smResp)
}

// GetSessionData - SessionManager function where a variable or the whole session object is retrieved. Responds by code:OK, sessionData:{sessionId: the session, sessioVarialbes: map of variables,values}
func GetSessionData(sessionID string, variableName string) (smResp SessionMngrResponse, err error) {

	//url := os.Getenv("SM_ENDPOINT") + "/getSessionData?sessionId=" + sessionID
	url := "http://vm.project-seal.eu:9090/sm/getSessionData?sessionId=" + sessionID + "&variableName=" + variableName
	req, _ := http.NewRequest("GET", url, nil)
	req = prepareRequestHeaders(req, url)
	resp, err := client.Do(req)
	if err != nil {
		return smResp, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return smResp, err
	}

	var teststring interface{}
	err = json.Unmarshal([]byte(body), &teststring)
	jsonM, _ := json.Marshal(teststring)
	smResp = SessionMngrResponse{}
	json.Unmarshal(jsonM, &smResp)

	return smResp, err
}

// ValidateSessionMngrResponse valites the fields in the received data in ValidateToken/GetSessionData
func ValidateSessionMngrResponse(smResp SessionMngrResponse) (err error) {
	if smResp.Code == "ERROR" {
		return errors.New(`"ERROR" code in the received SessionData: ` + smResp.Error)
	}
	return nil
}

func UpdateSessionData(sessionId string, dataObject string, variableName string) (body string, err error) {
	url := "http://vm.project-seal.eu:9090/sm/updateSessionData"
	up := &UpdateDataRequest{
		SessionId:    sessionId,
		DataObject:   dataObject,
		VariableName: variableName,
	}
	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(up)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyBytes.Bytes()))
	req = prepareRequestHeaders(req, url)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	bodybytes, _ := ioutil.ReadAll(resp.Body)
	body = string(bodybytes)
	return body, nil

}

package sm

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
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

// SessionMngrResponse is the response to /sm/ValidateToken - it represents the current Users's session
type SessionMngrResponse struct {
	AdditionalData string `json:"additionalData"`
	Code           string `json:"code"` //OK, ERROR, NEW
	Error          string `json:"error"`
	SessionData    struct {
		SessionID        string   `json:"sessionId"`
		SessionVariables struct { // Supposedly this is where the gmail Oauth tokens will be?
			GoogleDrive    string `json:"googleDrive"`
			OneDriveClient string `json:"onedriveclient"`
			OneDriveScopes string `json:"onedrivescopes"`
		} `json:"sessionVariables"`
	} `json:"sessionData"`
}

var (
	client http.Client
)

// ValidateToken - SessionManager function where the passed security token’s signature will be validated, as well as the validity as well as other validation measuresResponds by code: OK,
// sessionData.sessionId the sessionId used to gen. the jwt, and additionalData: extraData that were used to generate the jwt
func ValidateToken(token string) (smResp SessionMngrResponse, err error) {
	url := os.Getenv("SM_ENDPOINT") + "/validateToken?token=" + token
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
	err = json.Unmarshal([]byte(body), &smResp)
	return smResp, err
}

// GetSessionData - SessionManager function where a variable or the whole session object is retrieved. Responds by code:OK, sessionData:{sessionId: the session, sessioVarialbes: map of variables,values}
func GetSessionData(sessionID string, variableName string) (smResp SessionMngrResponse, err error) {
	// TODO handle variable name input
	url := os.Getenv("SM_ENDPOINT") + "/getSessionData?sessionId=" + sessionID
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
	err = json.Unmarshal([]byte(body), &smResp)

	return smResp, err
}

// ValidateSessionMngrResponse valites the fields in the received data in ValidateToken/GetSessionData
func ValidateSessionMngrResponse(smResp SessionMngrResponse, sessionToken string) (err error) {
	if smResp.Code == "ERROR" {
		return errors.New(`"ERROR" code in the received SessionData: ` + smResp.Error)
	}
	return nil
}

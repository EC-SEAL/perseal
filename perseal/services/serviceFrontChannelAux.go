package services

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/skip2/go-qrcode"
	"golang.org/x/oauth2"
)

// Generates MSToken to send to CCA with Success or Failure Data
func BuildDataOfMSToken(id, code, clientCallbackAddr string, message ...string) (string, string) {
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
	var receiver string
	if strings.Contains(clientCallbackAddr, "/rm/response") {
		receiver = model.EnvVariables.RM_ID
	} else if strings.Contains(clientCallbackAddr, "/per/retrieve") {
		receiver = model.EnvVariables.Perseal_Sender_Receiver
	} else {
		receiver = model.EnvVariables.APGW_ID
	}

	// TODO: Remove unecessary print
	log.Println("Receiver: " + receiver)
	tok1, err := sm.GenerateTokenWithPayload(model.EnvVariables.Perseal_Sender_Receiver, receiver, id, string(b))
	log.Println(tok1)
	if err != nil {
		return "", ""
	}
	tok2, err := sm.GenerateToken(model.EnvVariables.Perseal_Sender_Receiver, receiver, id)
	if err != nil {
		return "", ""
	}
	return tok1.AdditionalData, tok2.AdditionalData
}

func ClientCallbackAddrRedirect(token, clientCallbackAddr string) string {

	req, _ := http.NewRequest(http.MethodGet, clientCallbackAddr, nil)
	req.Header.Set("Content-Type", "multipart/form-data")
	q := req.URL.Query()
	q.Add("msToken", token)
	req.URL.RawQuery = q.Encode()

	log.Println("Redirect to: ", req.URL.String())
	return req.URL.String()
}

func QRCodePoll(id, op string) (respMethod string, obj dto.PersistenceDTO, err *model.HTMLResponse) {

	smResp, err := sm.GetSessionData(id)
	if err != nil {
		return
	}
	obj, err = dto.PersistenceFactory(id, smResp)
	if err != nil {
		return
	}

	log.Println("Current Persistence Object: ", obj)

	respMethod = smResp.SessionData.SessionVariables[model.EnvVariables.SessionVariables.FinishedPersealBackChannel]
	return
}

// Generates URL for user to select cloud account
func GetRedirectURL(dto dto.PersistenceDTO) (url string) {
	if dto.PDS == model.EnvVariables.Google_Drive_PDS && dto.GoogleAccessCreds.AccessToken == "" {
		url = getGoogleRedirectURL(dto.ID)
	} else if dto.PDS == model.EnvVariables.One_Drive_PDS && dto.OneDriveToken.AccessToken == "" {
		url = getOneDriveRedirectURL(dto.ID)
	}

	return
}

func UpdateTokenFromCode(dto dto.PersistenceDTO, code string) (dtoWithToken dto.PersistenceDTO, err *model.HTMLResponse) {
	var token *oauth2.Token
	dtoWithToken = dto
	if dto.PDS == model.EnvVariables.Google_Drive_PDS {
		token, err = updateNewGoogleDriveTokenFromCode(dto.ID, code)
		dtoWithToken.GoogleAccessCreds = *token
	} else if dto.PDS == model.EnvVariables.One_Drive_PDS {
		token, err = updateNewOneDriveTokenFromCode(dto.ID, code)
		dtoWithToken.OneDriveToken = *token
	}
	return
}

func GetCloudFileNames(dto dto.PersistenceDTO) (files, times []string, sizes []int64, err *model.HTMLResponse) {

	if dto.PDS == model.EnvVariables.Google_Drive_PDS {
		client := getGoogleDriveClient(dto.GoogleAccessCreds)
		var erro error
		filestmp, timestmp, sizestmp, erro := getGoogleDriveFiles(client)
		if erro != nil {
			err = model.BuildResponse(http.StatusNotFound, model.Messages.FailedGetCloudFiles+model.EnvVariables.Google_Drive_PDS, erro.Error())
			return
		}
		for i := range filestmp {
			if filestmp[i] != model.EnvVariables.DataStore_Folder_Name {
				files = append(files, filestmp[i])
				times = append(timestmp, timestmp[i])
				sizes = append(sizestmp, sizestmp[i])
			}
		}

		log.Println("Files Found: ", files)
	} else if dto.PDS == model.EnvVariables.One_Drive_PDS {
		var token *oauth2.Token
		token, err = checkOneDriveTokenExpiry(dto.OneDriveToken)
		if err != nil {
			return
		}
		resp, erro := getOneDriveItems(token)
		if erro != nil {
			err = model.BuildResponse(http.StatusNotFound, model.Messages.FailedGetCloudFiles+model.EnvVariables.One_Drive_PDS, erro.Error())
			return
		}
		for _, v := range resp.Values {
			if v.Name != model.EnvVariables.DataStore_Folder_Name {
				files = append(files, v.Name)
			}
		}
		log.Println("Files Found: ", resp.Values)
	}
	return
}

func GenerateCustomURL(dto dto.PersistenceDTO, r *http.Request) (token string) {
	//Defines Contents of QRCode/msToken
	contents := model.QRVariables{
		Method:    dto.Method,
		SessionId: dto.ID,
	}
	log.Println("Contents of QRCode/msToken: ", contents)

	// Generate msToken with the variables
	b, _ := json.Marshal(contents)
	token, _ = BuildDataOfMSToken(dto.ID, "OK", dto.ClientCallbackAddr, string(b))

	if strings.Contains(r.UserAgent(), "Android") || strings.Contains(r.UserAgent(), "iPhone") || strings.Contains(r.UserAgent(), "iPad") {
		log.Println("Mobile Device")
	} else {
		log.Println("Desktop Device")
	}

	// Sets session flag to signify back-channel hasn't finished yet
	sm.UpdateSessionData(dto.ID, "not finished", model.EnvVariables.SessionVariables.FinishedPersealBackChannel)

	return
}

func GenerateQRCode(obj dto.PersistenceDTO, variables model.QRVariables) (dto.PersistenceDTO, *model.HTMLResponse) {
	if containsEmpty(variables.SessionId, variables.Method) {
		resp := model.BuildResponse(http.StatusInternalServerError, model.Messages.IncompleteQRCode, obj.ID)
		return obj, resp
	}

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
	return obj, nil
}

func containsEmpty(stringArray ...string) bool {
	for _, s := range stringArray {
		if s == "" {
			return true
		}
	}
	return false
}

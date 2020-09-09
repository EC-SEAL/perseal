package services

import (
	"bytes"
	"encoding/json"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
	"golang.org/x/oauth2"
)

// Generates MSToken to send to CCA with Success or Failure Data
func BuildDataOfMSToken(id, code string, message ...string) (string, string) {
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
	tok1, err := sm.GenerateToken(model.EnvVariables.CCA_Sender, model.EnvVariables.Perseal_Sender_Receiver, id, string(b))
	log.Println(tok1.AdditionalData)
	if err != nil {
		return "", ""
	}
	tok2, err := sm.GenerateToken(model.EnvVariables.CCA_Sender, model.EnvVariables.Perseal_Sender_Receiver, id)
	if err != nil {
		return "", ""
	}
	return tok1.AdditionalData, tok2.AdditionalData
}

// Polls msToken to CCA
func ClientCallbackAddrPost(token, clientCallbackAddr string) {
	hc := http.Client{}
	b := bytes.Buffer{} // buffer to write the request payload into
	fw := multipart.NewWriter(&b)
	label, _ := fw.CreateFormField("msToken")
	label.Write([]byte(token))
	defer fw.Close()
	log.Println("POST to: ", clientCallbackAddr)
	req, _ := http.NewRequest(http.MethodPost, clientCallbackAddr, &b)
	req.Header.Set("Content-Type", fw.FormDataContentType())
	log.Println("Request: \n", req)
	log.Print("Result from ClientCallbackAddr: ")
	log.Println(hc.Do(req))
}

func QRCodePoll(id, op string) (respMethod string, obj dto.PersistenceDTO, err *model.HTMLResponse) {
	sm.UpdateSessionData(id, "not finished", model.EnvVariables.SessionVariables.FinishedPersealBackChannel)

	smResp, err := sm.GetSessionData(id)
	if err != nil {
		return
	}
	obj, err = dto.PersistenceFactory(id, smResp)
	if err != nil {
		return
	}

	log.Println("Current Persistence Object: ", obj)

	if op == model.EnvVariables.Load_Method {
		respMethod = model.Messages.LoadedDataStore
	} else if op == model.EnvVariables.Store_Method {
		respMethod = model.Messages.StoredDataStore
	}
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

func GetCloudFileNames(dto dto.PersistenceDTO) (files []string, err *model.HTMLResponse) {

	if dto.PDS == model.EnvVariables.Google_Drive_PDS {
		client := getGoogleDriveClient(dto.GoogleAccessCreds)
		var erro error
		files, erro = getGoogleDriveFiles(client)
		if erro != nil {
			err = model.BuildResponse(http.StatusNotFound, model.Messages.FailedGetCloudFiles+model.EnvVariables.Google_Drive_PDS, erro.Error())
			return
		}

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
			files = append(files, v.Name)
		}
		log.Println("Files Found: ", resp.Values)
	}
	return
}

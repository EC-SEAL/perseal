package dto

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
	"golang.org/x/oauth2"
)

type PersistenceDTO struct {
	//The SessionID
	ID string
	//The Microservice Token
	MSToken string
	//The PDS Location(googleDrive, oneDrive, Browser, Mobile)
	PDS string
	//The operation (Store or Load)
	Method string
	//The URL that the persistence redirects to when finishing its processes
	ClientCallbackAddr string

	//The password to encrypt or decrypt the DataStore
	Password string
	//The byte array of the selected local file, used in Browser Implementation
	LocalFileBytes []byte

	//The OAuth Tokens to access the cloud services
	GoogleAccessCreds oauth2.Token
	OneDriveToken     oauth2.Token

	//The Response that is written in the HTML page
	Response model.HTMLResponse

	//Option to be used in the Persistence Menus
	MenuOption string
	//The QRcode to be shown in the HTML, used in the Mobile implementation
	Image string

	//The DataStore Filename
	DataStoreFileName string

	Files FilesInfo

	UserError string
}

type FilesInfo struct {
	FileList []string
	TimeList []string
	SizeList []int64
}

// Builds Persistence DTO with its initial values
func PersistenceFactory(id string, smResp sm.SessionMngrResponse, method ...string) (dto PersistenceDTO, err *model.HTMLResponse) {

	client := smResp.SessionData.SessionVariables[model.EnvVariables.SessionVariables.ClientCallbackAddr]
	pds := smResp.SessionData.SessionVariables[model.EnvVariables.SessionVariables.PDS]

	dto = PersistenceDTO{
		ID:                 id,
		PDS:                pds,
		ClientCallbackAddr: client,
	}

	googleTokenBytes, oneDriveTokenBytes, err := getGoogleAndOneDriveTokens(dto, smResp)
	fmt.Println(string(googleTokenBytes))
	fmt.Println(string(oneDriveTokenBytes))
	if err != nil {
		fmt.Println(err)
		return
	}
	if string(googleTokenBytes) != "null" {
		log.Println("Found Google Token in Session")
	} else {
		log.Println("Not Found Google Token in Session")
	}

	if string(oneDriveTokenBytes) != "null" {
		log.Println("Found OneDrive Token in Session")
	} else {
		log.Println("Not Found OneDrive Token in Session")
	}

	json.Unmarshal(googleTokenBytes, &dto.GoogleAccessCreds)
	json.Unmarshal(oneDriveTokenBytes, &dto.OneDriveToken)

	if len(method) > 0 || method != nil {
		dto.Method = method[0]
	} else {
		dto.Method = smResp.SessionData.SessionVariables[model.EnvVariables.SessionVariables.CurrentMethod]
	}

	return
}

func getGoogleAndOneDriveTokens(dto PersistenceDTO, smResp sm.SessionMngrResponse) (googleTokenBytes, oneDriveTokenBytes []byte, err *model.HTMLResponse) {

	oneDriveToken := smResp.SessionData.SessionVariables[model.EnvVariables.SessionVariables.OneDriveToken]
	var token1 interface{}
	json.Unmarshal([]byte(oneDriveToken), &token1)
	oneDriveTokenBytes, erro := json.Marshal(token1)
	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedMarshall+model.EnvVariables.SessionVariables.OneDriveToken, erro.Error())
		return
	}

	googleDrive := smResp.SessionData.SessionVariables[model.EnvVariables.SessionVariables.GoogleDriveToken]
	var token2 interface{}
	json.Unmarshal([]byte(googleDrive), &token2)
	googleTokenBytes, erro = json.Marshal(token2)
	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedMarshall+model.EnvVariables.SessionVariables.GoogleDriveToken, erro.Error())
	}
	return
}

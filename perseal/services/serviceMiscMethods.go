package services

import (
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/model"
	"golang.org/x/oauth2"
)

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

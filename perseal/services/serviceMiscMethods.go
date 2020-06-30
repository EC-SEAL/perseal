package services

import (
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/externaldrive"
	"github.com/EC-SEAL/perseal/model"
	"golang.org/x/oauth2"
)

func GetRedirectURL(dto dto.PersistenceDTO) (url string) {

	if dto.PDS == "googleDrive" && dto.GoogleAccessCreds.AccessToken == "" {
		url = getGoogleRedirectURL(dto)
	} else if dto.PDS == "oneDrive" && dto.OneDriveToken.AccessToken == "" {
		url = getOneDriveRedirectURL(dto)
	}

	return
}

func UpdateTokenFromCode(dto dto.PersistenceDTO, code string) (dtoWithToken dto.PersistenceDTO, err *model.HTMLResponse) {
	var token *oauth2.Token
	dtoWithToken = dto
	log.Println(dtoWithToken)
	if dto.PDS == "googleDrive" {
		token, err = updateNewGoogleDriveTokenFromCode(dto.ID, code)
		log.Println(token)
		log.Println(err)
		dtoWithToken.GoogleAccessCreds = *token
	} else if dto.PDS == "oneDrive" {
		token, err = updateNewOneDriveTokenFromCode(dto.ID, code)
		dtoWithToken.OneDriveToken = *token
	}
	return
}

func GetCloudFileNames(dto dto.PersistenceDTO) (files []string, err *model.HTMLResponse) {

	if dto.PDS == "googleDrive" {
		var client *http.Client
		client = getGoogleDriveClient(dto.GoogleAccessCreds)
		var erro error
		files, erro = externaldrive.GetGoogleDriveFiles(client)
		if erro != nil {
			err = &model.HTMLResponse{
				Code:         404,
				Message:      "Could not Get GoogleDrive Files",
				ErrorMessage: erro.Error(),
			}
			return
		}

	} else if dto.PDS == "oneDrive" {
		var token *oauth2.Token
		token, err = checkOneDriveTokenExpiry(dto)
		if err != nil {
			return
		}
		resp, erro := externaldrive.GetOneDriveItems(token, "SEAL")
		if erro != nil {
			err = &model.HTMLResponse{
				Code:         404,
				Message:      "Couldn't Get One Drive Items",
				ErrorMessage: erro.Error(),
			}
			return
		}
		for _, v := range resp.Values {
			files = append(files, v.Name)
		}
		log.Println(resp.Values)
	}
	return
}

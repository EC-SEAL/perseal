package services

import (
	"encoding/json"
	"log"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/externaldrive"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/EC-SEAL/perseal/utils"
)

func GetRedirectURL(dto dto.PersistenceDTO) (url string, err *model.HTMLResponse) {

	if dto.PDS == "googleDrive" && dto.GoogleAccessCreds == "" {
		url, err = getGoogleRedirectURL(dto)
	} else if dto.PDS == "oneDrive" && dto.OneDriveToken.AccessToken == "" {
		url, err = getOneDriveRedirectURL(dto)
	}

	return
}

// Decrypts dataStore and loads it into session
func DecryptAndMarshallDataStore(dataStore *externaldrive.DataStore, dto dto.PersistenceDTO) (err *model.HTMLResponse) {

	erro := dataStore.Decrypt(dto.Password)
	log.Println(erro)
	dataStore.EncryptedData = ""
	if erro != nil {
		err = &model.HTMLResponse{
			Code:    500,
			Message: "Couldn't Decrypt DataStore. Check your password",
		}
		json.MarshalIndent(err, "", "\t")
		return
	}

	jsonM, erro := json.Marshal(dataStore)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         500,
			Message:      "Couldn't Parse Response Body from DataStore to Object",
			ErrorMessage: erro.Error(),
		}
		json.MarshalIndent(err, "", "\t")
		return
	}

	_, err = sm.UpdateSessionData(dto.ID, string(jsonM), "dataStore")

	return
}

func ValidateSignature(encrypted string, sigToValidate string) bool {
	sig, err := utils.GetSignature(encrypted)
	if err != nil {
		return false
	}

	log.Println("sig: ", sig)
	log.Println("toValidate: ", sigToValidate)
	if sig != sigToValidate {
		return false
	}
	log.Println("Validated")
	return true
}

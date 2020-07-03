package services

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/externaldrive"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/EC-SEAL/perseal/utils"
)

// Setup a persistence mechanism and load a secure storage into session.
func PersistenceLoad(dto dto.PersistenceDTO) (response, err *model.HTMLResponse) {
	log.Println("persistanceLoad")

	ds := &externaldrive.DataStore{}

	// Initialize Variables
	if dto.PDS == "googleDrive" || dto.PDS == "oneDrive" {
		ds, err = fetchCloudDataStore(dto, "datastore.seal")
	} else if dto.PDS == "Browser" {
		ds, err = readLocalFileDataStore(dto.LocalFileBytes)
	}

	if err != nil {
		return
	}

	// Validates signature of DataStore

	err = signAndDecryptDataStore(ds, dto)
	if err != nil {
		return
	}

	log.Println("Decrypted DataStore: ", ds)
	response = &model.HTMLResponse{
		Code:               200,
		Message:            "Loaded DataStore " + ds.ID,
		ClientCallbackAddr: dto.ClientCallbackAddr,
	}

	return
}

// UC 1.06 - Stores and Loads Datastore
func PersistenceStoreAndLoad(dto dto.PersistenceDTO) (response, err *model.HTMLResponse) {
	ds, err := storeCloudData(dto, "datastore.seal")
	if err != nil {
		return
	}

	err = signAndDecryptDataStore(ds, dto)

	log.Println("Decrypted DataStore: ", ds)
	if err != nil {
		return
	}

	response = &model.HTMLResponse{
		Code:               200,
		Message:            "Loaded DataStore " + ds.ID,
		ClientCallbackAddr: dto.ClientCallbackAddr,
	}
	return
}

func BackChannelDecryption(dto dto.PersistenceDTO, dataSstr string) (response, err *model.HTMLResponse) {
	dataStore, err := readLocalFileDataStore([]byte(dataSstr))
	if err != nil {
		return
	}

	err = signAndDecryptDataStore(dataStore, dto)
	if err != nil {
		return
	}
	log.Println("Decrypted DataStore: ", dataStore)

	response = &model.HTMLResponse{
		Code:               200,
		Message:            "Loaded DataStore " + dataStore.ID,
		ClientCallbackAddr: dto.ClientCallbackAddr,
	}
	return
}

// Service Method to Fetch the DataStore according to the PDS variable
func fetchCloudDataStore(dto dto.PersistenceDTO, filename string) (dataStore *externaldrive.DataStore, err *model.HTMLResponse) {
	var file *http.Response

	if dto.PDS == "googleDrive" {
		file, err = loadSessionDataGoogleDrive(dto, filename)
		if err != nil {
			return
		}
	} else if dto.PDS == "oneDrive" {
		file, err = loadSessionDataOneDrive(dto, filename)
		if err != nil {
			return
		}
	}
	log.Println("Managed to Fetch the DataStore File")
	body, erro := ioutil.ReadAll(file.Body)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:    500,
			Message: "Error Reading Responses",
		}
		return
	}
	dataStore, err = readCloudFileDataStore(body)
	log.Println("DataStore ID found: ", dataStore.ID)
	return
}

func validateSignature(encrypted string, sigToValidate string) bool {
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

// Decrypts dataStore and loads it into session
func signAndDecryptDataStore(dataStore *externaldrive.DataStore, dto dto.PersistenceDTO) (err *model.HTMLResponse) {
	log.Println(dataStore)
	if !validateSignature(dataStore.EncryptedData, dataStore.Signature) {
		err = &model.HTMLResponse{
			Code:    500,
			Message: "Error Validating Signature",
		}
		return
	}

	erro := dataStore.Decrypt(dto.Password)
	dataStore.EncryptedData = ""
	if erro != nil {
		err = &model.HTMLResponse{
			Code:    500,
			Message: "Couldn't Decrypt DataStore. Check your password",
		}
		json.MarshalIndent(err, "", "\t")
		return
	}

	jsonM, err := marshallDataStore(dataStore, dto)
	if err != nil {
		return
	}
	sm.UpdateSessionData(dto.ID, string(jsonM), "dataStore")
	return
}

func marshallDataStore(dataStore *externaldrive.DataStore, dto dto.PersistenceDTO) (jsonM []byte, err *model.HTMLResponse) {
	jsonM, erro := json.Marshal(dataStore)
	log.Println(erro)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:         500,
			Message:      "Couldn't Parse Response Body from DataStore to Object",
			ErrorMessage: erro.Error(),
		}
		json.MarshalIndent(err, "", "\t")
	}
	return
}

func readCloudFileDataStore(dataSstr []byte) (dataStore *externaldrive.DataStore, err *model.HTMLResponse) {
	var v interface{}
	json.Unmarshal(dataSstr, &v)
	jsonM, _ := json.Marshal(v)
	erro := json.Unmarshal(jsonM, &dataStore)
	log.Println(dataStore)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:    400,
			Message: "Bad Structure of DataStore",
		}
	}
	return

}

func readLocalFileDataStore(dataSstr []byte) (dataStore *externaldrive.DataStore, err *model.HTMLResponse) {
	var v string
	json.Unmarshal(dataSstr, &v)
	erro := json.Unmarshal([]byte(v), &dataStore)
	log.Println("DataStore ID found: ", dataStore.ID)
	if erro != nil {
		err = &model.HTMLResponse{
			Code:    400,
			Message: "The File must contain only a valid DataStore",
		}
	}
	return

}

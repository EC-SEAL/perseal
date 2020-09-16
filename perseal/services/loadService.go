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
	ds := &externaldrive.DataStore{}
	// Initialize Variables
	if dto.PDS == model.EnvVariables.Google_Drive_PDS || dto.PDS == model.EnvVariables.One_Drive_PDS {
		ds, err = fetchCloudDataStore(dto, model.EnvVariables.DataStore_File_Name)
	} else if dto.PDS == model.EnvVariables.Browser_PDS {
		ds, err = readLocalFileDataStore(dto.LocalFileBytes)
	}

	if err != nil {
		return
	}

	// Validates signature of DataStore
	err = validateSignAndDecryptDataStore(ds, dto)
	if err != nil {
		return
	}
	b, _ := json.MarshalIndent(ds, "", "\t")
	log.Println("Decrypted DataStore: ", string(b))

	response = model.BuildResponse(http.StatusOK, model.Messages.LoadedDataStore+ds.ID)
	return
}

// Stores and Loads Datastore
func PersistenceStoreAndLoad(dto dto.PersistenceDTO) (response, err *model.HTMLResponse) {
	var ds *externaldrive.DataStore
	if dto.PDS == model.EnvVariables.Google_Drive_PDS || dto.PDS == model.EnvVariables.One_Drive_PDS {
		ds, err = storeCloudData(dto)
	} else if dto.PDS == model.EnvVariables.Browser_PDS {
		var erro error
		ds, erro = externaldrive.StoreSessionData(dto)
		if erro != nil {
			err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedEncryption, erro.Error())
			return
		}
	}
	if err != nil {
		return
	}
	b, _ := json.MarshalIndent(ds, "", "\t")

	log.Println("Stored DataStore", string(b))
	err = validateSignAndDecryptDataStore(ds, dto)
	b, _ = json.MarshalIndent(ds, "", "\t")
	log.Println("Decrypted DataStore: ", string(b))
	if err != nil {
		return
	}

	response = model.BuildResponse(http.StatusOK, model.Messages.LoadedDataStore+ds.ID)

	if dto.PDS == model.EnvVariables.Browser_PDS {
		data, _ := json.Marshal(ds)
		response.DataStore = string(data)
	}
	return
}

func BackChannelDecryption(dto dto.PersistenceDTO, dataSstr string) (response, err *model.HTMLResponse) {
	dataStore, err := readLocalFileDataStore([]byte(dataSstr))
	if err != nil {
		return
	}

	err = validateSignAndDecryptDataStore(dataStore, dto)
	if err != nil {
		return
	}
	b, _ := json.MarshalIndent(dataStore, "", "\t")
	data, _ := json.Marshal(string(b))

	response = model.BuildResponse(http.StatusOK, model.Messages.LoadedDataStore+dataStore.ID)
	//response.ClientCallbackAddr = dto.ClientCallbackAddr
	response.DataStore = string(data)
	return
}

// Service Method to Fetch the DataStore according to the PDS variable
func fetchCloudDataStore(dto dto.PersistenceDTO, filename string) (dataStore *externaldrive.DataStore, err *model.HTMLResponse) {
	var file *http.Response

	if dto.PDS == model.EnvVariables.Google_Drive_PDS {
		file, err = loadSessionDataGoogleDrive(dto, filename)
		if err != nil {
			return
		}
	} else if dto.PDS == model.EnvVariables.One_Drive_PDS {
		file, err = loadSessionDataOneDrive(dto, filename)
		if err != nil {
			return
		}
	}
	log.Println("Managed to Fetch the DataStore File")
	body, erro := ioutil.ReadAll(file.Body)
	if erro != nil {
		err = model.BuildResponse(http.StatusNotFound, model.Messages.FileContentsNotFound, erro.Error())
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
	if sig != sigToValidate {
		return false
	}
	log.Println("Validated")
	return true
}

// Decrypts dataStore and loads it into session
func validateSignAndDecryptDataStore(dataStore *externaldrive.DataStore, dto dto.PersistenceDTO) (err *model.HTMLResponse) {
	if !validateSignature(dataStore.EncryptedData, dataStore.Signature) {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.InvalidSignature)
		return
	}

	erro := dataStore.Decrypt(dto.Password)

	tmp := dataStore.EncryptedData
	dataStore.EncryptedData = ""
	if erro != nil {
		err = model.BuildResponse(http.StatusBadRequest, model.Messages.InvalidPassword, erro.Error())
		err.FailedInput = "Password"
		log.Println("Wrong Password")
		return
	}

	log.Println(sm.NewSearch(dto.ID))
	sm.NewDelete(dto.ID)

	var t2 []sm.NewUpdateDataRequest
	json.Unmarshal([]byte(dataStore.ClearData.(string)), &t2)
	for _, element := range t2 {
		element.SessionId = dto.ID
		log.Println("Data: ", element.Data)
		if element.Data != "" {
			sm.NewAdd(element)
		}
	}

	smResp, _ := sm.NewSearch(dto.ID)
	log.Println(smResp.AdditionalData)
	dataStore.ClearData = smResp.AdditionalData
	dataStore.EncryptedData = tmp
	return
}

func marshallDataStore(dataStore *externaldrive.DataStore, dto dto.PersistenceDTO) (jsonM []byte, err *model.HTMLResponse) {
	jsonM, erro := json.Marshal(dataStore)
	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, model.Messages.FailedParseResponse+"DataStore", erro.Error())
		json.MarshalIndent(err, "", "\t")
	}
	return
}

func readCloudFileDataStore(dataSstr []byte) (dataStore *externaldrive.DataStore, err *model.HTMLResponse) {
	var v interface{}
	json.Unmarshal(dataSstr, &v)
	jsonM, _ := json.Marshal(v)
	erro := json.Unmarshal(jsonM, &dataStore)
	if erro != nil {
		err = model.BuildResponse(http.StatusBadRequest, model.Messages.InvalidDataStore, erro.Error())
	}
	return

}

func readLocalFileDataStore(dataSstr []byte) (dataStore *externaldrive.DataStore, err *model.HTMLResponse) {
	var v string
	json.Unmarshal(dataSstr, &v)
	erro := json.Unmarshal([]byte(v), &dataStore)
	if erro != nil {
		err = model.BuildResponse(http.StatusBadRequest, model.Messages.FileContainsDataStore, erro.Error())
		err.FailedInput = "File"
		log.Println("File Does Not Contain Valid DataStore")
		return
	}
	log.Println("DataStore ID found: ", dataStore.ID)
	return

}

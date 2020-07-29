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
	if dto.PDS == model.EnvVariables.Google_Drive_PDS || dto.PDS == model.EnvVariables.One_Drive_PDS {
		ds, err = fetchCloudDataStore(dto, model.EnvVariables.DataStore_File_Name)
	} else if dto.PDS == model.EnvVariables.Browser_PDS {
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
	b, _ := json.MarshalIndent(ds, "", "\t")
	log.Println("Decrypted DataStore: ", string(b))

	log.Println(sm.NewSearch(dto.ID))
	response = &model.HTMLResponse{
		Code:               200,
		Message:            "Loaded DataStore " + ds.ID,
		ClientCallbackAddr: dto.ClientCallbackAddr,
	}

	return
}

// UC 1.06 - Stores and Loads Datastore
func PersistenceStoreAndLoad(dto dto.PersistenceDTO) (response, err *model.HTMLResponse) {
	response = &model.HTMLResponse{}
	var ds *externaldrive.DataStore
	if dto.PDS == model.EnvVariables.Google_Drive_PDS || dto.PDS == model.EnvVariables.One_Drive_PDS {
		ds, err = storeCloudData(dto)
	} else if dto.PDS == model.EnvVariables.Browser_PDS {
		var erro error
		ds, erro = externaldrive.StoreSessionData(dto)
		if erro != nil {
			err = &model.HTMLResponse{
				Code:         500,
				Message:      "Encryption Failed",
				ErrorMessage: erro.Error(),
			}
			return
		}
		data, _ := json.Marshal(ds)
		response.DataStore = string(data)
		response.MSToken, err = utils.GenerateTokenAPI(dto.PDS, dto.ID)
	}
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Stored DataStore")

	err = signAndDecryptDataStore(ds, dto)

	log.Println(sm.NewSearch(dto.ID))

	b, _ := json.MarshalIndent(ds, "", "\t")
	log.Println("Decrypted DataStore: ", string(b))
	if err != nil {
		return
	}

	response.Code = 200
	response.Message = "Loaded DataStore " + ds.ID
	response.ClientCallbackAddr = dto.ClientCallbackAddr

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

	log.Println(sm.NewSearch(dto.ID))
	data, _ := json.Marshal(dataStore)

	response = &model.HTMLResponse{
		Code:               200,
		Message:            "Loaded DataStore " + dataStore.ID,
		ClientCallbackAddr: dto.ClientCallbackAddr,
		DataStore:          string(data),
	}
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
		err = &model.HTMLResponse{
			Code:    500,
			Message: "Error Reading Responses",
		}
		return
	}
	dataStore, err = readCloudFileDataStore(body)
	log.Println("DataStore ID found: ", dataStore.ID)
	log.Println("DataStore ClearData: ", dataStore.ClearData)
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

	log.Println(dataStore)
	sm.NewDelete(dto.ID)
	sm.NewAdd(dto.ID, dataStore.ClearData.(string), "dataSet")
	log.Println(sm.NewSearch(dto.ID))
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

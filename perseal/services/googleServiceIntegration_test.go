package services

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"testing"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/EC-SEAL/perseal/utils"
	"github.com/joho/godotenv"
)

var (
	id string
)

func InitIntegration(platform string) dto.PersistenceDTO {
	godotenv.Load("../.env")
	model.SetEnvVariables()
	tokenResp, _ := utils.StartSession("")
	id = tokenResp.Payload
	msToken, _ := utils.GenerateTokenAPI(platform, id)

	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(msToken)
	http.Get("http://localhost:8082/per/store?msToken=" + msToken)

	//simulate google login redirect
	sm.UpdateSessionData(id, model.EnvVariables.Store_Method, model.EnvVariables.SessionVariables.CurrentMethod)
	variables := map[string]string{
		"PDS":       platform,
		"dataStore": "{\"id\":\"DS_3a342b23-8b46-44ec-bb06-a03042135a5e\",\"encryptedData\":null,\"signature\":\"this is the signature\",\"signatureAlgorithm\":\"this is the signature algorithm\",\"encryptionAlgorithm\":\"this is the encryption algorithm\",\"clearData\":null}",
	}

	session := sm.SessionMngrResponse{}
	sessionData := session.SessionData
	sessionData.SessionID = id
	sessionData.SessionVariables = variables
	session.SessionData = sessionData

	obj, _ := dto.PersistenceFactory(id, session)

	sm.NewAdd(obj.ID, "this is linkRequest", "linkRequest")
	url := GetRedirectURL(obj)
	if url != "" {
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	}
	return obj
}

func preCloudConfig(obj dto.PersistenceDTO, smResp sm.SessionMngrResponse, password string) dto.PersistenceDTO {
	obj, _ = dto.PersistenceFactory(obj.ID, smResp, obj.Method)
	obj.Password = password
	return obj
}

func TestGoogleService(t *testing.T) {

	obj := InitIntegration("googleDrive")

	smResp, _ := sm.GetSessionData(obj.ID)

	// Test Correct GoogleDrive Store
	obj = preCloudConfig(obj, smResp, "qwerty")
	ds, err := storeCloudData(obj)
	log.Println(ds)
	if err != nil {
		t.Error("Thrown error, got: ", err)
	}

	// Test Incorrect GoogleDrive Store
	obj = preCloudConfig(obj, smResp, "qwerty")
	obj.GoogleAccessCreds.AccessToken += "123"
	ds, err = storeCloudData(obj)
	log.Println(ds)
	if err == nil {
		t.Error("Should have thrown error")
	}

	obj = preCloudConfig(obj, smResp, "qwerty")

	// Test Correct Load GoogleDrive Store
	ds, err = fetchCloudDataStore(obj, "datastore.seal")
	if err != nil {
		t.Error("Thrown error, got: ", err)
	}
	log.Println(ds)

	// Test Incorrect Load GoogleDrive Store
	ds, err = fetchCloudDataStore(obj, "datastorewrong.seal")
	if err == nil {
		t.Error("Should have thrown error")
	}

	// Test Get Cloud Files
	files, err := GetCloudFileNames(obj)
	if err != nil {
		t.Error("Thrown error, got: ", err)
	}
	if len(files) == 0 {
		t.Error("no files found")
	}

	// Test Get Cloud Files No GoogleCreds
	obj.GoogleAccessCreds.AccessToken = "1234"
	files, err = GetCloudFileNames(obj)
	if err == nil {
		t.Error("Should have thrown error")
	}

	obj = preCloudConfig(obj, smResp, "qwerty")
	_, erro := PersistenceStore(obj)
	log.Println(erro)
	if erro != nil {
		t.Error("Thrown error, got: ", err)
	}

	obj = preCloudConfig(obj, smResp, "qwerty")
	_, erro = PersistenceLoad(obj)
	log.Println(erro)
	if erro != nil {
		t.Error("Thrown error, got: ", err)
	}

	obj = preCloudConfig(obj, smResp, "qwerty")
	_, erro = PersistenceStoreAndLoad(obj)
	log.Println(erro)
	if erro != nil {
		t.Error("Thrown error, got: ", err)
	}

}

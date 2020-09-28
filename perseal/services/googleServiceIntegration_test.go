package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"testing"
	"time"

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
	var s sm.NewUpdateDataRequest = sm.NewUpdateDataRequest{
		SessionId: model.TestUser,
		Type:      "linkRequest",
		Data:      "this is",
	}
	sm.NewAdd(s)
	url := GetRedirectURL(obj)
	if url != "" {
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	}
	time.Sleep(4 * time.Second)
	return obj
}

func preCloudConfig(obj dto.PersistenceDTO, smResp sm.SessionMngrResponse, password string) dto.PersistenceDTO {
	obj, _ = dto.PersistenceFactory(obj.ID, smResp, obj.Method)
	obj.Password = password
	return obj
}

func TestGoogleService(t *testing.T) {

	var passed = "=================PASSED==============="
	var failed = "=================FAILED==============="

	obj := InitIntegration("googleDrive")

	smResp, _ := sm.GetSessionData(obj.ID)

	fmt.Println("\n=================Correct GoogleDrive Store====================")
	obj = preCloudConfig(obj, smResp, "qwerty")
	log.Println(obj)
	ds, err := storeCloudData(obj)
	if err != nil {
		fmt.Println(failed)
		t.Error("Thrown error, got: ", err)
	} else {
		fmt.Println(passed)
	}

	fmt.Println("\n=================Incorrect GoogleDrive Store - Bad Access Token====================")
	obj = preCloudConfig(obj, smResp, "qwerty")
	obj.GoogleAccessCreds.AccessToken += "123"
	ds, err = storeCloudData(obj)
	log.Println(ds)
	if err == nil {
		fmt.Println(failed)
		t.Error("Should have thrown error")
	} else {
		fmt.Println(passed)
	}

	obj = preCloudConfig(obj, smResp, "qwerty")

	fmt.Println("\n=================Correct Load GoogleDrive Store====================")
	ds, err = fetchCloudDataStore(obj)
	if err != nil {
		fmt.Println(failed)
		t.Error("Thrown error, got: ", err)
	} else {
		fmt.Println(passed)
	}

	fmt.Println("\n=================Incorrect Load GoogleDrive Store - Bad Filename====================")
	ds, err = fetchCloudDataStore(obj)
	if err == nil {
		fmt.Println(failed)
		t.Error("Should have thrown error")
	} else {
		fmt.Println(passed)
	}

	fmt.Println("\n=================Get Cloud Files====================")
	files, _, _, err := GetCloudFileNames(obj)
	if err != nil {
		fmt.Println(failed)
		t.Error("Thrown error, got: ", err)
	}
	if len(files) == 0 {
		fmt.Println(failed)
		t.Error("no files found")
	} else {
		fmt.Println(passed)
	}

	fmt.Println("\n=================Get Cloud Files - No Google Drive Creds====================")
	obj.GoogleAccessCreds.AccessToken = "1234"
	files, _, _, err = GetCloudFileNames(obj)
	if err == nil {
		fmt.Println(failed)
		t.Error("Should have thrown error")
	} else {
		fmt.Println(passed)
	}

	fmt.Println("\n=================Correct Persistence Store====================")
	obj = preCloudConfig(obj, smResp, "qwerty")
	_, erro := PersistenceStore(obj)
	if erro != nil {
		fmt.Println(failed)
		t.Error("Thrown error, got: ", err)
	} else {
		fmt.Println(passed)
	}

	fmt.Println("\n=================Correct Persistence Load====================")
	obj = preCloudConfig(obj, smResp, "qwerty")
	_, erro = PersistenceLoad(obj)
	if erro != nil {
		fmt.Println(failed)
		t.Error("Thrown error, got: ", err)
	} else {
		fmt.Println(passed)
	}

	fmt.Println("\n=================Correct Persistence Store And Load====================")
	obj = preCloudConfig(obj, smResp, "qwerty")
	_, erro = PersistenceStoreAndLoad(obj)
	if erro != nil {
		fmt.Println(failed)
		t.Error("Thrown error, got: ", err)
	} else {
		fmt.Println(passed)
	}

}

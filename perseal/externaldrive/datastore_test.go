package externaldrive

import (
	"fmt"
	"testing"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/joho/godotenv"
)

//go test -v -coverpkg=./... -coverprofile=profile.cov ./...
//go tool cover -html profile

var mockStringQwertyDataStore = "{\"id\":\"DS_3a342b23-8b46-44ec-bb06-a03042135a5e\",\"encryptedData\":\"xRbYPkPKxU8ikEmAKBp7CyG4h8DdybKd7K-jOMvFBe-Gw0fL0_sUAAyDL7wgNTDOxPcU0vlzniPCrIVhDPluVhcvxLZbqJviISLXqyiVucuc8C6uQvTI37MBpfuwuw==\",\"signature\":\"SkwdK6HW4jniY6Uw4j102yI-uVyVMe8G22kUT_j1GiUmOWRYnIjQ73RuX23hfL4UF1CVlgFtTiBSTKX_y5O8auPTD-o-AVJDNnOtXlhXn4xXqdV9zHbfrpXq2OtjtVmaw6uX7LeZrS64Nk5Ey0bFLCxomHfy7UakclsIWQ1_HX_Jgc6rcT1WVYJm8p8dp4JKxPRJR-GhdhoRMV14Jp7C8dG76_rDk25N0J8ggIuNs-wews0NTde7kmiE3K_9hpoRwbo5S5-vhLbNCSKaHzzZF7I1o8TEobwB79t9rvf4vVkHr59vsxeyXsctLv8DZwDt-guDx6zFCWqKTxtsHLkeLA==\",\"signatureAlgorithm\":\"rsa-sha256\",\"encryptionAlgorithm\":\"aes-cfb\"}"

func Init(pds string) dto.PersistenceDTO {
	godotenv.Load("../.env")
	model.SetEnvVariables()
	mockID := "123"
	variables := map[string]string{
		"PDS":       pds,
		"dataStore": "{\"id\":\"DS_3a342b23-8b46-44ec-bb06-a03042135a5e\",\"encryptedData\":null,\"signature\":\"this is the signature\",\"signatureAlgorithm\":\"this is the signature algorithm\",\"encryptionAlgorithm\":\"this is the encryption algorithm\",\"clearData\":null}",
	}

	session := sm.SessionMngrResponse{}
	sessionData := session.SessionData
	sessionData.SessionID = mockID
	sessionData.SessionVariables = variables
	session.SessionData = sessionData

	obj, _ := dto.PersistenceFactory(mockID, session)
	return obj
}

func TestDataStore(t *testing.T) {

	var passed = "=================PASSED==============="
	var failed = "=================FAILED==============="

	obj := Init("googleDrive")

	fmt.Println("\n=================New Store Session Data With Password====================")
	obj.Password = "qwerty"
	ds, err := StoreSessionData(obj)
	if err != nil {
		fmt.Println(failed)
		t.Error("Error occurred, got: ", err)
	}
	fmt.Println(passed)

	fmt.Println("\n=================Uploading Blob====================")
	//Test UploadingBlob With EncryptedData
	_, err = ds.UploadingBlob()
	if err != nil {
		fmt.Println(failed)
		t.Error("Error occurred, got: ", err)
	}
	fmt.Println(passed)

	fmt.Println("\n=================Uploading Blob With Encrypted Data====================")
	tmp := ds.EncryptedData
	ds.EncryptedData = ""
	_, err = ds.UploadingBlob()
	if err != nil {
		fmt.Println(failed)
		t.Error("Error occurred, got: ", err)
	}
	fmt.Println(passed)

	fmt.Println("\n=================Decrypt With Encrypted Data====================")
	ds.EncryptedData = tmp
	err = ds.Decrypt(obj.Password)
	if err != nil {
		fmt.Println(failed)
		t.Error("Error occurred, got: ", err)
	}
	fmt.Println(passed)
}

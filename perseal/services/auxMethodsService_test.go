package services

import (
	"testing"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/externaldrive"
	"github.com/EC-SEAL/perseal/sm"
)

//go test -v -coverpkg=./... -coverprofile=profile.cov ./...
//go tool cover -html profile

var mockStringQwertyDataStore = "{\"id\":\"DS_3a342b23-8b46-44ec-bb06-a03042135a5e\",\"encryptedData\":\"xRbYPkPKxU8ikEmAKBp7CyG4h8DdybKd7K-jOMvFBe-Gw0fL0_sUAAyDL7wgNTDOxPcU0vlzniPCrIVhDPluVhcvxLZbqJviISLXqyiVucuc8C6uQvTI37MBpfuwuw==\",\"signature\":\"SkwdK6HW4jniY6Uw4j102yI-uVyVMe8G22kUT_j1GiUmOWRYnIjQ73RuX23hfL4UF1CVlgFtTiBSTKX_y5O8auPTD-o-AVJDNnOtXlhXn4xXqdV9zHbfrpXq2OtjtVmaw6uX7LeZrS64Nk5Ey0bFLCxomHfy7UakclsIWQ1_HX_Jgc6rcT1WVYJm8p8dp4JKxPRJR-GhdhoRMV14Jp7C8dG76_rDk25N0J8ggIuNs-wews0NTde7kmiE3K_9hpoRwbo5S5-vhLbNCSKaHzzZF7I1o8TEobwB79t9rvf4vVkHr59vsxeyXsctLv8DZwDt-guDx6zFCWqKTxtsHLkeLA==\",\"signatureAlgorithm\":\"rsa-sha256\",\"encryptionAlgorithm\":\"aes-cfb\"}"

func Init(pds string) dto.PersistenceDTO {
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

	obj, _ := dto.PersistenceBuilder(mockID, session)
	return obj
}

func TestAuxMethodsServices(t *testing.T) {

	// Get Redirect URL Google Drive
	obj := Init("googleDrive")
	url, err := GetRedirectURL(obj)
	correcturl := "https://accounts.google.com/o/oauth2/auth?access_type=offline&client_id=425112724933-9o8u2rk49pfurq9qo49903lukp53tbi5.apps.googleusercontent.com&redirect_uri=http%3A%2F%2Flocalhost%3A8082%2Fper%2Fcode&response_type=code&scope=https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fdrive.file&state=" + obj.ID
	if err != nil {
		t.Error("Error occurred, got: ", err)
	}
	if url != correcturl {
		t.Error("URL is was incorrect, got: ", url, " but wanted ", correcturl)
	}

	// Get Redirect URL One Drive
	obj = Init("oneDrive")
	url, err = GetRedirectURL(obj)
	correcturl = "https://login.live.com/oauth20_authorize.srf?client_id=fff1cba9-7597-479d-b653-fd96c5d56b43&redirect_uri=http%3A%2F%2Flocalhost%3A8082%2Fper%2Fcode&response_type=code&scope=offline_access+files.read+files.read.all+files.readwrite+files.readwrite.all&state=" + obj.ID
	if err != nil {
		t.Error("Error occurred, got: ", err)
	}
	if url != correcturl {
		t.Error("URL is was incorrect, got: ", url, " but wanted ", correcturl)
	}

	//Decrypt DataStore Correct Password
	obj = Init("googleDrive")
	obj.Password = "qwerty"
	dataStore, _ := externaldrive.StoreSessionData(obj)
	err = DecryptAndMarshallDataStore(dataStore, obj)
	if err != nil {
		t.Error("Error occurred, got: ", err)
	}

	//Decrypt DataStore Incorrect Password
	obj = Init("googleDrive")
	obj.Password = "qwerty"
	dataStore, _ = externaldrive.StoreSessionData(obj)
	obj.Password = "qwerty12"
	err = DecryptAndMarshallDataStore(dataStore, obj)
	if err == nil {
		t.Error("Error occurred, should have throwed wrong password error")
	}

	//Validate Valid Signature
	obj = Init("googleDrive")
	obj.Password = "qwerty"
	dataStore, _ = externaldrive.StoreSessionData(obj)
	if !ValidateSignature(dataStore.EncryptedData, dataStore.Signature) {
		t.Error("Error occurred, should have detected valid signature")
	}

	//Validate Invalid Signature
	obj = Init("googleDrive")
	obj.Password = "qwerty"
	dataStore, _ = externaldrive.StoreSessionData(obj)
	if ValidateSignature(dataStore.EncryptedData, "invalid") {
		t.Error("Error occurred, should have detected invalid signature")
	}

}

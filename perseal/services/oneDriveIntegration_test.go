package services

import (
	"fmt"
	"testing"

	"github.com/EC-SEAL/perseal/sm"
)

func TestOneDriveService(t *testing.T) {

	var passed = "=================PASSED==============="
	var failed = "=================FAILED==============="

	obj := InitIntegration("oneDrive")
	smResp, _ := sm.GetSessionData(obj.ID)

	fmt.Println("\n=================Correct OneDrive Store====================")
	obj = preCloudConfig(obj, smResp, "qwerty")
	obj.DataStoreFileName = "dsTest0"
	_, err := storeCloudData(obj)
	if err != nil {
		fmt.Println(failed)
		t.Error("Thrown error, got: ", err)
	} else {
		fmt.Println(passed)
	}

	fmt.Println("\n=================Incorrect OneDrive Store - Bad Access Token====================")
	obj = preCloudConfig(obj, smResp, "qwerty")
	obj.OneDriveToken.AccessToken = ""
	_, err = storeCloudData(obj)
	if err == nil {
		fmt.Println(failed)
		t.Error("Should have thrown error")
	} else {
		fmt.Println(passed)
	}

	fmt.Println("\n=================Incorrect Fetch Cloud File - Bad Access Token====================")
	obj = preCloudConfig(obj, smResp, "qwerty")
	obj.OneDriveToken.AccessToken = ""
	_, err = fetchCloudDataStore(obj)
	if err == nil {
		fmt.Println(failed)
		t.Error("Should have thrown error")
	} else {
		fmt.Println(passed)
	}

	fmt.Println("\n=================Correct Fetch Cloud File====================")
	obj = preCloudConfig(obj, smResp, "qwerty")
	obj.DataStoreFileName = "dsTest0.seal"
	_, err = fetchCloudDataStore(obj)
	if err != nil {
		fmt.Println(failed)
		t.Error("Thrown error, got: ", err)
	} else {
		fmt.Println(passed)
	}

	fmt.Println("\n=================Correct Fetch Cloud Files====================")
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

	obj.DataStoreFileName = "dsTest0.seal"
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

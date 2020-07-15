package services

import (
	"log"
	"testing"
)

func TestOneDriveService(t *testing.T) {

	obj := InitIntegration("oneDrive")

	// Test OneDrive Store
	obj = preCloudConfig(obj, "qwerty")
	ds, err := storeCloudData(obj)
	log.Println(ds)
	log.Println(err)
	if err != nil {
		t.Error("Thrown error, got: ", err)
	}

	// Test Incorrect OneDrive Store
	log.Println("\n\n\nNEW INCORRECT")
	obj = preCloudConfig(obj, "qwerty")
	obj.OneDriveToken.AccessToken = ""
	log.Println("\n\n", obj.OneDriveToken.AccessToken)
	ds, err = storeCloudData(obj)
	log.Println(ds)
	log.Println(err)
	if err == nil {
		t.Error("Should have thrown error")
	}

	// Test Incorrect OneDrive Store
	log.Println("\n\n\nNEW INCORRECT")
	obj = preCloudConfig(obj, "qwerty")
	obj.OneDriveToken.AccessToken = ""
	log.Println("\n\n", obj.OneDriveToken.AccessToken)
	ds, err = fetchCloudDataStore(obj, "datastore.seal")
	log.Println(ds)
	log.Println(err)
	if err == nil {
		t.Error("Should have thrown error")
	}

	// Test Load OneDrive
	obj = preCloudConfig(obj, "qwerty")
	ds, err = fetchCloudDataStore(obj, "datastore.seal")
	if err != nil {
		t.Error("Thrown error, got: ", err)
	}
	log.Println(ds)

	// Test Get Cloud Files
	files, err := GetCloudFileNames(obj)
	if err != nil {
		t.Error("Thrown error, got: ", err)
	}
	if len(files) == 0 {
		t.Error("no files found")
	}

	obj = preCloudConfig(obj, "qwerty")
	_, erro := PersistenceStore(obj)
	log.Println(ds)
	log.Println(erro)
	if erro != nil {
		t.Error("Thrown error, got: ", err)
	}

	obj = preCloudConfig(obj, "qwerty")
	_, erro = PersistenceLoad(obj)
	log.Println(erro)
	if erro != nil {
		t.Error("Thrown error, got: ", err)
	}

	obj = preCloudConfig(obj, "qwerty")
	_, erro = PersistenceStoreAndLoad(obj)
	log.Println(erro)
	if erro != nil {
		t.Error("Thrown error, got: ", err)
	}

}

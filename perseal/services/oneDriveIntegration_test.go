package services

import (
	"log"
	"testing"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/sm"
)

func TestOneDriveService(t *testing.T) {

	obj := InitIntegration("oneDrive")

	// Test OneDrive Store
	session, _ := sm.GetSessionData(obj.ID, "")
	obj, _ = dto.PersistenceWithPasswordBuilder(obj.ID, session, "qwerty")
	log.Println(session)
	ds, err := storeCloudData(obj, "datastore.seal")
	log.Println(ds)
	log.Println(err)
	if err != nil {
		t.Error("Thrown error, got: ", err)
	}

	// Test Incorrect OneDrive Store
	log.Println("\n\n\nNEW INCORRECT")
	session, _ = sm.GetSessionData(obj.ID, "")
	obj, _ = dto.PersistenceWithPasswordBuilder(obj.ID, session, "qwerty")
	obj.OneDriveToken.AccessToken = ""
	log.Println("\n\n", obj.OneDriveToken.AccessToken)
	ds, err = storeCloudData(obj, "datastore.seal")
	log.Println(ds)
	log.Println(err)
	if err == nil {
		t.Error("Should have thrown error")
	}

	// Test Incorrect OneDrive Store
	log.Println("\n\n\nNEW INCORRECT")
	session, _ = sm.GetSessionData(obj.ID, "")
	obj, _ = dto.PersistenceWithPasswordBuilder(obj.ID, session, "qwerty")
	obj.OneDriveToken.AccessToken = ""
	log.Println("\n\n", obj.OneDriveToken.AccessToken)
	ds, err = fetchCloudDataStore(obj, "datastore.seal")
	log.Println(ds)
	log.Println(err)
	if err == nil {
		t.Error("Should have thrown error")
	}

	// Test Load OneDrive
	session, _ = sm.GetSessionData(obj.ID, "")
	obj, _ = dto.PersistenceWithPasswordBuilder(obj.ID, session, "qwerty")
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

	session, _ = sm.GetSessionData(obj.ID, "")
	obj, _ = dto.PersistenceWithPasswordBuilder(obj.ID, session, "qwerty")
	log.Println(session)
	_, erro := PersistenceStore(obj)
	log.Println(ds)
	log.Println(erro)
	if erro != nil {
		t.Error("Thrown error, got: ", err)
	}

	session, _ = sm.GetSessionData(obj.ID, "")
	obj, _ = dto.PersistenceWithPasswordBuilder(obj.ID, session, "qwerty")
	log.Println(session)
	_, erro = PersistenceLoad(obj)
	log.Println(erro)
	if erro != nil {
		t.Error("Thrown error, got: ", err)
	}

	session, _ = sm.GetSessionData(obj.ID, "")
	obj, _ = dto.PersistenceWithPasswordBuilder(obj.ID, session, "qwerty")
	log.Println(session)
	_, erro = PersistenceStoreAndLoad(obj)
	log.Println(erro)
	if erro != nil {
		t.Error("Thrown error, got: ", err)
	}

}

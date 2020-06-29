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
	ds, err := StoreCloudData(obj, "datastore.seal")
	log.Println(ds)
	log.Println(err)

	// Test Load OneDrive
	ds, err = FetchCloudDataStore(obj, "datastore.seal")
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
}

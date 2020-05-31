package services

import (
	"image/png"
	"log"
	"net/http"
	"os"

	"github.com/EC-SEAL/perseal/externaldrive"
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/skip2/go-qrcode"
)

// Service Method to Fetch the DataStore according to the PDS variable
func FetchCloudDataStore(smResp sm.SessionMngrResponse, pds string, filename string) (dataStore *externaldrive.DataStore, err *model.DashboardResponse) {
	
	smResp, err = checkClientId(smResp, pds)
	if err != nil {
		return
	}
	id := smResp.SessionData.SessionID
	var file *http.Response

	if pds == "googleDrive" {
		file, err = loadSessionDataGoogleDrive(smResp, id, filename, "load")
		if err != nil {
			return
		}
		dataStore, err = readBody(file, id)

	} else if pds == "oneDrive" {
		file, err = loadSessionDataOneDrive(smResp, id, filename, "load")
		if err != nil {
			return
		}
		dataStore, err = readBody(file, id)
		log.Println(dataStore)
	}
	return
}

func FetchLocalDataStore(pds string, clientCallback string, smResp sm.SessionMngrResponse) bool {
	qr, _ := qrcode.New(clientCallback+"/cl/persistence/"+pds+"/load?sessionID="+smResp.SessionData.SessionID, qrcode.Medium)
	im := qr.Image(256)
	out, _ := os.Create("./QRImg.png")
	_ = png.Encode(out, im)
	return true
}

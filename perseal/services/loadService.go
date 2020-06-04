package services

import (
	"image/png"
	"log"
	"net/http"
	"os"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/externaldrive"
	"github.com/EC-SEAL/perseal/model"
	"github.com/skip2/go-qrcode"
)

// Service Method to Fetch the DataStore according to the PDS variable
func FetchCloudDataStore(dto dto.PersistenceDTO, filename string) (returningdto dto.PersistenceDTO, dataStore *externaldrive.DataStore, err *model.DashboardResponse) {
	returningdto = dto

	returningdto.SMResp, err = checkClientId(returningdto)
	if err != nil {
		return
	}
	id := returningdto.ID
	var file *http.Response

	if returningdto.PDS == "googleDrive" {
		returningdto, file, err = loadSessionDataGoogleDrive(returningdto, filename)
		if err != nil {
			return
		}
		if returningdto.StopProcess {
			return
		}
		dataStore, err = readBody(file, id)

	} else if returningdto.PDS == "oneDrive" {
		returningdto, file, err = loadSessionDataOneDrive(returningdto, filename)
		if returningdto.StopProcess {
			return
		}
		if err != nil {
			return
		}
		dataStore, err = readBody(file, id)
		log.Println(dataStore)
	}
	return
}

func FetchLocalDataStore(dto dto.PersistenceDTO) bool {
	qr, _ := qrcode.New(dto.ClientCallbackAddr+"/cl/persistence/"+dto.PDS+"/load?sessionID="+dto.ID, qrcode.Medium)
	im := qr.Image(256)
	out, _ := os.Create("./QRImg.png")
	_ = png.Encode(out, im)
	return true
}

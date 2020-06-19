package services

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/dto"
	"github.com/EC-SEAL/perseal/externaldrive"
	"github.com/EC-SEAL/perseal/model"
)

// Service Method to Fetch the DataStore according to the PDS variable
func FetchCloudDataStore(dto dto.PersistenceDTO, filename string) (dataStore *externaldrive.DataStore, err *model.HTMLResponse) {
	if err != nil {
		return
	}
	id := dto.ID
	var file *http.Response

	if dto.PDS == "googleDrive" {
		file, err = loadSessionDataGoogleDrive(dto, filename)
		if err != nil {
			return
		}
		dataStore, err = readBody(file, id)

	} else if dto.PDS == "oneDrive" {
		file, err = loadSessionDataOneDrive(dto, filename)
		if err != nil {
			return
		}
		dataStore, err = readBody(file, id)
		log.Println(dataStore)
	}
	return
}

func FetchLocalDataStore(r *http.Request) (ds *externaldrive.DataStore) {
	file, handler, _ := r.FormFile("file")
	defer file.Close()
	f, _ := handler.Open()
	body, erro := ioutil.ReadAll(f)
	if erro != nil {
		return
	}

	var v string
	str := string(body)
	log.Println("string", str)
	json.Unmarshal([]byte(str), &v)

	log.Println(v)
	if erro != nil {
		return
	}

	err := json.Unmarshal([]byte(v), &ds)
	log.Println(err)
	return
	/*
		qr, _ := qrcode.New(dto.ClientCallbackAddr+"/cl/persistence/"+dto.PDS+"/load?sessionID="+dto.ID, qrcode.Medium)
		im := qr.Image(256)
		out, _ := os.Create("./QRImg.png")
		_ = png.Encode(out, im)
		return true
	*/
}

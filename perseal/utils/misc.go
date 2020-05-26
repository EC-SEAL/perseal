package utils

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
)

func WriteResponseMessage(w http.ResponseWriter, data interface{}, code int) http.ResponseWriter {
	w.WriteHeader(code)
	t, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	w.Write(t)
	return w
}

func GetSessionDataFromMSToken(msToken string) (id string, smResp sm.SessionMngrResponse, err *model.DashboardResponse) {
	id, err = sm.ValidateToken(msToken)
	if err != nil {
		return
	}
	smResp, err = sm.GetSessionData(id, "")

	if err != nil {
		return
	}

	log.Println(smResp)

	if err = sm.ValidateSessionMngrResponse(smResp); err != nil {
		return
	}

	return
}

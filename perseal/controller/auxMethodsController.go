package controller

import (
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/EC-SEAL/perseal/utils"
)

// Fetches msToken from Request, validates it and extracts sessionId
func getSessionDataFromMSToken(r *http.Request) (id string, smResp sm.SessionMngrResponse, err *model.DashboardResponse) {
	msToken, err := utils.ReadRequestBody(r)
	if err != nil {
		return
	}

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

//Auxiliary Method for Development: Resets Session Variables of a given SessionId
func Reset(w http.ResponseWriter, r *http.Request) {
	model.Password = nil
	model.ClientCallback = nil
	model.CheckFirstAccess = nil
	model.CloudLogin = nil
	model.Code = nil
	model.Redirect = nil
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w = utils.WriteResponseMessage(w, "", 200)
	return
}

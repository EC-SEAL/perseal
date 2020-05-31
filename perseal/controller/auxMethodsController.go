package controller

import (
	"encoding/json"
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
	sessionToken := r.FormValue("sessionToken")
	if sessionToken == "" {
		err := &model.DashboardResponse{
			Code:    400,
			Message: "Couldn't find Session Token",
		}
		w = utils.WriteResponseMessage(w, err, err.Code)
		return
	}

	sm.UpdateSessionData(sessionToken, "{}", "")
	ti, _ := sm.GetSessionData(sessionToken, "")
	w.WriteHeader(200)
	t, _ := json.MarshalIndent(ti, "", "\t")
	w.Write(t)
}

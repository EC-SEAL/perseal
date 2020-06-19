package dto

import (
	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
	"golang.org/x/oauth2"
)

type PersistenceDTO struct {
	ID                 string
	PDS                string
	Method             string
	ClientCallbackAddr string
	SMResp             sm.SessionMngrResponse
	Password           string
	StoreAndLoad       bool
	GoogleAccessCreds  string
	OneDriveToken      oauth2.Token
	DoesNotHaveFiles   bool
	Response           model.HTMLResponse
	IsLocal            bool
}

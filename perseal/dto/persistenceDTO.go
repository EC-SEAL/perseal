package dto

import (
	"github.com/EC-SEAL/perseal/sm"
	"golang.org/x/oauth2"
)

type PersistenceDTO struct {
	ID                 string
	PDS                string
	Method             string
	ClientCallbackAddr string
	SMResp             sm.SessionMngrResponse
	UUID               string
	Password           string
	StopProcess        bool
	Token              *oauth2.Token
}

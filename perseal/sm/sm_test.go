package sm

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/utils"
	"github.com/joho/godotenv"
)

func TestSessionManagerIntegration(t *testing.T) {

	godotenv.Load("../.env")
	model.SetEnvVariables()

	var passed = "=================PASSED==============="
	var failed = "=================FAILED==============="
	ID, _ := utils.StartSession()
	id := ID.Payload

	fmt.Println("\n=================Correct Generate and Validate Token====================")
	tok, err := GenerateToken(model.EnvVariables.Perseal_Sender_Receiver, model.EnvVariables.Perseal_Sender_Receiver, id)
	if err != nil {
		fmt.Println(failed)
		t.Error("Thrown error, got: ", err)
	}

	smResp, err := ValidateToken(tok.AdditionalData)
	if err != nil {
		fmt.Println(failed)
		t.Error("Thrown error, got: ", err)
	}
	if smResp.SessionData.SessionID != id {
		fmt.Println(failed)
		t.Error("ID's are not the same")
	} else {
		fmt.Println(passed)
	}

	fmt.Println("\n=================Correct Update Session Data====================")
	err = UpdateSessionData(id, "test", "variable")
	if err != nil {
		fmt.Println(failed)
		t.Error("Thrown error, got: ", err)
	} else {
		fmt.Println(passed)
	}

	fmt.Println("\n=================Correct Get Session Data====================")
	tok, err = GetSessionData(id)
	if err != nil {
		fmt.Println(failed)
		t.Error("Thrown error, got: ", err)
	} else if tok.SessionData.SessionVariables["variable"] != "test" {
		fmt.Println(failed)
		t.Error("Variable has not correct value")
	} else {
		fmt.Println(passed)
	}

	fmt.Println("\n=================Correct Generate and Validate Token With Payload====================")
	tok, err = GenerateTokenWithPayload(model.EnvVariables.Perseal_Sender_Receiver, model.EnvVariables.Perseal_Sender_Receiver, id, "payloaddata")
	if err != nil {
		fmt.Println(failed)
		t.Error("Thrown error, got: ", err)
	}

	smResp, err = ValidateToken(tok.AdditionalData)
	var params RequestParameters
	json.Unmarshal([]byte(smResp.AdditionalData), &params)
	if err != nil {
		fmt.Println(failed)
		t.Error("Thrown error, got: ", err)
	}
	if params.Data != "payloaddata" {
		fmt.Println(failed)
		t.Error("Payload data is not the same")
	} else {
		fmt.Println(passed)
	}

	fmt.Println("\n=================Correct Add and Search Entry to DataStore====================")

	var s RequestParameters = RequestParameters{
		SessionId: id,
		Type:      "linkRequest",
		Data:      `{"id": "_69a510337f266e8d0fbcbd15ab197634114f2d810a","type": "Request","issuer": "project-seal.eu_automatedlink","recipient": null,"inResponseTo": null,"loa": "low","notBefore": "2020-04-16T05:26:29Z","notAfter": "2020-04-16T05:31:29Z", "status": null, "uri":  "urn:mace:project-seal.eu:link:project-seal.eu_automatedlink:low:es/es/53377873w:eidas_es:farago@uji.es:uji.es", "attributes": [ { "name": "http://eidas.europa.eu/attributes/naturalperson/PersonIdentifier", "friendlyName": "ES/ES/53377873w", "encoding": null, "language": null, "mandatory": true, "values": null }, { "name": "http://eidas.europa.eu/attributes/naturalperson/FirstName", "friendlyName": "FRANCISCO JOSE", "encoding": null, "language": null, "mandatory": true, "values": null }, { "name": "http://eidas.europa.eu/attributes/naturalperson/CurrentFamilyName", "friendlyName": "ARAGO MONZONIS", "encoding": null, "language": null, "mandatory": true, "values": null } ], "properties": { "SAML_RelayState": "", "SAML_RemoteSP_RequestId": "_fb9b2627bb634cc7ed98b4f71ee53d93", "SAML_ForceAuthn": true, "SAML_isPassive": false, "SAML_NameIDFormat": "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified", "SAML_AllowCreate": false, "EIDAS_ProviderName": "AAAAAA_BBBBBB", "EIDAS_IdFormat": null, "EIDAS_SPType": null, "EIDAS_LoA": "http://eidas.europa.eu/LoA/low", "EIDAS_country": null, "NameIDFormat": null, "NameQualifier": null, "NameID": null } }`,
		ID:        "seal_link_test_bruno",
	}
	tok, err = NewAdd(s)
	if err != nil {
		fmt.Println(failed)
		t.Error("Thrown error, got: ", err)
	}

	smResp, err = NewSearch(id, s.Type)
	var params2 RequestParameters
	smResp.AdditionalData = trimFirstRune(smResp.AdditionalData)
	json.Unmarshal([]byte(smResp.AdditionalData), &params2)
	if err != nil {
		fmt.Println(failed)
		t.Error("Thrown error, got: ", err)
	}
	if params2.ID != s.ID {
		fmt.Println(failed)
		t.Error("Entry is not the same")
	} else {
		fmt.Println(passed)
	}
	fmt.Println("\n=================Correct Start Session====================")

	tok, err = StartSession(id)
	if err != nil {
		fmt.Println(failed)
		t.Error("Thrown error, got: ", err)
	}

	smResp, err = NewSearch(id)
	smResp.AdditionalData = trimFirstRune(smResp.AdditionalData)

	if err != nil {
		fmt.Println(failed)
		t.Error("Thrown error, got: ", err)
	} else if smResp.AdditionalData != "" {
		fmt.Println(failed)
		t.Error("Session is not empty")
	} else {
		fmt.Println(passed)
	}

}

func trimFirstRune(s string) string {

	sz := len(s)
	if sz > 0 && s[sz-1] == ']' {
		s = s[:sz-1]
	}

	for i := range s {
		if i > 0 {
			// The value i is the index in s of the second
			// rune.  Slice to remove the first rune.
			return s[i:]
		}
	}
	// There are 0 or 1 runes in the string.
	return ""
}

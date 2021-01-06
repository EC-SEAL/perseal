package controller

import (
	"html/template"
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/services"
	"github.com/EC-SEAL/perseal/sm"
)

func Link(w http.ResponseWriter, r *http.Request) {

	source := getQueryParameter(r, "source")
	id := getQueryParameter(r, "id")
	log.Println(r.URL.Path)
	log.Println(source)

	keyPair, key := loadCertificates()

	// TODO: To switch the SSO url when eIDAS and eduGAIN nodes are configured
	data, erro := services.MakeSAMLRequest(key, keyPair, SSOUrlTest.String(), source, id)
	if erro != nil {
		w.WriteHeader(erro.Code)
		w.Write([]byte(erro.Message))
		return
	}

	t := template.New("saml")
	t, _ = t.Parse("<html><body style=\"display: none\" onload=\"document.frm.submit()\"><form method=\"post\" name=\"frm\" action=\"{{.URL}}\"><input type=\"hidden\" name=\"SAMLRequest\" value=\"{{.Base64AuthRequest}}\" /><input type=\"submit\" value=\"Submit\" /></form></body></html>")
	t.Execute(w, data)

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

func getLinkRequest(msToken string) (link sm.SessionMngrResponse, encodedXML string) {
	smResp, _ := sm.ValidateToken(msToken)
	id := smResp.SessionData.SessionID

	smResp, _ = sm.GetSessionData(id)

	link, _ = sm.NewSearch(id, "linkRequest")
	encodedXML = smResp.SessionData.SessionVariables["SAMLResponseEncoded"]
	return
}

// EP to receive the LinkRequest object from the PDS
func RetrieveDSData(w http.ResponseWriter, r *http.Request) {
	keyPair, _ := loadCertificates()

	msToken := getQueryParameter(r, "msToken")
	link, encodedXML := getLinkRequest(msToken)

	link.AdditionalData = trimFirstRune(link.AdditionalData)

	// parses the contents of the msToken to a id-type-data struct
	var i interface{}
	i = services.UnMarshallJson(link.AdditionalData, "data")
	linkRequestData := i.(sm.RequestParameters).Data

	user, err := services.HandleLinkRequest(keyPair, encodedXML, linkRequestData)
	if err != nil {
		w.WriteHeader(err.Code)
		w.Write([]byte(err.Message))
		return
	}

	log.Println(user)
	log.Println(user.PersonalInformation.DisplayName)
	log.Println(user.PersonalInformation.EduOrgLocation)
	log.Println(user.PersonalInformation.Mobile)
	model.PersonalInformation = user.PersonalInformation
	return
}

package controller

import (
	"crypto"
	"crypto/tls"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/services"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/EC-SEAL/perseal/utils"
	"software.sslmate.com/src/go-pkcs12"
)

var IDPMetadataURL, _ = url.Parse("https://stork.uji.es/ujiidp/saml2/idp/metadata.php")

// SEAL SSO page to retrieve identity from <source>
var SSOurl, _ = url.Parse("https://vm.project-seal.eu:9060/auth/saml2/idp/SSOService.php")

// Paco's SSO page to retrieve identity from <source>
var SSOUrlTest, _ = url.Parse("https://stork.uji.es/ujiidp/saml2/idp/SSOService.php")

type IDMsToken struct {
	ID      string
	MSToken string
}

// EP for connecting to the SAML SP, specifying the source of the attribute retrieval (eIDAS, eduGAIN, etc)
func TestSAML(w http.ResponseWriter, r *http.Request) {

	source := getQueryParameter(r, "source")
	log.Println(r.URL.Path)
	log.Println(source)

	keyPair, key := loadCertificates()

	// TODO: To switch the SSO url when eIDAS and eduGAIN nodes are configured
	data, erro := services.MakeSAMLRequest(key, keyPair, SSOUrlTest.String(), source, "")
	if erro != nil {
		w.WriteHeader(erro.Code)
		w.Write([]byte(erro.Message))
		return
	}

	t := template.New("saml")
	t, _ = t.Parse("<html><body style=\"display: none\" onload=\"document.frm.submit()\"><form method=\"post\" name=\"frm\" action=\"{{.URL}}\"><input type=\"hidden\" name=\"SAMLRequest\" value=\"{{.Base64AuthRequest}}\" /><input type=\"submit\" value=\"Submit\" /></form></body></html>")
	t.Execute(w, data)

}

// ACS URL EP, receives attributes from the specified IdP
func TestSAMLResponse(w http.ResponseWriter, r *http.Request) {

	keyPair, key := loadCertificates()

	encodedXML := r.FormValue("SAMLResponse")

	response, erro := services.RetrieveSAMLResponse(keyPair.PrivateKey, keyPair, encodedXML)
	if erro != nil {
		w.WriteHeader(erro.Code)
		w.Write([]byte(erro.Message))
		return
	}

	id := response.InResponseTo
	// false if "id" represents a user in the system
	var newId bool

	// if the ID represents a user in the system, it means he requested a linkRequest oo the PDS
	possibleToken, _ := utils.GenerateTokenAPI("Browser", id)
	log.Println(possibleToken)
	if possibleToken == "" {
		smResp, _ := utils.StartSession()
		model.TestUser = smResp.Payload
		id = model.TestUser
		newId = true
	} else {
		newId = false
	}

	model.MSToken, _ = utils.GenerateTokenAPI("Browser", id)

	sm.UpdateSessionData(id, encodedXML, "SAMLResponseEncoded")

	//TODO: change url
	sm.UpdateSessionData(id, "http://localhost:8082/per/retrieve", model.EnvVariables.SessionVariables.ClientCallbackAddr)

	saml, _ := services.RetrieveSAMLResponse(key, keyPair, encodedXML)
	var user model.Client
	user = services.LowAttributes(saml, user)
	user = services.MediumAttributes(saml, user)
	user = services.HighAttributes(saml, user)

	// TODO: This is a Mock for Accessing the PDS file to retrieve the LinkRequest
	// Will be altered once the RM can handle the flow: 1 - read PDS; 2 - retrieve the linkRequest; 3 - redirect it to the UPorto SP
	if !newId {
		IDMsToken := IDMsToken{
			ID:      id,
			MSToken: model.MSToken,
		}

		t := template.New("saml")

		//TODO: change url (the RM might do this automatically)
		t, _ = t.Parse("<html><body style=\"display: none\" onload=\"document.frm.submit()\"><form action=\"http://localhost:8082/per/load\"  name=\"frm\" method=\"get\"><input name=\"msToken\" type=\"hidden\" value=\"{{.MSToken}}\" /><input type=\"submit\"/></form></body></html>")
		t.Execute(w, IDMsToken)
		return
	}

	// The model.PersonalInformation Represents the current user data in the system
	model.PersonalInformation = user.PersonalInformation

	//TODO: change url
	//If linkRequest not solicited, go back to the main page
	http.Redirect(w, r, "http://localhost:8082/per/test/simulateDashboard", 302)

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
	linkRequestData := i.(sm.NewUpdateDataRequest).Data

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

func loadCertificates() (keyPair tls.Certificate, key interface{}) {
	p12_data, err := ioutil.ReadFile(os.Getenv("SSL_P12"))
	if err != nil {
		log.Println(err)
		return
	}
	key, cert, err := pkcs12.Decode(p12_data, os.Getenv("SSL_KEY_PASS")) // Note the order of the return values.
	if err != nil {
		log.Println(err)
		return
	} else {
		log.Println("Decoded the key and cert")
	}

	keyPair = tls.Certificate{
		Certificate: [][]byte{cert.Raw},
		PrivateKey:  key.(crypto.PrivateKey),
		Leaf:        cert,
	}
	return
}

func getLinkRequest(msToken string) (link sm.SessionMngrResponse, encodedXML string) {
	smResp, _ := sm.ValidateToken(msToken)
	id := smResp.SessionData.SessionID

	smResp, _ = sm.GetSessionData(id)

	link, _ = sm.NewSearch(id, "linkRequest")
	encodedXML = smResp.SessionData.SessionVariables["SAMLResponseEncoded"]
	return
}

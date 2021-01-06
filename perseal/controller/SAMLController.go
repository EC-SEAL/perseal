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
	"software.sslmate.com/src/go-pkcs12"
)

//var IDPMetadataURL, _ = url.Parse("https://stork.uji.es/ujiidp/saml2/idp/metadata.php")

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

	source = ""
	keyPair, key := loadCertificates()

	// TODO: To switch the SSO url when eIDAS and eduGAIN nodes are configured
	data, erro := services.MakeSAMLRequest(key, keyPair, SSOurl.String(), source, "")
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

	keyPair, _ := loadCertificates()

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
	/*
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
	*/
	sm.UpdateSessionData(id, encodedXML, "SAMLResponseEncoded")

	//TODO: change url
	//sm.UpdateSessionData(id, "http://localhost:8082/per/retrieve", model.EnvVariables.SessionVariables.ClientCallbackAddr)

	log.Println(response)
	var user = model.Client{
		PersonalInformation: model.PersonalInformation,
	}
	user = services.LowAttributes(response, user)
	user = services.MediumAttributes(response, user)
	user = services.HighAttributes(response, user)

	log.Println(user)
	newId = true
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
	http.Redirect(w, r, "http://localhost:8082/per/test/showdetails", 302)

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

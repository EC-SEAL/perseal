package services

import (
	"bytes"
	"crypto/rsa"
	"crypto/tls"
	"encoding/base64"
	"log"
	"net/http"
	"net/url"
	"strings"
	"text/template"

	"github.com/EC-SEAL/perseal/model"
	samlrp "github.com/RobotsAndPencils/go-saml"
	"github.com/beevik/etree"
	"github.com/crewjam/saml"
	dsig "github.com/russellhaering/goxmldsig"
	"golang.org/x/net/html"
)

//var IDPMetadataURL, _ = url.Parse("https://stork.uji.es/ujiidp/saml2/idp/metadata.php")
//var SSOurl, _ = url.Parse("https://vm.project-seal.eu:9060/auth/saml2/idp/SSOService.php")
//var SSOUrlTest, _ = url.Parse("https://stork.uji.es/ujiidp/saml2/idp/SSOService.php")
var SPMetadataURL, _ = url.Parse("https://vm.project-seal.eu:9060/auth/saml2/idp/metdata.php")
var AcsURL, _ = url.Parse("http://localhost:8082/per/testsamlresp?id=")

type SAMLVariables struct {
	Base64AuthRequest string
	URL               string
}

var sp = saml.ServiceProvider{
	EntityID:    "http://uporto.pt/",
	MetadataURL: *SPMetadataURL,
	AcsURL:      *AcsURL,
}

func MakeSAMLRequest(key interface{}, keyPair tls.Certificate, ssourl, source, id string) (samlvar SAMLVariables, err *model.HTMLResponse) {

	sp.Certificate = keyPair.Leaf
	sp.Key = key.(*rsa.PrivateKey)

	keyStore := dsig.TLSCertKeyStore(keyPair)
	signingContext := dsig.NewDefaultSigningContext(keyStore)

	var req *saml.AuthnRequest
	req, err = signSAMLAuthNRequest(&sp, signingContext, ssourl, source, id)

	body, err := generateSAMLRequestForm(req)
	if err != nil {
		return
	}

	samlrequest := findSAMLRequestValue(body)
	samlvar = SAMLVariables{Base64AuthRequest: samlrequest, URL: ssourl}

	return

}

func signSAMLAuthNRequest(sp *saml.ServiceProvider, signingContext *dsig.SigningContext, url, source, id string) (authnRequest *saml.AuthnRequest, err *model.HTMLResponse) {

	signingContext.Canonicalizer = dsig.MakeC14N10ExclusiveCanonicalizerWithPrefixList("")
	erro := signingContext.SetSignatureMethod(dsig.RSASHA256SignatureMethod)
	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, "SAML REQUEST FAILURE", erro.Error())
		return
	}

	authnRequest, erro = sp.MakeAuthenticationRequest(url)
	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, "SAML REQUEST FAILURE", erro.Error())
		return
	}
	/*
		authnRequest.ProtocolBinding = "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
		authnRequest.Destination = "https://preprod.eidas.autenticacao.gov.pt/EidasNode/IdpResponse"
		authnRequest.Issuer = &saml.Issuer{XMLName: xml.Name{Local: "UPORTO"}, Value: "ABC"}
	*/
	if id != "" {
		authnRequest.ID = id
	}
	elem := authnRequest.Element()

	log.Println(source)
	if source != "" {
		scope := elem.CreateElement("samlp:Scoping")
		idplist := scope.CreateElement("samlp:IDPList")
		idpe := idplist.CreateElement("samlp:IDPEntry")
		idpe.CreateAttr("ProviderID", source)
	}

	sign, erro := signingContext.SignEnveloped(elem)
	if erro != nil {
		err = model.BuildResponse(http.StatusBadRequest, model.Messages.FileContainsDataStore, erro.Error())
		return
	}

	authnRequest.Signature = sign
	return
}

func generateSAMLRequestForm(req *saml.AuthnRequest) (body []byte, err *model.HTMLResponse) {

	readXMLDoc := etree.NewDocument()
	readXMLDoc.SetRoot(req.Signature)

	reqBuf, erro := readXMLDoc.WriteToBytes()
	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, "SAML REQUEST FAILURE", erro.Error())
		return
	}
	encodedReqBuf := base64.StdEncoding.EncodeToString(reqBuf)

	tmpl := template.Must(template.New("saml-post-form").Parse(`` +
		`<form method="post" action="{{.URL}}" id="SAMLRequestForm">` +
		`<input type="hidden" name="SAMLRequest" value="{{.SAMLRequest}}" />` +
		`<input id="SAMLSubmitButton" type="submit" value="Submit" />` +
		`</form>` +
		`<script>document.getElementById('SAMLSubmitButton').style.visibility="hidden";` +
		`document.getElementById('SAMLRequestForm').submit();</script>`))
	data := struct {
		URL         string
		SAMLRequest string
	}{
		URL:         req.Destination,
		SAMLRequest: encodedReqBuf,
	}

	rv := bytes.Buffer{}
	tmpl.Execute(&rv, data)
	// The Form
	body = rv.Bytes()
	return
}

func findSAMLRequestValue(body []byte) (samlrequest string) {
	n, _ := html.Parse(strings.NewReader(string(body)))
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Data == "input" {
			for _, a := range n.Attr {
				if a.Key == "value" && a.Val != "" && a.Val != "Submit" {
					samlrequest = a.Val
					return
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return
}

func RetrieveSAMLResponse(key interface{}, keyPair tls.Certificate, encodedXML string) (response *samlrp.Response, err *model.HTMLResponse) {
	sp.Certificate = keyPair.Leaf
	sp.Key = key.(*rsa.PrivateKey)

	var erro error
	response, erro = samlrp.ParseEncodedResponse(encodedXML)
	if erro != nil {
		err = model.BuildResponse(http.StatusInternalServerError, "SAML REQUEST FAILURE", erro.Error())
	}

	return
}

func LowAttributes(response *samlrp.Response, user model.Client) model.Client {
	if user.PersonalInformation.PersonIdentifier == "" {
		user.PersonalInformation.PersonIdentifier = response.GetAttribute("PersonIdentifier")
	}
	if user.PersonalInformation.FirstName == "" {
		user.PersonalInformation.FirstName = response.GetAttribute("GivenName")
	}
	if user.PersonalInformation.FamilyName == "" {
		user.PersonalInformation.FamilyName = response.GetAttribute("FamilyName")
	}
	if user.PersonalInformation.Mail == "" {
		user.PersonalInformation.Mail = response.GetAttribute("mail")
	}
	if user.PersonalInformation.DisplayName == "" {
		user.PersonalInformation.DisplayName = response.GetAttribute("displayName")
	}
	if user.PersonalInformation.Surname == "" {
		user.PersonalInformation.Surname = response.GetAttribute("sn")
	}
	if user.PersonalInformation.GivenName == "" {
		user.PersonalInformation.GivenName = response.GetAttribute("givenName")
	}
	return user
}

func MediumAttributes(response *samlrp.Response, user model.Client) model.Client {
	if user.PersonalInformation.DateOfBirth == "" {
		user.PersonalInformation.DateOfBirth = response.GetAttribute("DateOfBirth")
	}
	if user.PersonalInformation.EduPersonAffiliation == "" {
		user.PersonalInformation.EduPersonAffiliation = response.GetAttribute("eduPersonAffiliation")
	}
	if user.PersonalInformation.EduPersonPrincipalName == "" {
		user.PersonalInformation.EduPersonPrincipalName = response.GetAttribute("eduPersonPrincipalName")
	}
	if user.PersonalInformation.StudyProgram == "" {
		user.PersonalInformation.StudyProgram = response.GetAttribute("studyProgram")
	}
	if user.PersonalInformation.SchacHomeOrganization == "" {
		user.PersonalInformation.SchacHomeOrganization = response.GetAttribute("schacHomeOrganization")
	}
	if user.PersonalInformation.SchacHomeOrganizationType == "" {
		user.PersonalInformation.SchacHomeOrganizationType = response.GetAttribute("schacHomeOrganizationType")
	}
	if user.PersonalInformation.EduOrgLegalName == "" {
		user.PersonalInformation.EduOrgLegalName = response.GetAttribute("eduOrgLegalName")
	}
	if user.PersonalInformation.EduOrgLocation == "" {
		user.PersonalInformation.EduOrgLocation = response.GetAttribute("eduOrgL")
	}
	if user.PersonalInformation.EduOrgHomePageURI == "" {
		user.PersonalInformation.EduOrgHomePageURI = response.GetAttribute("eduOrgHomePageURI")
	}
	return user
}

func HighAttributes(response *samlrp.Response, user model.Client) model.Client {
	if user.PersonalInformation.Mobile == "" {
		user.PersonalInformation.Mobile = response.GetAttribute("mobile")
	}
	if user.PersonalInformation.SchacExpiryDate == "" {
		user.PersonalInformation.SchacExpiryDate = response.GetAttribute("schacExpiryDate")
	}
	if user.PersonalInformation.EduOrgPostalAddress == "" {
		user.PersonalInformation.EduOrgPostalAddress = response.GetAttribute("eduOrgPostalAddress")
	}
	return user
}

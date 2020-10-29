package services

import (
	"bytes"
	"crypto/rsa"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
	"text/template"

	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
	samlrp "github.com/RobotsAndPencils/go-saml"
	"github.com/beevik/etree"
	"github.com/crewjam/saml"
	dsig "github.com/russellhaering/goxmldsig"
	"golang.org/x/net/html"
)

var IDPMetadataURL, _ = url.Parse("https://stork.uji.es/ujiidp/saml2/idp/metadata.php")
var SSOurl, _ = url.Parse("https://vm.project-seal.eu:9060/auth/saml2/idp/SSOService.php")
var SSOUrlTest, _ = url.Parse("https://stork.uji.es/ujiidp/saml2/idp/SSOService.php")
var SPMetadataURL, _ = url.Parse("https://vm.project-seal.eu:9060/auth/saml2/idp/metadata.php")
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

// the Link Request object contained in the PDS and/or SSI wallet
type linkRequest struct {
	URI string `json:"uri"`
	LOA string `json:"loa"`
}

// the main parts of the URI of the linkRequest
type linkURI struct {
	LinkIssuerID string
	LOA          string
	SubjectA     string
	IssuerA      string
	SubjectB     string
	IssuerB      string
}

// The authorized eduGAIN issuers
var eduGAINDomains = []string{"uji.es"}

// The authorized eIDAS issuers
var eIDASDomains = []string{"eidas_es"}

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

	authnRequest.ProtocolBinding = "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
	if id != "" {
		authnRequest.ID = id
	}
	elem := authnRequest.Element()

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

func HandleLinkRequest(keyPair tls.Certificate, encodedXML, linkRequestData string) (user model.Client, err *model.HTMLResponse) {

	// parses the "data" field of the retrieved linkRequest entry to a linkRequest struct
	var i2 interface{}
	i2 = UnMarshallJson(linkRequestData, "linkRequest")
	linkRequest := i2.(linkRequest)
	log.Println("linkRequest\n\n", linkRequest)

	suburi := strings.Split(linkRequest.URI, "urn:mace:project-seal.eu:link:")

	suburi2 := strings.Split(suburi[1], ":")
	linkURIObj := linkURI{
		LinkIssuerID: suburi2[0],
		LOA:          suburi2[1],
		SubjectA:     suburi2[2],
		SubjectB:     suburi2[4],
		IssuerA:      suburi2[3],
		IssuerB:      suburi2[5],
	}

	log.Println("\n\n", linkURIObj)

	response, err := RetrieveSAMLResponse(keyPair.PrivateKey, keyPair, encodedXML)
	if err != nil {
		return
	}

	// if one of the issuers is eIDAS
	var containseIDAS, containseduGAIN bool
	if contains(eIDASDomains, linkURIObj.IssuerA) || contains(eIDASDomains, linkURIObj.IssuerB) {
		containseIDAS = true
	}

	// if one of the issuers is eduGAIN
	if contains(eduGAINDomains, linkURIObj.IssuerA) || contains(eduGAINDomains, linkURIObj.IssuerB) {
		containseduGAIN = true
	}

	log.Println(containseIDAS, containseduGAIN)

	pID := response.GetAttribute("PersonIdentifier")
	log.Println(pID)
	if pID != "" && !containseIDAS {
		err = model.BuildResponse(http.StatusBadRequest, "Link Request does Not Contain a eIDAS link")
		return
	}

	homeOrg := response.GetAttribute("schacHomeOrganization")
	log.Println(homeOrg)
	if homeOrg != "" && !containseduGAIN {
		err = model.BuildResponse(http.StatusBadRequest, "Link Request does Not Contain a eduGAIN link")
		return
	}

	user = model.Client{
		PersonalInformation: model.PersonalInformation,
	}
	// if the PersonalIdentifier is the same as one of the linkRequest's subjects. if the corporate email is the same as one of the linkRequest's subjects.
	if (pID == linkURIObj.SubjectA || pID != linkURIObj.SubjectB) && (homeOrg == linkURIObj.SubjectA || homeOrg != linkURIObj.SubjectB) {

		log.Println("\n\nLevel of Assurance: ", linkURIObj.LOA)

		if linkURIObj.LOA == "low" {
			user.PersonalInformation.LoALvl = 1
		} else if linkURIObj.LOA == "medium" {
			user.PersonalInformation.LoALvl = 2
		} else if linkURIObj.LOA == "high" {
			user.PersonalInformation.LoALvl = 3
		}

		// "LOW" level of assurance attributes to be filled (may be customized at will)
		if user.PersonalInformation.LoALvl >= 1 {
			user = LowAttributes(response, user)
		}

		// "MEDIUM" level of assurance attributes to be filled (may be customized at will)
		if user.PersonalInformation.LoALvl >= 2 {
			user = MediumAttributes(response, user)
		}

		// "HIGH" level of assurance attributes to be filled (may be customized at will)
		if user.PersonalInformation.LoALvl >= 3 {
			user = HighAttributes(response, user)
		}
	}

	// EWP open login Page
	/*
		type loginStatus struct {
			Tried   bool
			Valid   bool
			Message string
			User    model.Client
		}
		gui := loginStatus{
			Tried:   true,
			Valid:   true,
			Message: "Nice",
			User: model.Client{
				LoggedIn: true,
				Username: response.GetAttribute("displayName"),
			},
		}
		t, _ := template.ParseFiles("ui/login.html")
		t.Execute(w, gui)
	*/
	return
}

func UnMarshallJson(s, t string) interface{} {
	var result interface{}
	err := json.Unmarshal([]byte(s), &result)
	if err != nil {
		return nil
	}

	jsonM, err := json.Marshal(result)
	if err != nil {
		return nil
	}

	if t == "data" {
		var d sm.RequestParameters
		json.Unmarshal(jsonM, &d)
		return d
	} else if t == "linkRequest" {
		var l linkRequest
		json.Unmarshal(jsonM, &l)
		return l
	}
	return ""

}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func LowAttributes(response *samlrp.Response, user model.Client) model.Client {
	if user.PersonalInformation.FirstName == "" {
		user.PersonalInformation.FirstName = response.GetAttribute("FirstName")
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

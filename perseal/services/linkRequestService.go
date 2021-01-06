package services

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
)

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
		err = model.BuildResponse(http.StatusBadRequest, "Link Request does Not Contain a eIDAS link", "")
		return
	}

	homeOrg := response.GetAttribute("schacHomeOrganization")
	log.Println(homeOrg)
	if homeOrg != "" && !containseduGAIN {
		err = model.BuildResponse(http.StatusBadRequest, "Link Request does Not Contain a eduGAIN link", "")
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
	/*
		// EWP open login Page

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

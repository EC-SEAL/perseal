package controller

import (
	"html/template"
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/EC-SEAL/perseal/utils"
)

func SimulateDashboard(w http.ResponseWriter, r *http.Request) {

	type testStruct struct {
		ID        string
		MSToken   string
		DataStore string
	}

	testing := testStruct{
		ID:        model.TestUser,
		MSToken:   model.MSToken,
		DataStore: model.DataStore,
	}
	t, _ := template.ParseFiles("ui/simulateDashboard.html")
	t.Execute(w, testing)
}

func StartSession(w http.ResponseWriter, r *http.Request) {
	//	resp, _ := utils.StartSession()

	respo, _ := utils.StartSession("")
	log.Println(respo)
	model.TestUser = respo.Payload
	/*
		var s sm.NewUpdateDataRequest = sm.NewUpdateDataRequest{
			SessionId: model.TestUser,
			Type:      "dataSet",
			Data:      `{"id":"2d47b5dd-2ea2-4a1d-8fc4-b1756e1ec72a","type":"eduGAIN","categories":null,"issuerId":"This is the user ID.","subjectId":null,"loa":null,"issued":"Wed, 16 Sep 2020 12:59:54 gmt","expiration":null,"attributes":[{"name":"urn:oid:1.3.6.1.4.1.5923.1.1.1.10","friendlyname":"edupersontargetedid","encoding":null,"language":null,"values":[null]},{"name":"urn:oid:2.5.4.42","friendlyname":"givenname","encoding":null,"language":null,"values":["seal"]},{"name":"urn:oid:0.9.2342.19200300.100.1.3","friendlyname":"mail","encoding":null,"language":null,"values":["seal-test0@example.com"]},{"name":"urn:oid:2.5.4.3","friendlyName":"cn","encoding":null,"language":null,"values":["Tester0 SEAL"]},{"name":"urn:oid:2.5.4.4","friendlyName":"sn","encoding":null,"language":null,"values":["Tester0"]},{"name":"urn:oid:2.16.840.1.113730.3.1.241","friendlyName":"displayName","encoding":null,"language":null,"values":["SEAL tester0"]},{"name":"urn:oid:1.3.6.1.4.1.5923.1.1.1.6","friendlyname":"edupersonprincipalname","encoding":null,"language":null,"values":["128052@gn-vho.grnet.gr"]},{"name":"urn:oid:1.3.6.1.4.1.5923.1.1.1.7","friendlyname":"edupersonentitlement","encoding":null,"language":null,"values":["urn:mace:grnet.gr:seal:test"]}],"properties":null}`,
			ID:        "eIDASeidas.gr/gr/ermis-11076669",
		}

		var s2 sm.NewUpdateDataRequest = sm.NewUpdateDataRequest{
			SessionId: model.TestUser,
			Type:      "dataSet",
			Data:      `{"id":"2d47b5dd-2ea2-4a1d-8fc4-b1756e1ec72a","type":"eduGAIN","categories":null,"issuerId":"This is the user ID.","subjectId":null,"loa":null,"issued":"Wed, 16 Sep 2020 12:59:54 gmt","expiration":null,"attributes":[{"name":"urn:oid:1.3.6.1.4.1.5923.1.1.1.10","friendlyname":"edupersontargetedid","encoding":null,"language":null,"values":[null]},{"name":"urn:oid:2.5.4.42","friendlyname":"givenname","encoding":null,"language":null,"values":["seal"]},{"name":"urn:oid:0.9.2342.19200300.100.1.3","friendlyname":"mail","encoding":null,"language":null,"values":["seal-test0@example.com"]},{"name":"urn:oid:2.5.4.3","friendlyName":"cn","encoding":null,"language":null,"values":["Tester0 SEAL"]},{"name":"urn:oid:2.5.4.4","friendlyName":"sn","encoding":null,"language":null,"values":["Tester0"]},{"name":"urn:oid:2.16.840.1.113730.3.1.241","friendlyName":"displayName","encoding":null,"language":null,"values":["SEAL tester0"]},{"name":"urn:oid:1.3.6.1.4.1.5923.1.1.1.6","friendlyname":"edupersonprincipalname","encoding":null,"language":null,"values":["128052@gn-vho.grnet.gr"]},{"name":"urn:oid:1.3.6.1.4.1.5923.1.1.1.7","friendlyname":"edupersonentitlement","encoding":null,"language":null,"values":["urn:mace:grnet.gr:seal:test"]}],"properties":null}`,
			ID:        "eduPersonTargetedID",
		}
		sm.NewAdd(s)
		sm.NewAdd(s2)
	*/
	/*
		var s sm.NewUpdateDataRequest = sm.NewUpdateDataRequest{
			SessionId: model.TestUser,
			Type:      "ahahachanged",
			Data:      "this is !!!!!!!!!!!!!!!!!!!",
			ID:        uuid.New().String(),
		}
		sm.NewAdd(s)
		var s2 sm.NewUpdateDataRequest = sm.NewUpdateDataRequest{
			SessionId: model.TestUser,
			Type:      "new tipo like what?",
			Data:      "this is smt very coiso",
			ID:        uuid.New().String(),
		}

		sm.NewAdd(s2)
		log.Println(sm.NewSearch(model.TestUser))
	*/
	var url string
	model.DataStore = `"{\"id\":\"efc9b9f0-333e-45e1-9e98-34e0eab3b067\",\"encryptedData\":\"wEbokpp5bg9-oQ9dJu4LpOznRDs=\",\"signature\":\"dM5CePWx7CGtajSpx8WpkwFLkMci_v-_ogAo5Heh7vu6Ex9ucpHRE3sIsvxaLcc3hux_QH9yyISZhCGSwnC6XBGOBDftDp00PiwCOP_2gaZp9ZmVRoLSajIkPXhpPWpd6vEHGlN0GYyKXqCV_NWwovm0iFRM5YGi9j3Bw6MKvIq9TReNPrHYRq5YSGSE3-7mEFuU34uND1Di7ZdHDe2CzE3Y8q2vN26uan7GNDxQw-Vt4CTrGZZIJvBsjMujZMQCRXx1FdOh39cvIGd5nl3gkVhQFCTmaT1XogycMsNGfyHGWSdu8wpjtVpovbMm1xB3oUqV1pyh0pXA5d6pTL_Pcg==\",\"signatureAlgorithm\":\"rsa-sha256\",\"encryptionAlgorithm\":\"aes-cfb\",\"clearData\":\"[]\"}"`
	if model.Test {
		url = "http://localhost:8082"
	} else {
		url = "http://perseal.seal.eu:8082"
	}
	http.Redirect(w, r, url+"/per/test/simulateDashboard", 302)
}

func Token(w http.ResponseWriter, r *http.Request) {
	log.Println(sm.NewSearch(model.TestUser))
	var method string
	if keys, ok := r.URL.Query()["method"]; ok {
		method = keys[0]
	}

	var err *model.HTMLResponse
	model.MSToken, err = utils.GenerateTokenAPI(method, model.TestUser)
	if err != nil {
		log.Println(err)
	}
	log.Println(sm.GetSessionData(model.TestUser))
	log.Println(model.MSToken)
	var url string
	if model.Test {
		url = "http://localhost:8082"
	} else {
		url = "http://perseal.seal.eu:8082"
	}
	http.Redirect(w, r, url+"/per/test/simulateDashboard", 302)
}

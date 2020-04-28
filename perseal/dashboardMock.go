package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
)

func dashboardMock(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	moduleId := r.FormValue("moduleId")
	log.Println(code)
	url := os.Getenv("REDIRECT_URL") + "?code=" + code + "&sessionId=4162ccba-be88-4954-8fe2-f25f8aaa88a3&" + "moduleId=" + moduleId
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	resp, _ := http.Get(url)
	log.Println(url)
	log.Println(resp)
}

// Recieve Code from the redirect url in browser
func getCodeFromDashboard(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	sessionID := r.FormValue("sessionId")
	module := r.FormValue("module")
	persistenceStoreWithCode(code, sessionID, module)
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(200)
}

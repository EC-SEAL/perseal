package main

import (
	"crypto/tls"
	"log"
	"net/http"
)

func dashboardMock(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	moduleId := r.FormValue("moduleId")
	log.Println(code)
	url := "https://localhost:8082/per/code?code=" + code + "&sessionId=81f9937d-a17e-462c-a3e9-aeb0c1d5ad1e&" + "moduleId=" + moduleId
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

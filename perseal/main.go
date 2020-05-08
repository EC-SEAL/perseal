package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// const sm_endpoint = "https://sm"

// GoogleDrive storage info for A GoogleDrive storage

// type User struct {
// 	Name   string
// 	stores []DataStore
// }

type systemOpts struct {
	Port string
}

var local = false

func main() {
	// load systemOpts

	r := newRouter()

	tlsConfig := &tls.Config{
		ClientAuth:         tls.RequestClientCert,
		InsecureSkipVerify: true,
	}

	tlsConfig.BuildNameToCertificate()

	server := &http.Server{
		TLSConfig: tlsConfig,
		Addr:      os.Getenv("HOST"),
		//Addr:         "localhost:8082",
		Handler:      r,
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
	}

	//start listening on port 8082

	log.Fatal(server.ListenAndServe())
	fmt.Println("Persistent SEAL module running on HTTP port " + os.Getenv("PERSEAL_EXT_PORT"))
}

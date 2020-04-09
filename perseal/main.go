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

var (
	PBKDF2SALT string = os.Getenv("PBKDF2SALT") // to make
)

func main() {
	// load systemOpts

	r := newRouter()

	tlsConfig := &tls.Config{
		ClientAuth:         tls.RequestClientCert,
		InsecureSkipVerify: true,
	}

	tlsConfig.BuildNameToCertificate()

	server := &http.Server{
		TLSConfig:    tlsConfig,
		Addr:         os.Getenv("HOST"),
		Handler:      r,
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
	}

	//start listening on port 8082

	fmt.Println("Persistent SEAL module running on HTTP port " + os.Getenv("PERSEAL_INT_PORT"))
	log.Fatal(server.ListenAndServeTLS("./certificate.crt", "./private.key"))

}

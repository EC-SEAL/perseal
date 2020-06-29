package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/EC-SEAL/perseal/model"
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

func main() {
	// load systemOpts

	r := newRouter()

	var addr string
	if model.Local {
		addr = "localhost:8082"
		fmt.Println("local development")
	} else {
		addr = os.Getenv("HOST")
		fmt.Println("container development")
	}

	tlsConfig := &tls.Config{
		ClientAuth:         tls.RequestClientCert,
		InsecureSkipVerify: true,
		ServerName:         addr,
	}

	tlsConfig.BuildNameToCertificate()

	server := &http.Server{
		TLSConfig:    tlsConfig,
		Addr:         addr,
		Handler:      r,
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
	}
	//start listening on port 8082

	http.Handle("/per/ui/", http.StripPrefix("/per/ui/", http.FileServer(http.Dir("ui/"))))
	fmt.Println("Persistent SEAL module running on HTTP port " + os.Getenv("PERSEAL_EXT_PORT"))
	server.ListenAndServe()
}

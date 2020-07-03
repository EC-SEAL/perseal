package main

import (
	"bytes"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/EC-SEAL/perseal/model"
	"software.sslmate.com/src/go-pkcs12"
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

	log.Println("entered main function")

	r := newRouter()

	var addr string
	if model.Test {
		addr = "localhost:8082"
		fmt.Println("testing mode")
	} else {
		addr = os.Getenv("HOST")
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

	if model.Test {
		fmt.Println("Persistent SEAL module running on HTTP port " + os.Getenv("PERSEAL_EXT_PORT"))
		server.ListenAndServe()
	} else {
		listenAndServeTLS(server)
	}

}

func listenAndServeTLS(server *http.Server) {

	p12_data, err := ioutil.ReadFile(os.Getenv("INTER_SSL"))
	if err != nil {
		log.Println(err)
	} else {
		log.Println("I read it!")
	}
	key, cert, err := pkcs12.Decode(p12_data, os.Getenv("SSL_KEY_PASS")) // Note the order of the return values.
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Decoded the key and cert")
	}
	tlsCert := tls.Certificate{
		Certificate: [][]byte{cert.Raw},
		PrivateKey:  key.(crypto.PrivateKey),
		Leaf:        cert,
	}

	log.Println("Everything worked correctly. You should now have a cert and key file in your working directory")

	privBytes, err := x509.MarshalPKCS8PrivateKey(tlsCert.PrivateKey)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("x509 Marshall PKCS8 Private Key")
	}
	addr := server.Addr
	if addr == "" {
		addr = ":https"
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("listening")
	}

	config := server.TLSConfig.Clone()
	config.Certificates = make([]tls.Certificate, 1)
	var certBuf, keyBuf bytes.Buffer
	pem.Encode(&certBuf, &pem.Block{Type: "CERTIFICATE", Bytes: tlsCert.Leaf.Raw})
	pem.Encode(&keyBuf, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	config.Certificates[0], err = tls.X509KeyPair(certBuf.Bytes(), keyBuf.Bytes())
	if err != nil {
		log.Println(err)
	} else {
		log.Println("config certificates found")
		fmt.Println("Persistent SEAL module running on HTTPS port " + os.Getenv("PERSEAL_EXT_PORT"))
	}
	defer ln.Close()

	listen := tls.NewListener(ln, config)
	server.Serve(listen)
}

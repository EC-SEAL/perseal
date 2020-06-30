package main

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/EC-SEAL/perseal/model"
	"github.com/pavel-v-chernykh/keystore-go"
	"golang.org/x/crypto/pkcs12"
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
	if model.Test {
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
	//log.Fatalln(server.ListenAndServeTLS("./certificate_ssl.crt", "./private_ssl.key"))
}

func writeKeyStore(keyStore keystore.KeyStore, filename string, password []byte) {
	o, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer o.Close()
	err = keystore.Encode(o, keyStore, password)
	if err != nil {
		log.Fatal(err)
	}
}

func readKeyStore(filename string, password []byte) keystore.KeyStore {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	keyStore, err := keystore.Decode(f, password)
	if err != nil {
		log.Fatal(err)
	}
	return keyStore
}

func httpstest() {
	/*
		log.Println(exec.Command("keytool", "-importkeystore", "-srckeystore"+os.Getenv("SSL_KEYSTORE_PATH"), "-srcstorepass"+os.Getenv("SSL_KEY_PASS"), "-srcalias"+os.Getenv("SSL_CERT_ALIAS"), "-destalias"+os.Getenv("SSL_CERT_ALIAS"), "-destkeystore inter_ssl.p12", "-deststoretype PKCS12", "-deststorepass"+os.Getenv("SSL_KEY_PASS")).Start())

		log.Println(exec.Command("openssl", "pkcs12", "-in inter_ssl.p12", "-nodes", "-nocerts", "-out private_ssl.key", "-passin pass:"+os.Getenv("SSL_KEYSTORE_PATH")).Start())
		log.Println(exec.Command("openssl", "pkcs12", "-in inter_ssl.p12", "-nodes", "-out certificate_ssl.crt", "-passin pass:"+os.Getenv("SSL_KEYSTORE_PATH")).Start())
		log.Println(exec.Command("ls", ".").Start())
	*/

	//	password := "keystorepass"
	//passwordByte := []byte(password)
	/*
		ks, _ := jceks.LoadFromFile("keystore.jks", passwordByte)
		entry := ks.Entries["perseal"]

		log.Println(entry)

		privKeyEntry := entry.(*jceks.PrivateKeyEntry)
		privBlock, rest := pem.Decode(privKeyEntry.EncodedKey)

		log.Println(privBlock)
		log.Println(string(rest))
		log.Println(x509.ParsePKCS8PrivateKey(rest))
	*/

	p12_data, err := ioutil.ReadFile("INTER_SSL")
	if err != nil {
		log.Println(err)
	}

	key, cert, err := pkcs12.Decode(p12_data, "SSL_KEY_PASS") // Note the order of the return values.
	if err != nil {
		log.Println(err)
	}

	tlsCert := tls.Certificate{
		Certificate: [][]byte{cert.Raw},
		PrivateKey:  key.(crypto.PrivateKey),
		Leaf:        cert,
	}
	certOut, err := os.Create("certificate_ssl.crt")
	if err != nil {
		log.Println("Failed to open certificate_ssl.crt for writing: ", err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: tlsCert.Leaf.Raw}); err != nil {
		log.Println("Failed to write data to certificate_ssl.crt: ", err)
	}
	keyOut, err := os.OpenFile("private_ssl.key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Println("Failed to open private_ssl.key for writing: ", err)
		return
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(tlsCert.PrivateKey)
	if err != nil {
		log.Println("Unable to marshal private key: ", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		log.Println("Failed to write data to private_ssl.key: ", err)
	}
}

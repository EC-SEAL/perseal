package utils

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"testing"
)

//go test -v -coverpkg=./... -coverprofile=profile.cov ./...
//go tool cover -html profile

var password = "qwerty"

func TestEncryption(t *testing.T) {

	//Test Encrypt With Empty Data
	clearData := ""
	data, _ := json.MarshalIndent(clearData, "", "\t")
	_, err := AESEncrypt(Padding([]byte(password), 16), data)
	if err != nil {
		t.Error("Error occurred, got: ", err)
	}

	//Test Correct Encrypt
	clearData = "{\"id\":\"DS_3a342b23-8b46-44ec-bb06-a03042135a5e\",\"encryptedData\":null,\"signature\":\"this is the signature\",\"signatureAlgorithm\":\"this is the signature algorithm\",\"encryptionAlgorithm\":\"this is the encryption algorithm\",\"clearData\":null}"
	data, _ = json.MarshalIndent(clearData, "", "\t")
	blob, err := AESEncrypt(Padding([]byte(password), 16), data)
	if err != nil {
		t.Error("Error occurred, got: ", err)
	}

	encryptedData := base64.URLEncoding.EncodeToString(blob)

	//Test Decryption
	data, err = AESDecrypt(Padding([]byte(password), 16), encryptedData)
	log.Println(string(data))
	if err != nil {
		t.Error("Error occurred, got: ", err)
	}

	//Test SHA256 Hashing
	hashed := HashSUM256(password)
	if hashed == password {
		t.Error("Hashing didn't work")
	}
}

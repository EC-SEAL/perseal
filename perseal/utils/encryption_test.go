package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
)

//go test -v -coverpkg=./... -coverprofile=profile.cov ./...
//go tool cover -html profile

var password = "qwerty"

func TestEncryption(t *testing.T) {

	var passed = "=================PASSED==============="
	var failed = "=================FAILED==============="

	fmt.Println("\n=================Encrypt With Empty Data====================")
	clearData := ""
	data, _ := json.MarshalIndent(clearData, "", "\t")
	_, err := AESEncrypt(Padding([]byte(password), 16), data)
	if err != nil {
		fmt.Println(failed)
		t.Error("Error occurred, got: ", err)
	} else {
		fmt.Println(passed)
	}

	fmt.Println("\n=================Correct Encrypt====================")
	clearData = "{\"id\":\"DS_3a342b23-8b46-44ec-bb06-a03042135a5e\",\"encryptedData\":null,\"signature\":\"this is the signature\",\"signatureAlgorithm\":\"this is the signature algorithm\",\"encryptionAlgorithm\":\"this is the encryption algorithm\",\"clearData\":null}"
	data, _ = json.MarshalIndent(clearData, "", "\t")
	blob, err := AESEncrypt(Padding([]byte(password), 16), data)
	if err != nil {
		fmt.Println(failed)
		t.Error("Error occurred, got: ", err)
	} else {
		fmt.Println(passed)
	}

	encryptedData := base64.URLEncoding.EncodeToString(blob)

	fmt.Println("\n=================Correct Decrypt====================")
	data, err = AESDecrypt(Padding([]byte(password), 16), encryptedData)
	if err != nil {
		fmt.Println(failed)
		t.Error("Error occurred, got: ", err)
	} else {
		fmt.Println(passed)
	}

	fmt.Println("\n=================SHA256 Hashing====================")
	hashed := HashSUM256(password)
	if hashed == password {
		fmt.Println(failed)
		t.Error("Hashing didn't work")
	} else {
		fmt.Println()
	}
}

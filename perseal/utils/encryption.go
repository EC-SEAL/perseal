package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"os"

	"golang.org/x/crypto/pbkdf2"
)

var (
	PBKDF2SALT string = os.Getenv("PBKDF2SALT") // to make
)

// Pbkdf2 Fixes key length for AES Usage
func Pbkdf2(key []byte) []byte {
	log.Println("salt:", string(PBKDF2SALT))
	return pbkdf2.Key(key, []byte(PBKDF2SALT), 4096, 32, sha512.New)
}

func Padding(cipherText []byte, length int) []byte {
	hasher := sha512.New()
	hasher.Write(cipherText)
	ret := hasher.Sum(nil)[:length]
	// log.Println(ret)
	return ret
}

func AESEncrypt(key, plainText []byte) (encmess []byte, err error) {
	// plainText := []byte(message)
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	return cipherText, nil
}

func AESDecrypt(key []byte, securemess string) (decodedmess []byte, err error) {
	cipherText, err := base64.URLEncoding.DecodeString(securemess)
	if err != nil {
		return
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	if len(cipherText) < aes.BlockSize {
		err = errors.New("Ciphertext block size is too short!")
		return
	}

	//IV needs to be unique, but doesn't have to be secure.
	//It's common to put it at the beginning of the ciphertext.
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(cipherText, cipherText)

	// decodedmess = string(cipherText)
	return cipherText, nil
}

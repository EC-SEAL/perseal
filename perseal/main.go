package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
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
	opts := &systemOpts{
		Port: ":" + os.Getenv("PERSEAL_INT_PORT"),
	}
	//	fmt.Println(os.Getenv("AUTH_URL"))
	//	fmt.Println(os.Getenv("AUTH_URL"))
	//	fmt.Println(os.Getenv("REDIRECT_URL"))
	//	fmt.Println(os.Getenv("FETCH_TOKEN_URL"))
	//	fmt.Println(os.Getenv("CREATE_FOLDER_URL"))
	//	fmt.Println(os.Getenv("GET_FOLDER_URL"))
	//	fmt.Println(os.Getenv("SM_ENDPOINT"))

	//	fmt.Println(os.Getenv("GOOGLE_DRIVE_CLIENT"))
	//	fmt.Println(os.Getenv("ONE_DRIVE_CLIENT"))
	//	fmt.Println(os.Getenv("ONE_DRIVE_SCOPES"))
	//	fmt.Println(os.Getenv("GOOGLE_DRIVE"))
	//	fmt.Println(os.Getenv("ONE_DRIVE"))

	//	fmt.Println(os.Getenv("PERSEAL_INT_PORT"))
	//	fmt.Println(os.Getenv("PERSEAL_EXT_PORT"))
	//	fmt.Println(os.Getenv("PERSEAL_EMAIL")
	fmt.Println("Persistent SEAL module running on HTTP port " + os.Getenv("PERSEAL_INT_PORT"))
	r := newRouter()
	err := http.ListenAndServe(opts.Port, r)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

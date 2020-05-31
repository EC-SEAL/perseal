package utils

import (
	"log"

	"github.com/EC-SEAL/perseal/model"
)

// Send Redirect to UI
func SendRedirect(modelRedirect model.RedirectStruct) {
	log.Println("On Service")
	log.Println("working on redirect ", model.Redirect)
	if model.Redirect == nil {
		model.Redirect = make(chan model.RedirectStruct)
	}
	model.Redirect <- modelRedirect
	log.Println("modelredirect ", model.Redirect)
}

// Request Code From UI
func RecieveCode() string {
	log.Println(model.Code)
	if model.Code == nil {
		model.Code = make(chan string)
	}
	code := <-model.Code
	model.Code = nil
	return code
}

func RecievePassword() string {
	model.Password = make(chan string)
	password := <-model.Password
	model.Password = nil
	return password
}

func RecieveCheckFirstAccess() bool {
	log.Println("waiting for ToStore")
	if model.CheckFirstAccess == nil {
		model.CheckFirstAccess = make(chan bool)
	}
	CheckFirstAccess := <-model.CheckFirstAccess

	log.Println(CheckFirstAccess)
	return CheckFirstAccess
}

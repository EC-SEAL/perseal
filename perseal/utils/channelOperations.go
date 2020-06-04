package utils

import (
	"log"
	"time"

	"github.com/EC-SEAL/perseal/model"
)

// Send Redirect to UI
func SendRedirect(modelRedirect model.RedirectStruct) (verification bool) {
	log.Println("On Service")
	log.Println("working on redirect ", model.Redirect)
	if model.Redirect == nil {
		model.Redirect = make(chan model.RedirectStruct)
	}
	if model.End == nil {
		model.End = make(chan bool)
	}

	log.Println("este ", modelRedirect)
	select {
	case model.Redirect <- modelRedirect:
		log.Println("read")
	case end := <-model.End:
		log.Println("failed to send redirect")
		verification = end
	}
	model.End = nil
	log.Println("modelredirect ", model.Redirect)
	return
}

func SendLink(str string) {
	log.Println(str)
	log.Println("working on cliencallback ", model.ClientCallback)
	if model.ClientCallback == nil {
		model.ClientCallback = make(chan string)
		log.Println(model.ClientCallback)
	}
	model.ClientCallback <- str
	log.Println(model.ClientCallback)
}

// Request Code From UI
func RecieveCode() string {
	log.Println(model.Code)
	if model.Code == nil {
		model.Code = make(chan string)
	}
	var code string

	code = <-model.Code
	model.Code = nil
	return code
}

func RecievePassword() (verification bool, password string) {
	log.Println("Recieve Password")
	if model.Password == nil {
		model.Password = make(chan string)
	}
	if model.End == nil {
		log.Println("Made End in Pass")
		model.End = make(chan bool)
	}

	select {
	case password = <-model.Password:
		log.Println("read")
	case end := <-model.End:
		log.Println("failed to read pass")
		verification = end
	}

	model.End = nil
	return verification, password
}

func RecieveCheckFirstAccess() bool {
	log.Println("waiting for ToStore")
	if model.CheckFirstAccess == nil {
		model.CheckFirstAccess = make(chan bool)
	}

	timeout := make(chan bool)
	go func() {
		time.Sleep(1 * time.Second)
		timeout <- true
		log.Println("timeout")

	}()

	var CheckFirstAccess bool
	select {
	case <-timeout:
		CheckFirstAccess = true
	case CheckFirstAccess = <-model.CheckFirstAccess:
		log.Println(CheckFirstAccess)
	}

	model.CheckFirstAccess = nil
	return CheckFirstAccess
}

package model

type DashboardResponse struct {
	Code         int    `json:"code"`
	Message      string `json:"message"`
	ErrorMessage string `json:"error"`
}

type RedirectStruct struct {
	Redirect bool   `json:"redirect"`
	URL      string `json:"url"`
}

var Local = true

var C, Filename, Password, CloudLogin chan string
var Redirect chan RedirectStruct

package controller

import (
	"flag"
	"log"
	"net/http"

	"github.com/EC-SEAL/perseal/model"
	"github.com/EC-SEAL/perseal/services"
	"github.com/EC-SEAL/perseal/sm"
	"github.com/EC-SEAL/perseal/utils"
	"github.com/gorilla/mux"
	"golang.org/x/net/webdav"
)

// Main Entry Point For Front-Channel Operations
func FrontChannelOperations(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)

	method := mux.Vars(r)["method"]
	cipherPassword := getQueryParameter(r, "cipherPassword")
	msToken := getQueryParameter(r, "msToken")

	obj, _, err := initialEPSetup(w, msToken, method, false, cipherPassword)
	if err != nil {
		return
	}
	url := redirectToOperation(obj, w, r)
	if url != "" {
		http.Redirect(w, r, url, http.StatusFound)
	}
}

//Handles DataStore operation (store or load) after password insertion
func DataStoreHandling(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)

	method := mux.Vars(r)["method"]
	msToken := r.FormValue("msToken")

	dto, _, err := initialEPSetup(w, msToken, method, false)
	dto.MenuOption = "BrowserOption"
	if err != nil {
		return
	}

	password := r.FormValue("password")
	if password == "" {
		err = model.BuildResponse(http.StatusBadRequest, model.Messages.NoPassword, dto.ID)
		err.FailedInput = "Password"
		writeResponseMessage(w, dto, *err)
		return
	}
	sha := utils.HashSUM256(password)
	dto.Password = sha

	dto.DataStoreFileName = r.FormValue("dataStoreName")
	sm.UpdateSessionData(dto.ID, dto.DataStoreFileName, "DSFilename")

	var response *model.HTMLResponse
	if dto.Method == model.EnvVariables.Store_Method {
		response, err = services.PersistenceStore(dto)
		if err != nil {
			writeResponseMessage(w, dto, *err)
			return
		}
	} else if dto.Method == model.EnvVariables.Load_Method {

		if dto.PDS == model.EnvVariables.Browser_PDS {
			dto.LocalFileBytes, err = fetchLocalDataStore(r)
			if err != nil {
				writeResponseMessage(w, dto, *err)
				return
			}
		}

		response, err = services.PersistenceLoad(dto)
	} else if dto.Method == model.EnvVariables.Store_Load_Method {
		response, err = services.PersistenceStoreAndLoad(dto)
	}

	if err != nil {
		dto.Files.FileList = err.Files.FileList
		dto.Files.SizeList = err.Files.SizeList
		dto.Files.TimeList = err.Files.TimeList
		writeResponseMessage(w, dto, *err)
		return
	} else if response != nil {
		writeResponseMessage(w, dto, *response)
		return
	}
}

func TestWebDav(w http.ResponseWriter, r *http.Request) {

	dirFlag := flag.String("d", "./", "Directory to serve from. Default is CWD")

	flag.Parse()

	dir := *dirFlag

	srv := &webdav.Handler{
		FileSystem: webdav.Dir(dir),
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			if err != nil {
				log.Printf("WEBDAV [%s]: %s, ERROR: %s\n", r.Method, r.URL, err)
			} else {
				log.Printf("WEBDAV [%s]: %s \n", r.Method, r.URL)
			}
		},
	}

	r.Method = "GET"
	r.URL.Path = ""
	srv.ServeHTTP(w, r)
}

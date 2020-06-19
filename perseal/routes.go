package main

import (
	"net/http"

	"github.com/EC-SEAL/perseal/controller"
	"github.com/gorilla/mux"
)

type route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type routes []route

func newRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	rest := router.PathPrefix("/per").Subrouter()
	for _, route := range perRoutes {
		rest.
			HandleFunc(route.Pattern, route.HandlerFunc).
			Methods(route.Method).
			Name(route.Name)
	}

	staticPaths := map[string]string{
		"ui": "./ui/",
	}

	for pathName, pathValue := range staticPaths {
		pathPrefix := "/" + pathName + "/"
		router.
			PathPrefix(pathPrefix).
			Handler(http.
				StripPrefix(pathPrefix, http.FileServer(http.Dir(pathValue))))
	}

	return router
}

// As per in SPEC https://github.com/EC-SEAL/interface-specs/blob/master/SEAL_Interfaces.yaml
var perRoutes = routes{
	route{
		"Setup a persistence mechanism and load a secure storage into session.",
		"POST",
		"/loadFile",
		controller.PersistenceLoad,
	},
	route{
		"Save session data to the configured persistence mechanism (front channel).",
		"POST",
		"/storeFile",
		controller.PersistenceStore,
	},
	route{
		"Store And Load",
		"POST",
		"/storeAndLoadFile",
		controller.PersistenceStoreAndLoad,
	},
	route{
		"Internal Method to Send Code from Cloud Login to Retrieve the Access Token",
		"GET",
		"/code",
		controller.RetrieveCode,
	},
	route{
		"Internal Method to Send Code from Cloud Login to Retrieve the Access Token",
		"POST",
		"/insertPassword",
		controller.InsertPasswordStoreAndLoad,
	},
	route{
		"Internal Method to Send Code from Cloud Login to Retrieve the Access Token",
		"GET",
		"/save",
		controller.Save,
	},
	route{
		"Intitial Configuration And Main Entry Point For Cloud Operations",
		"GET",
		"/{method}",
		controller.InitialCloudConfig,
	},
	route{
		"Intitial Configuration And Main Entry Point For Local Operations",
		"GET",
		"/{method}/{sessionToken}",
		controller.InitialLocalConfig,
	},
	/*
		route{
			"Generate msToken",
			"GET",
			"/generateToken",
			controller.GenerateToken,
		},
		route{
			"Start Session",
			"POST",
			"/startSession",
			controller.StartSession,
		},

		route{
			"Start Session",
			"GET",
			"/updateLocal",
			controller.UpdateSessionData,
		},
	*/
}

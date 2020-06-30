package main

import (
	"net/http"

	"github.com/EC-SEAL/perseal/controller"
	"github.com/EC-SEAL/perseal/model"
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
	if model.Test {
		testRoutes := routes{
			route{
				"Internal Method to Send Code from Cloud Login to Retrieve the Access Token",
				"GET",
				"/test/session",
				controller.StartSession,
			},
			route{
				"Internal Method to Send Code from Cloud Login to Retrieve the Access Token",
				"GET",
				"/test/token",
				controller.Token,
			},
			route{
				"SimulatesDashboardBehaviour",
				"GET",
				"/test/simulateDashboard",
				controller.SimulateDashboard,
			},
			route{
				"Generate msToken",
				"POST",
				"/test/{method}",
				controller.Test,
			},
		}
		perRoutes = append(testRoutes, perRoutes...)
	}
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

	//internal endpoints
	route{
		"Setup a persistence mechanism and load a secure storage into session.",
		"POST",
		"/insertPassword",
		controller.DataStoreHandling,
	},
	route{
		"Internal Method to Send Code from Cloud Login to Retrieve the Access Token",
		"GET",
		"/aux/{method}",
		controller.Save,
	},

	//external endpoints
	route{
		"Internal Method to Send Code from Cloud Login to Retrieve the Access Token",
		"GET",
		"/code",
		controller.RetrieveCode,
	},

	route{
		"Initial Configuration And Main Entry Point For Cloud Operations",
		"GET",
		"/{method}",
		controller.FrontChannelOperations,
	},

	route{
		"Intitial Configuration And Main Entry Point For Local Operations",
		"POST",
		"/load/{sessionToken}",
		controller.BackChannelDecryption,
	},
}

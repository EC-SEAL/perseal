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
		"Operations after inserting the Password for Store or Load of the DataStore",
		"POST",
		"/insertPassword/{method}",
		controller.DataStoreHandling,
	},
	route{
		"Auxiliary Endpoints",
		"GET",
		"/aux/{method}",
		controller.AuxiliaryEndpoints,
	},

	//external endpoints
	route{
		"Recieves Code from Cloud Login to Retrieve the Access Token",
		"GET",
		"/code",
		controller.RetrieveCode,
	},

	route{
		"Redirects to ClientCallbackAddr",
		"GET",
		"/pollcca",
		controller.PollToClientCallback,
	},

	route{
		"Generates QR code",
		"GET",
		"/QRcode",
		controller.GenerateQRCode,
	},
	route{
		"Initial Configuration And Main Entry Point For Front-Channel Operations",
		"GET",
		"/{method}",
		controller.FrontChannelOperations,
	},

	route{
		"Initial Configuration And Main Entry Point For Back-Channel Loading",
		"POST",
		"/{method}",
		controller.BackChannelLoading,
	},
}

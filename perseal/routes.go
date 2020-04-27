package main

import (
	"net/http"

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
	return router
}

// As per in SPEC https://github.com/EC-SEAL/interface-specs/blob/master/SEAL_Interfaces.yaml
var perRoutes = routes{
	route{
		"Setup a persistence mechanism and load a secure storage into session.",
		"POST",
		"/load",
		persistenceLoad,
	},
	route{
		"Save session data to the configured persistence mechanism (front channel).",
		"POST",
		"/store",
		persistenceStore,
	},
	/*route{
		"Silent setup of a persistence mechanism by loading a user-provided secure storage into session. (back channel).",
		"POST",
		"/load/{sessionToken}",
		persistenceLoadWithToken,
	},*/
	route{
		"Save session data to the configured persistence mechanism (back channel). Might return the signed and possibly encrypted datastore",
		"GET",
		"/store/{sessionToken}",
		persistenceStoreWithToken,
	},
	route{
		"Fetches Code for Access Token",
		"GET",
		"/code",
		getCodeFromDashboard,
	},
	route{
		"Mocks a Dashboard behaviour to send the code to persistence module",
		"GET",
		"/dashboardMock",
		dashboardMock,
	},
}

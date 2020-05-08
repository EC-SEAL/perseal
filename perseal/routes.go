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
	return router
}

// As per in SPEC https://github.com/EC-SEAL/interface-specs/blob/master/SEAL_Interfaces.yaml
var perRoutes = routes{
	route{
		"Setup a persistence mechanism and load a secure storage into session.",
		"POST",
		"/load",
		controller.PersistenceLoad,
	},
	route{
		"Save session data to the configured persistence mechanism (front channel).",
		"POST",
		"/store",
		controller.PersistenceStore,
	},
	route{
		"Silent setup of a persistence mechanism by loading a user-provided secure storage into session. (back channel).",
		"POST",
		"/load/{sessionToken}",
		controller.PersistenceLoadWithToken,
	},
	route{
		"Save session data to the configured persistence mechanism (back channel). Might return the signed and possibly encrypted datastore",
		"GET",
		"/store/{sessionToken}",
		controller.PersistenceStoreWithToken,
	},
	route{
		"Fetches Code for Access Token",
		"GET",
		"/code",
		controller.GetCodeFromDashboard,
	},
	route{
		"Reset",
		"GET",
		"/reset",
		controller.Reset,
	},
}

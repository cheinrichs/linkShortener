package main

import (
	"net/http"
)

//Route contains all the data a router would need to handle a route
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

//Routes holds all the Routes for initializing the router
type Routes []Route

var routes = Routes{
	Route{
		"CreateLink",
		"POST",
		"/createLink",
		createLinkEndpoint,
	},
	Route{
		"LinkStatistics",
		"GET",
		"/linkStatistics/{redirectHash}",
		linkStatisticsEndpoint,
	},
	Route{
		"Redirect",
		"GET",
		"/{redirectHash}",
		redirectEndpoint,
	},
}

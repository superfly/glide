package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func router() http.Handler {
	r := mux.NewRouter()
	v1 := r.PathPrefix("/api/v1").Subrouter()

	v1.Path("/deploy").Methods("POST").Name("deploy").Handler(http.HandlerFunc(deployApp))

	return r
}

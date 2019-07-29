package restApi

import (
	"net/http"

	"github.com/gorilla/mux"
)

func BuildRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/users", ListAllUsers).Methods("GET")
	r.HandleFunc("/users", CreateUser).Methods("POST")

	http.Handle("/", r)
	return r
}

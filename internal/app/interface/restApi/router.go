package restApi

import (
	"net/http"

	"github.com/gorilla/mux"
)

func BuildRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/users", ListAllUsers).Methods("GET")
	r.HandleFunc("/users", CreateUser).Methods("POST")
	r.HandleFunc("/users/{id}", RemoveUser).Methods("DELETE")
	r.HandleFunc("/users/{id}", FindUserByID).Methods("GET")

	http.Handle("/", r)
	return r
}

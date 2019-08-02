package restApi

import (
	"net/http"

	"github.com/gorilla/mux"
)

func BuildRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/{persistance_type}/users", ListAllUsers).Methods("GET")
	r.HandleFunc("/{persistance_type}/users", CreateUser).Methods("POST")
	r.HandleFunc("/{persistance_type}/users/{id}", RemoveUser).Methods("DELETE")
	r.HandleFunc("/{persistance_type}/users/{id}", FindUserByID).Methods("GET")

	http.Handle("/", r)
	return r
}

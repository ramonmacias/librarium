package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ramonmacias/librarium/internal/app/domain/service"
	"github.com/ramonmacias/librarium/internal/app/interface/persistance/memory"
	"github.com/ramonmacias/librarium/internal/app/interface/persistance/postgres"

	"github.com/ramonmacias/librarium/internal/app/usecase"
)

type UserRequestBody struct {
	Email    string `json:email`
	Name     string `json:name`
	LastName string `json:lastName`
}

var (
	memoryInteractor   usecase.UserInteractor
	postgresInteractor usecase.UserInteractor
)

const (
	customPersistanceHeader = "X-Persistance-Type"
)

func init() {
	memoryInteractor = usecase.NewUserInteractor(
		*memory.NewUserController(),
		service.NewUserService(memory.NewUserController()),
	)
	db := postgres.NewClient("localhost", "5432", "ramon", "librarium_database", "ramon_postgres_pass").Connect().DB()
	postgresInteractor = usecase.NewUserInteractor(
		*postgres.NewUserController(db),
		service.NewUserService(postgres.NewUserController(db)),
	)
}

func ListAllUsers(w http.ResponseWriter, r *http.Request) {
	log.Println("Init of ListAllUsers endpoint")
	var err error
	var users []*usecase.User

	switch r.Header.Get(customPersistanceHeader) {
	case "memory":
		users, err = memoryInteractor.ListUser()
	case "postgres":
		users, err = postgresInteractor.ListUser()
	default:
		err = fmt.Errorf("Persistance type not available")
	}
	if err != nil {
		log.Printf("Error while try to find all the users: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	log.Println("Init of Create User endpoint")
	var err error
	userRequest := &UserRequestBody{}
	json.NewDecoder(r.Body).Decode(userRequest)
	switch r.Header.Get(customPersistanceHeader) {
	case "memory":
		err = memoryInteractor.RegisterUser(userRequest.Email, userRequest.Name, userRequest.LastName)
	case "postgres":
		err = postgresInteractor.RegisterUser(userRequest.Email, userRequest.Name, userRequest.LastName)
	default:
		err = fmt.Errorf("Persistance type not available")
	}
	if err != nil {
		log.Printf("Error while try to register a new user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func RemoveUser(w http.ResponseWriter, r *http.Request) {
	log.Println("Init of remove user endpoint")
	var err error
	switch r.Header.Get(customPersistanceHeader) {
	case "memory":
		err = memoryInteractor.RemoveUser(mux.Vars(r)["id"])
	case "postgres":
		err = postgresInteractor.RemoveUser(mux.Vars(r)["id"])
	default:
		err = fmt.Errorf("Persistance type not available")
	}
	if err != nil {
		log.Printf("Error removing a user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func FindUserByID(w http.ResponseWriter, r *http.Request) {
	log.Println("Init of find user by ID endpoint")
	var err error
	var user *usecase.User

	switch r.Header.Get(customPersistanceHeader) {
	case "memory":
		user, err = memoryInteractor.FindByID(mux.Vars(r)["id"])
	case "postgres":
		user, err = postgresInteractor.FindByID(mux.Vars(r)["id"])
	default:
		err = fmt.Errorf("Persistance type not available")
	}
	if err != nil {
		log.Printf("Error trying to find a user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if user == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

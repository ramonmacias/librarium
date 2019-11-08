package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/ramonmacias/librarium/internal/app/domain/service"
	"github.com/ramonmacias/librarium/internal/app/interface/persistence/memory"
	"github.com/ramonmacias/librarium/internal/app/interface/persistence/postgres"

	"github.com/ramonmacias/librarium/internal/app/usecase"
)

type UserRequestBody struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	LastName string `json:"lastName"`
}

var (
	memoryInteractor   usecase.UserInteractor
	postgresInteractor usecase.UserInteractor
)

const (
	customPersistenceHeader = "X-Persistence-Type"
)

func init() {
	memoryInteractor = usecase.NewUserInteractor(
		*memory.NewUserController(),
		service.NewUserService(memory.NewUserController()),
	)
	db := postgres.NewClient(os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_DATABASE"), os.Getenv("POSTGRES_PASSWORD")).Connect().DB()
	postgresInteractor = usecase.NewUserInteractor(
		*postgres.NewUserController(db),
		service.NewUserService(postgres.NewUserController(db)),
	)
}

func ListAllUsers(w http.ResponseWriter, r *http.Request) {
	log.Println("Init of ListAllUsers endpoint")
	var err error
	var users []*usecase.User

	switch r.Header.Get(customPersistenceHeader) {
	case "memory":
		users, err = memoryInteractor.ListUser()
	case "postgres":
		users, err = postgresInteractor.ListUser()
	default:
		err = fmt.Errorf("Persistence type not available")
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
	defer r.Body.Close()

	switch r.Header.Get(customPersistenceHeader) {
	case "memory":
		err = memoryInteractor.RegisterUser(userRequest.Email, userRequest.Name, userRequest.LastName)
	case "postgres":
		err = postgresInteractor.RegisterUser(userRequest.Email, userRequest.Name, userRequest.LastName)
	default:
		err = fmt.Errorf("Persistence type not available")
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
	switch r.Header.Get(customPersistenceHeader) {
	case "memory":
		err = memoryInteractor.RemoveUser(mux.Vars(r)["id"])
	case "postgres":
		err = postgresInteractor.RemoveUser(mux.Vars(r)["id"])
	default:
		err = fmt.Errorf("Persistence type not available")
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

	switch r.Header.Get(customPersistenceHeader) {
	case "memory":
		user, err = memoryInteractor.FindByID(mux.Vars(r)["id"])
	case "postgres":
		user, err = postgresInteractor.FindByID(mux.Vars(r)["id"])
	default:
		err = fmt.Errorf("Persistence type not available")
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

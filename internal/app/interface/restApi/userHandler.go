package restApi

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ramonmacias/librarium/internal/app/domain/service"
	"github.com/ramonmacias/librarium/internal/app/interface/persistance/memory"

	"github.com/ramonmacias/librarium/internal/app/usecase"
)

type UserRequestBody struct {
	Email    string `json:email`
	Name     string `json:name`
	LastName string `json:lastName`
}

var (
	interactor usecase.UserInteractor
)

func init() {
	interactor = usecase.NewUserInteractor(
		*memory.NewUserController(),
		service.NewUserService(memory.NewUserController()),
	)
}

func ListAllUsers(w http.ResponseWriter, r *http.Request) {
	log.Println("Init of ListAllUsers endpoint")
	users, err := interactor.ListUser()
	if err != nil {
		log.Printf("Error fetching all the users: %v", err)
		w.WriteHeader(http.StatusConflict)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	log.Println("Init of Create User endpoint")
	userRequest := &UserRequestBody{}
	json.NewDecoder(r.Body).Decode(userRequest)
	if err := interactor.RegisterUser(userRequest.Email, userRequest.Name, userRequest.LastName); err != nil {
		log.Printf("Error while try to register a new user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func RemoveUser(w http.ResponseWriter, r *http.Request) {
	log.Println("Init of remove user endpoint")
	if err := interactor.RemoveUser(mux.Vars(r)["id"]); err != nil {
		log.Printf("Error removing a user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func FindUserByID(w http.ResponseWriter, r *http.Request) {
	log.Println("Init of find user by ID endpoint")
	user, err := interactor.FindByID(mux.Vars(r)["id"])
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

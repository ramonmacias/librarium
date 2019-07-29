package restApi

import (
	"encoding/json"
	"log"
	"net/http"

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
	log.Println("Init of ListAllUsers")
	users, err := interactor.ListUser()
	if err != nil {
		log.Printf("Error fetching all the users: %v", err)
		w.WriteHeader(http.StatusConflict)
		return
	}
	log.Printf("Users received: %+v", users)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	log.Println("Init of Create User")
	userRequest := &UserRequestBody{}
	json.NewDecoder(r.Body).Decode(userRequest)
	if err := interactor.RegisterUser(userRequest.Email, userRequest.Name, userRequest.LastName); err != nil {
		log.Printf("Error while try to register a new user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusCreated)
}

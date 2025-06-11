package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"librarium/internal/auth"
	"librarium/internal/user"
)

// AuthController holds all the dependencies needed to
// handle all the http requests related with auth domain.
type AuthController struct {
	userRepo user.Repository
}

// NewAuthController builds a new auth controller to handle http requests
// using the given data, all the params received are mandatory.
// It returns an error if some mandatory data is missing.
func NewAuthController(userRepo user.Repository) (*AuthController, error) {
	if userRepo == nil {
		return nil, errors.New("user repository is mandatory for auth controller")
	}

	return &AuthController{userRepo}, nil
}

func (ac *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	loginReq := &auth.LoginRequest{}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(loginReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("error decoding login request")
		return
	}

	librarian, err := ac.userRepo.GetLibrarian(loginReq.LibrarianID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("error getting librarian")
		return
	}
	if librarian == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode("librarian not found")
		return
	}

	session, err := auth.Login(loginReq, librarian)
	if err != nil {
		// TODO: Check on handle differently the errors based on type
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("error loging librarian")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(session)
}

package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"librarium/internal/auth"
	"librarium/internal/onboarding"
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
		WriteResponse(w, http.StatusBadRequest, errors.New("error decoding login request"))
		return
	}

	librarian, err := ac.userRepo.GetLibrarianByEmail(loginReq.Email)
	if err != nil {
		log.Println("error getting librarian while login", err)
		WriteResponse(w, http.StatusInternalServerError, errors.New("error getting librarian"))
		return
	}
	if librarian == nil {
		WriteResponse(w, http.StatusNotFound, errors.New("librarian not found"))
		return
	}

	session, err := auth.Login(loginReq, librarian)
	if err != nil {
		log.Println("error while login", err)
		WriteResponse(w, http.StatusInternalServerError, err)
		return
	}

	WriteResponse(w, http.StatusOK, session)
}

func (ac *AuthController) Signup(w http.ResponseWriter, r *http.Request) {
	librarianRequest := &onboarding.LibrarianRequest{}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(librarianRequest); err != nil {
		log.Println("error decoding request while signup", err)
		WriteResponse(w, http.StatusBadRequest, errors.New("error decoding signup request"))
		return
	}

	librarian, err := ac.userRepo.GetLibrarianByEmail(librarianRequest.Email)
	if err != nil {
		log.Println("error getting librarian while signup", err)
		WriteResponse(w, http.StatusInternalServerError, errors.New("error getting librarian"))
		return
	}
	if librarian != nil {
		WriteResponse(w, http.StatusConflict, errors.New("email already registered"))
		return
	}

	librarian, err = onboarding.Librarian(librarianRequest)
	if err != nil {
		log.Println("error onboarding librarian while signup", err)
		WriteResponse(w, http.StatusBadRequest, err)
		return
	}

	if err := ac.userRepo.CreateLibrarian(librarian); err != nil {
		log.Println("error getting creating librarian while signup", err)
		WriteResponse(w, http.StatusInternalServerError, errors.New("error creating librarian"))
		return
	}

	WriteResponse(w, http.StatusOK, struct {
		ID string `json:"id"`
	}{ID: librarian.ID.String()})
}

// AuthMiddleware checks that we provide an Authorization header with a valid token.
// It returns specific errors in case of non passing the auth checks.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/signup" || r.URL.Path == "/login" {
			next.ServeHTTP(w, r)
			return
		}

		token := r.Header.Get("Authorization")
		if token == "" {
			WriteResponse(w, http.StatusUnauthorized, errors.New("unauthorized"))
			return
		}

		_, err := auth.DecodeAndValidate(token)
		if err != nil {
			WriteResponse(w, http.StatusUnauthorized, errors.New("unauthorized"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

package http

import (
	"encoding/json"
	"net/http"

	"librarium/internal/auth"
	"librarium/internal/user"
)

type AuthController struct {
	userRepo user.Repository
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

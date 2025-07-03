package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

// DecodeRequest decodes a JSON request body into the provided generic type T.
// It returns an error in case of failure.
func DecodeRequest[T any](r *http.Request) (*T, error) {
	if r.Body == nil {
		return nil, errors.New("empty request body")
	}
	defer r.Body.Close()

	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return nil, err
	}
	return &v, nil
}

// WriteResponse writes the provided rsp into the http.ResponseWriter, it handles
// that all the responses we provided keeps the Content-Type header as application/json.
// In case the provided rsp is an error, we return a json response format.
// It logs any error that can happen during the json encoding process.
func WriteResponse(w http.ResponseWriter, statusCode int, rsp any) {
	if rsp == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err, ok := rsp.(error)
	if ok {
		w.WriteHeader(statusCode)
		if err = json.NewEncoder(w).Encode(struct {
			Error string `json:"error"`
		}{Error: err.Error()}); err != nil {
			log.Println("error encoding error message", err)
		}
		return
	}

	w.WriteHeader(statusCode)
	if err = json.NewEncoder(w).Encode(rsp); err != nil {
		log.Println("error encoding response message", err)
	}
}

// JsonContentTypeMiddleware provides an http middleware that checks that all the requests that contains some information in their
// body must be provided with the Content-Type application/json.
// If not a specific error is returned
func JsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if (r.Method == http.MethodPost || r.Method == http.MethodPut) && r.Header.Get("Content-Type") != "application/json" {
			WriteResponse(w, http.StatusBadRequest, errors.New("Content-Type must be application/json"))
		}

		next.ServeHTTP(w, r)
	})
}

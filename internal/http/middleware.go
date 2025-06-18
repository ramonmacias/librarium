package http

import (
	"encoding/json"
	"log"
	"net/http"

	"librarium/internal/auth"
)

func withMiddlewares(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, m := range middlewares {
		h = m(h)
	}
	return h
}

type jsonResponseWriter struct {
	http.ResponseWriter
}

func (w *jsonResponseWriter) Write(b []byte) (int, error) {
	if len(b) > 0 {
		w.Header().Set("Content-Type", "application/json")
	}

	return w.ResponseWriter.Write(b)
}

func jsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if (r.Method == http.MethodPost || r.Method == http.MethodPut) && r.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode("server only accepts application/json content type")
			return
		}

		jw := &jsonResponseWriter{ResponseWriter: w}
		next.ServeHTTP(jw, r)
	})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/signup" || r.URL.Path == "/login" {
			next.ServeHTTP(w, r)
			return
		}

		token := r.Header.Get("Authorization")
		if token == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode("unauthorized")
			return
		}

		_, err := auth.DecodeAndValidate(token)
		if err != nil {
			log.Println("error decoding token", err)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode("unauthorized")
			return
		}

		next.ServeHTTP(w, r)
	})
}

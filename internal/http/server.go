package http

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"time"
)

const (
	_shutdownPeriod = 15 * time.Second
)

type Server struct {
	*http.Server
	address               string
	stopOngoingGracefully context.CancelFunc

	authController     *AuthController
	catalogController  *CatalogController
	customerController *CustomerController
	rentalController   *RentalController
}

// NewServer builds a new http.Server using the provided dependencies.
// All the dependencies provided are mandatory, if we miss some of them an error
// will be returned.
func NewServer(address string, authController *AuthController, catalogController *CatalogController, customerController *CustomerController, rentalController *RentalController) (*Server, error) {
	if address == "" {
		return nil, errors.New("http server address is mandatory")
	}
	if authController == nil {
		return nil, errors.New("auth controller is mandatory")
	}
	if catalogController == nil {
		return nil, errors.New("catalog controller is mandatory")
	}
	if customerController == nil {
		return nil, errors.New("customer controller is mandatory")
	}
	if rentalController == nil {
		return nil, errors.New("rental controller is mandatory")
	}
	srv := &Server{
		address:            address,
		authController:     authController,
		catalogController:  catalogController,
		customerController: customerController,
		rentalController:   rentalController,
	}
	// Ensure in-flight requests aren't cancelled immediately on SIGTERM
	ongoingCtx, stopOngoingGracefully := context.WithCancel(context.Background())
	srv.Server = &http.Server{
		Addr:    address,
		Handler: srv.router(),
		BaseContext: func(_ net.Listener) context.Context {
			return ongoingCtx
		},
	}
	srv.stopOngoingGracefully = stopOngoingGracefully

	return srv, nil
}

func (s *Server) ListenAndServe() {
	go func() {
		log.Println("server starting at ", s.address)
		if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
}

func (s *Server) Shutdown() error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), _shutdownPeriod)
	defer cancel()
	err := s.Server.Shutdown(shutdownCtx)
	s.stopOngoingGracefully()

	return err
}

// router defines all the routing to our API, currently we only allow the librarian
// to access to it, so all the action will be taken by him.
func (s *Server) router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /signup", s.authController.Signup)
	mux.HandleFunc("POST /login", s.authController.Login)

	mux.HandleFunc("POST /catalog/assets", s.catalogController.Create)
	mux.HandleFunc("DELETE /catalog/assets/{id}", s.catalogController.Delete)
	mux.HandleFunc("GET /catalog/assets", s.catalogController.Find)

	mux.HandleFunc("GET /customers", s.customerController.Find)
	mux.HandleFunc("POST /customers", s.customerController.Create)
	mux.HandleFunc("PUT /customers/{id}/suspend", s.customerController.Suspend)
	mux.HandleFunc("PUT /customers/{id}/unsuspend", s.customerController.UnSuspend)

	mux.HandleFunc("GET /rentals", s.rentalController.Find)
	mux.HandleFunc("POST /rentals", s.rentalController.Create)
	mux.HandleFunc("PUT /rentals/{id}/return", s.rentalController.Return)
	mux.HandleFunc("PUT /rentals/{id}/extend", s.rentalController.Extend)

	return withMiddlewares(mux, JsonContentTypeMiddleware, AuthMiddleware)
}

func withMiddlewares(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, m := range middlewares {
		h = m(h)
	}
	return h
}

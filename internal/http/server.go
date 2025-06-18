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
	address               string
	srv                   *http.Server
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

	return &Server{
		address:            address,
		authController:     authController,
		catalogController:  catalogController,
		customerController: customerController,
		rentalController:   rentalController,
	}, nil
}

func (s *Server) ListenAndServe() {
	// Ensure in-flight requests aren't cancelled immediately on SIGTERM
	ongoingCtx, stopOngoingGracefully := context.WithCancel(context.Background())
	s.srv = &http.Server{
		Addr:    ":4000",
		Handler: s.router(),
		BaseContext: func(_ net.Listener) context.Context {
			return ongoingCtx
		},
	}
	s.stopOngoingGracefully = stopOngoingGracefully

	go func() {
		log.Println("server starting at :4000")
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
}

func (s *Server) Shutdown() error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), _shutdownPeriod)
	defer cancel()
	err := s.srv.Shutdown(shutdownCtx)
	s.stopOngoingGracefully()

	return err
}

// router defines all the routing to our API, currently we only allow the librarian
// to access to it, so all the action will be taken by him.
func (s *Server) router() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /signup", s.authController.Signup)
	mux.HandleFunc("POST /login", s.authController.Login)

	mux.HandleFunc("POST /catalog/assets", s.catalogController.CreateCatalogAsset)
	mux.HandleFunc("DELETE /catalog/assets/{id}", s.catalogController.DeleteCatalogAsset)
	mux.HandleFunc("GET /catalog/assets", s.catalogController.FindCatalogAssets)

	// TODO: Refactor controller name methods to avoid redundancy
	mux.HandleFunc("GET /customers", s.customerController.FindCustomers)
	mux.HandleFunc("POST /customers", s.customerController.CreateCustomer)
	mux.HandleFunc("PUT /customers/{id}/suspend", s.customerController.SuspendCustomer)
	mux.HandleFunc("PUT /customers/{id}/unsuspend", s.customerController.UnSuspendCustomer)

	mux.HandleFunc("GET /rentals", s.rentalController.Find)
	mux.HandleFunc("POST /rentals", s.rentalController.Create)
	mux.HandleFunc("PUT /rentals/{id}/return", s.rentalController.Return)
	mux.HandleFunc("PUT /rentals/{id}/extend", s.rentalController.Extend)
	return mux
}

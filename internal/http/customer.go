package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"librarium/internal/onboarding"
	"librarium/internal/user"

	"github.com/google/uuid"
)

// CustomerController holds all the dependencies needed to
// handle all the http requests related with the customer domain.
type CustomerController struct {
	userRepository user.Repository
}

// NewCustomerController builds a new customer controller to handle http requests
// using the given data, all the params received are mandatory.
// It returns an error if some mandatory data is missing.
func NewCustomerController(userRepository user.Repository) (*CustomerController, error) {
	if userRepository == nil {
		return nil, errors.New("user repository is mandatory")
	}

	return &CustomerController{userRepository}, nil
}

func (cc *CustomerController) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	customerReq := &onboarding.CustomerRequest{}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(customerReq); err != nil {
		log.Println("error decoding customer request", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("error decoding customer request")
		return
	}

	customer, err := onboarding.Customer(customerReq)
	if err != nil {
		log.Println("error onboarding customer", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("error onboarding customer")
		return
	}

	if err := cc.userRepository.CreateCustomer(customer); err != nil {
		log.Println("error creating customer", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("error creating customer")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct{ ID string }{ID: customer.ID.String()})
}

func (cc *CustomerController) FindCustomers(w http.ResponseWriter, r *http.Request) {
	customers, err := cc.userRepository.FindCustomers()
	if err != nil {
		log.Println("error finding customers", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("error finding customers")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(customers)
}

func (cc *CustomerController) SuspendCustomer(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) != 3 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("invalida expected path")
		return
	}
	customerID, err := uuid.Parse(parts[1])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("invalid customer ID format, expected UUID")
		return
	}

	customer, err := cc.userRepository.GetCustomer(customerID)
	if err != nil {
		log.Println("error getting customer", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("error getting customer")
		return
	}

	if err := customer.Suspend(); err != nil {
		log.Println("error suspending customer", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("error suspending customer")
		return
	}

	if err := cc.userRepository.UpdateCustomer(customer); err != nil {
		log.Println("error updating customer", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("error updating customer")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cc *CustomerController) UnSuspendCustomer(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) != 3 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("invalida expected path")
		return
	}
	customerID, err := uuid.Parse(parts[1])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("invalid customer ID format, expected UUID")
		return
	}

	customer, err := cc.userRepository.GetCustomer(customerID)
	if err != nil {
		log.Println("error getting customer", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("error getting customer")
		return
	}

	if err := customer.Unsuspend(); err != nil {
		log.Println("error unsuspending customer", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("error unsuspending customer")
		return
	}

	if err := cc.userRepository.UpdateCustomer(customer); err != nil {
		log.Println("error updating customer", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("error updating customer")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

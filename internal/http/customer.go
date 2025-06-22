package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"librarium/internal/onboarding"
	"librarium/internal/query"
	"librarium/internal/user"
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

func (cc *CustomerController) Create(w http.ResponseWriter, r *http.Request) {
	customerReq := &onboarding.CustomerRequest{}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(customerReq); err != nil {
		log.Println("error decoding customer request", err)
		WriteResponse(w, http.StatusBadRequest, errors.New("error decoding customer request"))
		return
	}

	customer, err := onboarding.Customer(customerReq)
	if err != nil {
		log.Println("error onboarding customer", err)
		WriteResponse(w, http.StatusBadRequest, errors.New("error onboarding customer"))
		return
	}

	if err := cc.userRepository.CreateCustomer(customer); err != nil {
		log.Println("error creating customer", err)
		WriteResponse(w, http.StatusInternalServerError, errors.New("error creating customer"))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(struct {
		ID string `json:"id"`
	}{ID: customer.ID.String()})
}

func (cc *CustomerController) Find(w http.ResponseWriter, r *http.Request) {
	pagination, err := query.PaginationFromHTTPRequest(r)
	if err != nil {
		log.Println("error getting pagination from the request", err)
		WriteResponse(w, http.StatusBadRequest, err)
		return
	}

	customers, err := cc.userRepository.FindCustomers(
		query.FiltersFromHTTPRequest(r),
		query.SortingFromHTTPRequest(r),
		pagination,
	)
	if err != nil {
		log.Println("error finding customers", err)
		WriteResponse(w, http.StatusInternalServerError, errors.New("error finding customers"))
		return
	}

	WriteResponse(w, http.StatusOK, customers)
}

func (cc *CustomerController) Suspend(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) != 3 {
		WriteResponse(w, http.StatusBadRequest, errors.New("invalid expected path"))
		return
	}
	customerID, err := uuid.Parse(parts[1])
	if err != nil {
		WriteResponse(w, http.StatusBadRequest, errors.New("invalid customer ID format, expected UUID"))
		return
	}

	customer, err := cc.userRepository.GetCustomer(customerID)
	if err != nil {
		log.Println("error getting customer", err)
		WriteResponse(w, http.StatusBadRequest, errors.New("error getting customer"))
		return
	}

	if err := customer.Suspend(); err != nil {
		log.Println("error suspending customer", err)
		WriteResponse(w, http.StatusBadRequest, errors.New("error suspending customer"))
		return
	}

	if err := cc.userRepository.UpdateCustomer(customer); err != nil {
		log.Println("error updating customer", err)
		WriteResponse(w, http.StatusBadRequest, errors.New("error updating customer"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cc *CustomerController) UnSuspend(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) != 3 {
		WriteResponse(w, http.StatusBadRequest, errors.New("invalid expected path"))
		return
	}
	customerID, err := uuid.Parse(parts[1])
	if err != nil {
		WriteResponse(w, http.StatusBadRequest, errors.New("invalid customer ID format, expected UUID"))
		return
	}

	customer, err := cc.userRepository.GetCustomer(customerID)
	if err != nil {
		log.Println("error getting customer", err)
		WriteResponse(w, http.StatusInternalServerError, errors.New("error getting customer"))
		return
	}

	if err := customer.Unsuspend(); err != nil {
		log.Println("error unsuspending customer", err)
		WriteResponse(w, http.StatusBadRequest, err)
		return
	}

	if err := cc.userRepository.UpdateCustomer(customer); err != nil {
		log.Println("error updating customer", err)
		WriteResponse(w, http.StatusInternalServerError, errors.New("error updating customer"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

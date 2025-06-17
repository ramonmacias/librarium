package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"librarium/internal/catalog"
	"librarium/internal/rental"
	"librarium/internal/user"
)

// RentalController holds all the dependencies needed to
// handle all the http requests related with the rental domain.
type RentalController struct {
	rentalRepository  rental.Repository
	userRepository    user.Repository
	catalogRepository catalog.Repository
}

// NewRentalController builds a new rental controller to handle http requests
// using the given data, all the params received are mandatory.
// It returns an error if some mandatory data is missing.
func NewRentalController(rentalRepository rental.Repository, userRepository user.Repository, catalogRepository catalog.Repository) (*RentalController, error) {
	if rentalRepository == nil {
		return nil, errors.New("rental repository is mandatory")
	}
	if userRepository == nil {
		return nil, errors.New("user repository is mandatory")
	}
	if catalogRepository == nil {
		return nil, errors.New("catalog repository is mandatory")
	}

	return &RentalController{rentalRepository, userRepository, catalogRepository}, nil
}

func (rc *RentalController) Find(w http.ResponseWriter, r *http.Request) {
	rentals, err := rc.rentalRepository.FindRentals()
	if err != nil {
		log.Println("error finding rentals", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("error finding rentals")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rentals)
}

func (rc *RentalController) Create(w http.ResponseWriter, r *http.Request) {
	rentalReq := &rental.RentalRequest{}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(rentalReq); err != nil {
		log.Println("error decoding rental request", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("error decoding rental request")
		return
	}

	customer, err := rc.userRepository.GetCustomer(rentalReq.CustomerID)
	if err != nil {
		log.Println("error getting customer", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("error getting customer")
		return
	}
	asset, err := rc.catalogRepository.GetAsset(rentalReq.AssetID)
	if err != nil {
		log.Println("error getting asset", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("error getting asset")
		return
	}
	activeRental, err := rc.rentalRepository.GetActiveRental(customer.ID, asset.ID)
	if err != nil {
		log.Println("error getting active rental", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("error getting active rental")
		return
	}

	// TODO: Need to define the filters for repositories
	ren, err := rental.Rent(customer, asset, activeRental, nil)
	if err != nil {
		log.Println("error renting asset", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("error renting asset")
		return
	}

	if err := rc.rentalRepository.CreateRental(ren); err != nil {
		log.Println("error creating rental", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("error creating rental")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct{ ID string }{ID: ren.ID.String()})
}

func (rc *RentalController) Return(w http.ResponseWriter, r *http.Request) {

}

func (rc *RentalController) Extend(w http.ResponseWriter, r *http.Request) {

}

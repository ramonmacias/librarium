package http

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"librarium/internal/catalog"
	"librarium/internal/query"
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
	pagination, err := query.PaginationFromHTTPRequest(r)
	if err != nil {
		log.Println("error getting pagination from the request", err)
		WriteResponse(w, http.StatusBadRequest, err)
		return
	}

	rentals, err := rc.rentalRepository.FindRentals(
		query.FiltersFromHTTPRequest(r),
		query.SortingFromHTTPRequest(r),
		pagination,
	)
	if err != nil {
		log.Println("error finding rentals", err)
		WriteResponse(w, http.StatusInternalServerError, errors.New("error finding rentals"))
		return
	}

	WriteResponse(w, http.StatusOK, rentals)
}

func (rc *RentalController) Create(w http.ResponseWriter, r *http.Request) {
	rentalReq, err := DecodeRequest[rental.RentalRequest](r)
	if err != nil {
		log.Println("error decoding rental request", err)
		WriteResponse(w, http.StatusBadRequest, errors.New("error decoding rental request"))
		return
	}

	customer, err := rc.userRepository.GetCustomer(rentalReq.CustomerID)
	if err != nil {
		log.Println("error getting customer", err)
		WriteResponse(w, http.StatusInternalServerError, errors.New("error getting customer"))
		return
	}
	asset, err := rc.catalogRepository.GetAsset(rentalReq.AssetID)
	if err != nil {
		log.Println("error getting asset", err)
		WriteResponse(w, http.StatusInternalServerError, errors.New("error getting asset"))
		return
	}
	activeRental, err := rc.rentalRepository.GetActiveRental(customer.ID, asset.ID)
	if err != nil {
		log.Println("error getting active rental", err)
		WriteResponse(w, http.StatusInternalServerError, errors.New("error getting active rental"))
		return
	}
	customerRentals, err := rc.rentalRepository.FindRentals(query.Filters{
		"customer_id": query.Filter{
			Type:  query.FilterTypeEqual,
			Value: rentalReq.CustomerID.String(),
		},
	}, nil, nil)
	if err != nil {
		log.Println("error getting customer rentals", err)
		WriteResponse(w, http.StatusInternalServerError, errors.New("error getting customer rentals"))
		return
	}

	ren, err := rental.Rent(customer, asset, activeRental, customerRentals)
	if err != nil {
		log.Println("error renting asset", err)
		WriteResponse(w, http.StatusBadRequest, err)
		return
	}

	if err := rc.rentalRepository.CreateRental(ren); err != nil {
		log.Println("error creating rental", err)
		WriteResponse(w, http.StatusInternalServerError, errors.New("error creating rental"))
		return
	}

	WriteResponse(w, http.StatusOK, struct {
		ID string `json:"id"`
	}{ID: ren.ID.String()})
}

func (rc *RentalController) Return(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		WriteResponse(w, http.StatusBadRequest, errors.New("invalid expected path"))
		return
	}
	rentalID, err := uuid.Parse(parts[2])
	if err != nil {
		WriteResponse(w, http.StatusBadRequest, errors.New("invalid rental ID format, expected UUID"))
		return
	}

	ren, err := rc.rentalRepository.GetRental(rentalID)
	if err != nil {
		log.Println("error getting rental", err)
		WriteResponse(w, http.StatusInternalServerError, errors.New("error getting rental"))
		return
	}

	returnedRental, err := rental.Return(ren)
	if err != nil {
		log.Println("error returning rental", err)
		WriteResponse(w, http.StatusBadRequest, errors.New("error returning rental"))
		return
	}

	if err := rc.rentalRepository.UpdateRental(returnedRental); err != nil {
		log.Println("error updating rental", err)
		WriteResponse(w, http.StatusInternalServerError, errors.New("error updating rental"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (rc *RentalController) Extend(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		WriteResponse(w, http.StatusBadRequest, errors.New("invalid expected path"))
		return
	}
	rentalID, err := uuid.Parse(parts[2])
	if err != nil {
		WriteResponse(w, http.StatusBadRequest, errors.New("invalid rental ID format, expected UUID"))
		return
	}

	ren, err := rc.rentalRepository.GetRental(rentalID)
	if err != nil {
		log.Println("error getting rental", err)
		WriteResponse(w, http.StatusInternalServerError, errors.New("error getting rental"))
		return
	}

	extendedRental, err := rental.Extend(ren)
	if err != nil {
		log.Println("error extending rental", err)
		WriteResponse(w, http.StatusBadRequest, err)
		return
	}

	if err := rc.rentalRepository.UpdateRental(extendedRental); err != nil {
		log.Println("error updating rental", err)
		WriteResponse(w, http.StatusInternalServerError, errors.New("error updating rental"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

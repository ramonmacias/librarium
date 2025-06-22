// Package rental provides the domain logic for managing the rental lifecycle
// of catalog assets by library customers.
//
// It defines the Rental entity to represent the relationship between a customer
// and a catalog asset during a defined rental period. The package enforces key
// business rules such as:
//   - Maximum number of concurrent rentals per customer.
//   - Restriction on overdue rentals.
//   - Rental duration and extension limits.
//
// Key features include:
//   - Rent: to initiate a rental, performing all necessary validations.
//   - Return: to mark a rental as completed.
//   - Extend: to allow limited extensions of the rental period.
//
// The package also includes rental status enumeration and constants for rental constraints.
package rental

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"librarium/internal/catalog"
	"librarium/internal/query"
	"librarium/internal/user"
)

const (
	maxNumberOfRentals        = 5
	maxNumberOfExtendedMonths = 3
)

// RentalStatus defines the different states a rental might have
type RentalStatus string

const (
	// RentalStatusActive indicates that the rental is alive and waiting to be returned
	RentalStatusActive RentalStatus = "ACTIVE"
	// RentalStatusReturned indicates that the rental was ended successfully
	RentalStatusReturned RentalStatus = "RETURNED"
	// RentalStatusOverdue indicates that the rental was not returned in the expected due date
	RentalStatusOverdue RentalStatus = "OVERDUE"
)

// RentalRequest
type RentalRequest struct {
	CustomerID uuid.UUID `json:"customer_id"` // Unique customer identifier
	AssetID    uuid.UUID `json:"asset_id"`    // Unique asset identifier
}

// Rental defines the concept of renting an asset catalog, this is the relationship
// between a customer and an asset withing a period of time
type Rental struct {
	ID         uuid.UUID    `json:"id"`          // Unique rental ID
	CustomerID uuid.UUID    `json:"customer_id"` // Reference to the customer who rented the asset
	AssetID    uuid.UUID    `json:"asset_id"`    // Reference to the catalog asset being rented
	RentedAt   time.Time    `json:"rented_at"`   // Timestamp when the rental started
	DueAt      time.Time    `json:"due_at"`      // When the asset is due for return
	ReturnedAt *time.Time   `json:"returned_at"` // When the asset was actually returned (nil if not returned yet)
	Status     RentalStatus `json:"status"`      // Enum: Active, Returned, Overdue, etc.
}

// Rent creates a new rental between the customer and the asset given.
// This function checks if the customer already reached the max number of rentals per customer, if the
// asset catalog is already rented and if the customer has the appropiate status to rent any asset, for all those
// checks the function returns specific errors.
// If not constraints are broken, the function return an active rental and no error.
func Rent(customer *user.Customer, asset *catalog.Asset, activeRental *Rental, customerRentals []*Rental) (*Rental, error) {
	if activeRental != nil {
		return nil, errors.New("catalog asset already rented")
	}
	if len(customerRentals) >= maxNumberOfRentals {
		return nil, errors.New("customer max number of rentals reached")
	}
	for index := range customerRentals {
		if customerRentals[index].Status == RentalStatusOverdue {
			return nil, errors.New("the customer has already a rental in overdue")
		}
	}
	if customer.Status == user.CustomerStatusSuspended {
		return nil, errors.New("cannot rent the asset, customer is suspended")
	}

	return &Rental{
		ID:         uuid.New(),
		CustomerID: customer.ID,
		AssetID:    asset.ID,
		RentedAt:   time.Now(),
		DueAt:      time.Now().AddDate(0, 1, 0), // 1 month per rental
		Status:     RentalStatusActive,
	}, nil
}

// Return function defined the action when we close a rental by returning the asset rented.
// It only checks if the provided rental was already returned, returning an error if so.
// If the function success then we return a rental with the final returned status and
// the returnedAt timestamps updated.
func Return(r *Rental) (*Rental, error) {
	if r.Status == RentalStatusReturned {
		return nil, errors.New("the rental is already returned")
	}

	r.Status = RentalStatusReturned
	now := time.Now().UTC()
	r.ReturnedAt = &now
	return r, nil
}

// Extend function expands the rental time of the provided rental.
// It checks if by trying to expand the rental we reached the max time.
// The extend is in a monthly basis, meaning we will expand one month each
// time we call this function.
// If the rental is already returned or we reach the max number of rental months
// the function returns an error.
// Otherwise we return the rental with the status active and the updated due date.
func Extend(r *Rental) (*Rental, error) {
	if r.Status == RentalStatusReturned {
		return nil, errors.New("the rental is already returned")
	}

	if r.DueAt.AddDate(0, 1, 0).After(r.RentedAt.AddDate(0, maxNumberOfExtendedMonths, 0)) {
		return nil, errors.New("extend max months reached")
	}

	r.Status = RentalStatusActive
	r.DueAt = r.DueAt.AddDate(0, 1, 0)
	return r, nil
}

// Repository defines all the interactions between the rental domain and the persistence layer
type Repository interface {
	// CreateRental inserts the provided rental into the database.
	// It returns an error if something fails.
	CreateRental(rental *Rental) error
	// UpdateRental save the new data into the database with the provided rental.
	// It returns an error if something fails.
	UpdateRental(rental *Rental) error
	// FindRentals retrieves the rentals already persisted into the database.
	// It returns a paginated response, and use the filters in order to return a subset
	// of the rentals based on the provided ones.
	// Uses the sorting to order desc or asc the results.
	// It returns an empty slice and no error in case no rentals found.
	// It returns an error if something fails.
	FindRentals(filters query.Filters, sorting *query.Sorting, pagination *query.Pagination) ([]*Rental, error)
	// GetActiveRental retrieves the rental that matches the provided customer and asset IDs and
	// is in an Active status.
	// It returns nil, nil in case of not found.
	// It returns an error if something fails.
	GetActiveRental(customerID, assetID uuid.UUID) (*Rental, error)
	// GetRental retrieves the rental linked to the provided ID.
	// It returns nil, nil in case not found.
	// It returns an error if something fails.
	GetRental(id uuid.UUID) (*Rental, error)
}

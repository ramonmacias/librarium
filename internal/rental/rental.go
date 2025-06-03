package rental

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"librarium/internal/catalog"
	"librarium/internal/user"
)

const (
	maxNumberOfRentals = 5
)

type RentalStatus string

const (
	RentalStatusActive   RentalStatus = "ACTIVE"
	RentalStatusReturned RentalStatus = "RETURNED"
	RentalStatusOverdue  RentalStatus = "OVERDUE"
)

type Rental struct {
	ID         uuid.UUID    // Unique rental ID
	CustomerID uuid.UUID    // Reference to the customer who rented the asset
	AssetID    uuid.UUID    // Reference to the catalog asset being rented
	RentedAt   time.Time    // Timestamp when the rental started
	DueAt      time.Time    // When the asset is due for return
	ReturnedAt *time.Time   // When the asset was actually returned (nil if not returned yet)
	Status     RentalStatus // Enum: Active, Returned, Overdue, etc.
}

func Rent(customer *user.Customer, asset *catalog.Asset, activeRental *Rental, customerRentals []*Rental) (*Rental, error) {
	if activeRental != nil {
		return nil, errors.New("catalog item already rented")
	}
	if len(customerRentals) >= maxNumberOfRentals {
		return nil, errors.New("customer max number of rentals reached")
	}
	for index := range customerRentals {
		if customerRentals[index].Status == RentalStatusOverdue {
			return nil, errors.New("the customer has already a rental in overdue")
		}
	}
	// TODO: Check on if the customer is not already banned

	return &Rental{
		ID:         uuid.New(),
		CustomerID: customer.ID,
		AssetID:    asset.ID,
		RentedAt:   time.Now(),
		DueAt:      time.Now().AddDate(0, 1, 0), // 1 month per rental
		Status:     RentalStatusActive,
	}, nil
}

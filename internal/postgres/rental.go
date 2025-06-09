package postgres

import (
	"database/sql"
	"errors"
	"fmt"

	"librarium/internal/rental"
)

type rentalRepository struct {
	db *sql.DB
}

// NewRentalRepository builds a new rental.Repository implemented in postgres.
// It returns an error if the provided db connection is nil.
func NewRentalRepository(db *sql.DB) (rental.Repository, error) {
	if db == nil {
		return nil, errors.New("error while building rental repository, db is nil")
	}
	return &rentalRepository{db}, nil
}

// CreateRental inserts the provided rental into the database.
// It returns an error if something fails.
func (rr *rentalRepository) CreateRental(rental *rental.Rental) error {
	_, err := rr.db.Exec(
		"INSERT INTO rentals (id, customer_id, asset_id, rented_at, due_at, returned_at, status) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		rental.ID,
		rental.CustomerID,
		rental.AssetID,
		rental.RentedAt,
		rental.DueAt,
		rental.ReturnedAt,
		rental.Status,
	)
	if err != nil {
		return fmt.Errorf("error inserting rental in postgres %w", err)
	}

	return nil
}

// UpdateRental save the new data into the database with the provided rental.
// It returns an error if something fails.
func (rr *rentalRepository) UpdateRental(rental *rental.Rental) error {
	_, err := rr.db.Exec(
		"UPDATE rentals SET customer_id = $1, asset_id = $2, rented_at = $3, due_at = $4, returned_at = $5, status = $6 WHERE id = $7",
		rental.CustomerID,
		rental.AssetID,
		rental.RentedAt,
		rental.DueAt,
		rental.ReturnedAt,
		rental.Status,
		rental.ID,
	)
	if err != nil {
		return fmt.Errorf("error inserting rental in postgres %w", err)
	}

	return nil
}

// FindRentals retrieves the rentals already persisted into the database.
// It returns an empty slice and no error in case no rentals found.
// It returns an error if something fails.
func (rr *rentalRepository) FindRentals() ([]*rental.Rental, error) {
	rows, err := rr.db.Query("SELECT id, customer_id, asset_id, rented_at, due_at, returned_at, status FROM rentals")
	if err != nil {
		return nil, fmt.Errorf("error querying for finding rentals %w", err)
	}
	defer rows.Close()

	rentals := []*rental.Rental{}
	for rows.Next() {
		rental := &rental.Rental{}

		if err := rows.Scan(rental.ID, rental.CustomerID, rental.AssetID, rental.RentedAt, rental.DueAt, rental.ReturnedAt, rental.Status); err != nil {
			return nil, fmt.Errorf("error scanning while finding rentals %w", err)
		}

		rentals = append(rentals, rental)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error while going through the rentals rows %w", err)
	}
	return rentals, nil
}

package postgres

import (
	"database/sql"
	"errors"
	"fmt"

	"librarium/internal/user"
)

type userRepository struct {
	db *sql.DB
}

// NewUserRepository builds a new user.Repository implemented in postgres.
// It returns an error if the provided db connection is nil.
func NewUserRepository(db *sql.DB) (user.Repository, error) {
	if db == nil {
		return nil, errors.New("error while building user repository, db is nil")
	}
	return &userRepository{db}, nil
}

// CreateLibrarian inserts the provided librarian into the system.
// It returns an error in case of failure.
func (us *userRepository) CreateLibrarian(librarian *user.Librarian) error {
	_, err := us.db.Exec(
		"INSERT INTO librarians (id, name, email, password) VALUES ($1, $2, $3, $4)",
		librarian.ID.String(),
		librarian.Name,
		librarian.Email,
		librarian.Password,
	)
	if err != nil {
		return fmt.Errorf("error inserting librarian in postgres %w", err)
	}

	return nil
}

// CreateCustomer inserts the provided customer into the system.
// It returns an error in case of failure.
func (us *userRepository) CreateCustomer(customer *user.Customer) error {
	_, err := us.db.Exec(
		"INSERT INTO customers (id, name, last_name, national_id, email, phone_number, street, city, state, postal_code, country) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
		customer.ID.String(),
		customer.Name,
		customer.LastName,
		customer.NationalID,
		customer.ContactDetails.Email,
		customer.ContactDetails.PhoneNumber,
		customer.ContactDetails.Address.Street,
		customer.ContactDetails.Address.City,
		customer.ContactDetails.Address.State,
		customer.ContactDetails.Address.PostalCode,
		customer.ContactDetails.Address.Country,
	)
	if err != nil {
		return fmt.Errorf("error inserting customer in postgres %w", err)
	}

	return nil
}

// GetLibrarianByEmail retrieves the librarian linked to the provided email.
// It return nil, nil in case we can't find the librarian.
// It returns an error in case of failure.
func (us *userRepository) GetLibrarianByEmail(email string) (*user.Librarian, error) {
	librarian := &user.Librarian{}
	err := us.db.QueryRow("SELECT id, name, email, password FROM librarians WHERE email = $1", email).Scan(&librarian.ID, &librarian.Name, &librarian.Email, &librarian.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting librarian, with email %s err %w", email, err)
	}

	return librarian, nil
}

// FindCustomers retrieves all the customers from the system.
// It returns an empty slice and no error in case of not found.
// It returns an error in case of failure.
func (us *userRepository) FindCustomers() ([]*user.Customer, error) {
	return nil, nil
}

// Package onboarding provides functionality to register new users into the system,
// including librarians and customers.
//
// It defines structured request types for both user roles (LibrarianRequest and CustomerRequest)
// and exposes functions to create corresponding domain entities by delegating to the `user`
// and `auth` packages for password handling and object construction.
//
// Specifically, the package includes:
//   - LibrarianRequest and CustomerRequest types to represent incoming onboarding data.
//   - Librarian and Customer functions to process the data, perform necessary validations,
//     and return the appropriate user type (librarian or customer).
//
// This package acts as a dedicated layer to handle the onboarding flow in a clean,
// validated, and consistent way.
package onboarding

import (
	"librarium/internal/auth"
	"librarium/internal/user"
)

// LibrarianRequest defines the needed data at the moment we want to
// onboard a new librarian into the system.
type LibrarianRequest struct {
	Name     string `json:"name"`     // Name of the librarian
	Email    string `json:"email"`    // Email of the librarian to be used for auth
	Password string `json:"password"` // Password linked to the provided info to handle auth
}

// Librarian function receives an onboarding.LibrarianRequest, call the auth
// package to handle the hashing of the password and it builds a new librarian.
// It returns an error if hashing the password flow fails or if we have some missing
// data at the moment we build the librarian.
// It returns no error and librarian object created otherwise.
func Librarian(req *LibrarianRequest) (*user.Librarian, error) {
	hashedPass, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	return user.BuildLibrarian(req.Name, req.Email, hashedPass)
}

// CustomerRequest defines the needed data at the moment we want to
// onboard a new customer into the system.
type CustomerRequest struct {
	Name        string // Name of the customer
	LastName    string // LastName of the customer
	NationalID  string // National identificator, for example DNI in Spain
	Email       string // Email address
	PhoneNumber string // Phone number (can include country code)
	Street      string // Street name and number
	City        string // City name
	State       string // State or province
	PostalCode  string // Postal or ZIP code
	Country     string // Country name
}

// Customer function receives an onboarding.CustomerRequest, this is used to call the builder
// from the user package so we are able to create a new customer in the system.
// It returns an error if any mandatory data is missing.
// It returns no error and the customer object created otherwise.
func Customer(req *CustomerRequest) (*user.Customer, error) {
	return user.BuildCustomer(req.Name, req.LastName, req.NationalID, &user.ContactDetails{
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		Address: &user.Address{
			Street:     req.Street,
			City:       req.City,
			State:      req.State,
			PostalCode: req.PostalCode,
			Country:    req.Country,
		},
	})
}

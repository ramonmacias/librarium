// Package user defines the core domain entities and logic for managing users
// of the library system, including both customers and librarians.
//
// The package supports:
//   - Representation and validation of library customers, including personal
//     identification and contact information.
//   - Representation and creation of librarians with secure credential handling.
//   - Structs for user-related metadata such as contact details and physical addresses,
//     each with validation methods to ensure data integrity.
//
// Builders for both Customer and Librarian enforce necessary constraints to guarantee
// that required information is provided before persisting or using user records in the system.
package user

import (
	"errors"

	"github.com/google/uuid"
)

// CustomerStatus defines the different customer statuses
type CustomerStatus string

const (
	// CustomerStatusActive determines when the customer is active, this means is allowed
	// to perform all the available actions for customers.
	CustomerStatusActive CustomerStatus = "ACTIVE"
	// CustomerStatusSuspended determines when the customer is suspended, this means is blocked
	// from performing any previous available action for the customer.
	CustomerStatusSuspended CustomerStatus = "SUSPENDED"
)

// Customer is the person who wants to benefit from the Library by
// being able to read physically in there or rent any of the
// Library catalog items available.
type Customer struct {
	ID             uuid.UUID       // Unique identifier
	Name           string          // Name of the customer
	LastName       string          // LastName of the customer
	Status         CustomerStatus  // Determines if the customer is active or suspended
	NationalID     string          // National identificator, for example DNI in Spain
	ContactDetails *ContactDetails // Contact details for the customer
}

// Suspend performs the action on changing the status of the customer from active to suspended
func (c *Customer) Suspend() error {
	if c.Status == CustomerStatusSuspended {
		return errors.New("customer already suspended")
	}

	c.Status = CustomerStatusSuspended
	return nil
}

// Unsuspend performs the action on changing the status of the customer from the suspended to active
func (c *Customer) Unsuspend() error {
	if c.Status != CustomerStatusSuspended {
		return errors.New("customer should be suspended to be unsuspend")
	}

	c.Status = CustomerStatusActive
	return nil
}

// BuildCustomer generates a new Customer using the given data.
// It returns an error if any specific mandatory data is missing.
// It returns no error and the customer object created in case of success.
func BuildCustomer(name, lastName, nationalID string, contactDetails *ContactDetails) (*Customer, error) {
	if name == "" {
		return nil, errors.New("customer name field is mandatory")
	}
	if lastName == "" {
		return nil, errors.New("customer last name field is mandatory")
	}
	if nationalID == "" {
		return nil, errors.New("customer national ID field is mandatory")
	}
	if contactDetails == nil {
		return nil, errors.New("customer contact details is mandatory")
	}
	if err := contactDetails.validate(); err != nil {
		return nil, err
	}
	return &Customer{
		ID:             uuid.New(),
		Name:           name,
		LastName:       lastName,
		NationalID:     nationalID,
		ContactDetails: contactDetails,
	}, nil
}

// Librarian is the Library's administrator, the one in charge
// that is able to manage the library catalog and the customers
// registered into the platform.
type Librarian struct {
	ID       uuid.UUID // Unique identifier
	Name     string    // Name of the librarian
	Email    string    // Email of the librarian
	Password string    // Password used by the librarian to auth
}

// BuildLibrarian generates a new Librarian using the given data.
// It returns an error if the provided data is missing or empty.
// It returns no error and the librarian object created in case of success.
func BuildLibrarian(name, email, password string) (*Librarian, error) {
	if name == "" {
		return nil, errors.New("librarian name field is mandatory")
	}
	if email == "" {
		return nil, errors.New("librarian email field is mandatory")
	}
	if password == "" {
		return nil, errors.New("librarian password field is mandatory")
	}
	return &Librarian{
		ID:       uuid.New(),
		Name:     name,
		Email:    email,
		Password: password,
	}, nil
}

// Address holds the physical address information.
type Address struct {
	Street     string // Street name and number
	City       string // City name
	State      string // State or province
	PostalCode string // Postal or ZIP code
	Country    string // Country name
}

func (a *Address) validate() error {
	if a.Street == "" {
		return errors.New("address street field is mandatory")
	}
	if a.City == "" {
		return errors.New("address city field is mandatory")
	}
	if a.State == "" {
		return errors.New("address state field is mandatory")
	}
	if a.PostalCode == "" {
		return errors.New("address postal code field is mandatory")
	}
	if a.Country == "" {
		return errors.New("address country field is mandatory")
	}
	return nil
}

// ContactDetails holds different ways to contact a person.
type ContactDetails struct {
	Email       string   // Email address
	PhoneNumber string   // Phone number (can include country code)
	Address     *Address // Physical address
}

func (cd *ContactDetails) validate() error {
	if cd.Email == "" {
		return errors.New("contact details email is mandatory")
	}
	if cd.PhoneNumber == "" {
		return errors.New("contact details phone number is mandatory")
	}
	if cd.Address == nil {
		return errors.New("contact details physical addres is mandatory")
	}
	if err := cd.Address.validate(); err != nil {
		return err
	}
	return nil
}

// Repository defines all the interactions between the user domain and the persistence layer
type Repository interface {
	// CreateLibrarian inserts the provided librarian into the system.
	// It returns an error in case of failure
	CreateLibrarian(librarian *Librarian) error
	// CreateCustomer inserts the provided customer into the system.
	// It returns an error in case of failure
	CreateCustomer(customer *Customer) error
	// UpdateCustomer sets the provided customer data as updated data for the provided customer.
	// It returns an error in case of failure.
	UpdateCustomer(customer *Customer) error
	// GetLibrarianByEmail retrieves the librarian linked to the provided email.
	// It returns nil, nil in case we can't find the librarian.
	// It returns an error in case of failure.
	GetLibrarianByEmail(email string) (*Librarian, error)
	// GetCustomer retrieves the customer linked to the provided customer.
	// It returns nil, nil in case the customer is not found.
	// It returns an error in case of failure.
	GetCustomer(id uuid.UUID) (*Customer, error)
	// FindCustomers retrieves all the customers from the system.
	// It returns an empty slice and no error in case of not found.
	// It returns an error in case of failure.
	FindCustomers() ([]*Customer, error)
}

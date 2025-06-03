package user

import (
	"errors"

	"github.com/google/uuid"
)

// Customer is the person who wants to benefit from the Library by
// being able to read physically in there or rent any of the
// Library catalog items available.
type Customer struct {
	ID             uuid.UUID       // Unique identifier
	Name           string          // Name of the customer
	LastName       string          // LastName of the customer
	NationalID     string          // National identificator, for example DNI in Spain
	ContactDetails *ContactDetails // Contact details for the customer
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
	ID   uuid.UUID
	Name string
}

// BuildLibrarian generates a new Librarian using the given data.
// It returns an error if the provided data is missing or empty.
// It returns no error and the librarian object created in case of success.
func BuildLibrarian(name string) (*Librarian, error) {
	if name == "" {
		return nil, errors.New("librarian name field is mandatory")
	}
	return &Librarian{
		ID:   uuid.New(),
		Name: name,
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

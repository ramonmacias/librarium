package onboarding

import "librarium/internal/user"

type LibrarianRequest struct {
	Name string // Name of the librarian
}

func Librarian(req *LibrarianRequest) (*user.Librarian, error) {
	// TODO: We should add in here all the auth handling so we generate a bearer token.
	return user.BuildLibrarian(req.Name)
}

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

func Customer(req *CustomerRequest) (*user.Customer, error) {
	// TODO: We should add in here all the auth handling so we generate a bearer token.
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

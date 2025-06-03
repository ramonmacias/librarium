package user

import (
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID             uuid.UUID
	Name           string
	LastName       string
	NationalID     string
	ContactDetails *ContactDetails
}

type Librarian struct {
	ID   uuid.UUID
	Name string
}

// Address holds the physical address information.
type Address struct {
	Street     string // Street name and number
	City       string // City name
	State      string // State or province
	PostalCode string // Postal or ZIP code
	Country    string // Country name
}

// ContactDetails holds different ways to contact a person.
type ContactDetails struct {
	Email       string  // Email address
	PhoneNumber string  // Phone number (can include country code)
	Address     Address // Physical address
}

type VirtualID struct {
	Barcode   string // barcode representation (could be a code string or encoded data)
	IssuedAt  time.Time
	ExpiresAt *time.Time // optional expiry date
}

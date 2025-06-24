package onboarding_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"librarium/internal/onboarding"
	"librarium/internal/user"
)

func TestOnboardLibrarian(t *testing.T) {
	testCases := map[string]struct {
		librarianReq    *onboarding.LibrarianRequest
		expectedErr     error
		assertLibrarian func(librarian *user.Librarian)
	}{
		"it should return an error if we have an error while hashing the password": {
			librarianReq: &onboarding.LibrarianRequest{
				Password: "",
			},
			expectedErr:     errors.New("cannot hash an empty password"),
			assertLibrarian: func(librarian *user.Librarian) {},
		},
		"it should return an error if the librarian request misses some mandatory field": {
			librarianReq: &onboarding.LibrarianRequest{
				Email:    "john.doe@test.com",
				Password: "test-password",
			},
			expectedErr:     errors.New("librarian name field is mandatory"),
			assertLibrarian: func(librarian *user.Librarian) {},
		},
		"it should return no error and the expected librarian created": {
			librarianReq: &onboarding.LibrarianRequest{
				Name:     "John Doe",
				Email:    "john.doe@test.com",
				Password: "test-password",
			},
			assertLibrarian: func(librarian *user.Librarian) {
				assert.Equal(t, "John Doe", librarian.Name)
				assert.Equal(t, "john.doe@test.com", librarian.Email)
				assert.NotZero(t, librarian.Password)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			librarian, err := onboarding.Librarian(tc.librarianReq)
			assert.Equal(t, tc.expectedErr, err)
			tc.assertLibrarian(librarian)
		})
	}
}

func TestOnboardCustomer(t *testing.T) {
	testCases := map[string]struct {
		customerReq    *onboarding.CustomerRequest
		expectedErr    error
		assertCustomer func(customer *user.Customer)
	}{
		"it should return an error if any missing customer field is missing": {
			customerReq:    &onboarding.CustomerRequest{},
			expectedErr:    errors.New("customer name field is mandatory"),
			assertCustomer: func(customer *user.Customer) {},
		},
		"it should return no error and the expected customer created": {
			customerReq: &onboarding.CustomerRequest{
				Name:        "John",
				LastName:    "Smith",
				NationalID:  "3349938",
				Email:       "john.smith@test.com",
				PhoneNumber: "+34 76898959",
				Street:      "Street DF",
				City:        "New York",
				State:       "NY",
				PostalCode:  "000394",
				Country:     "US",
			},
			assertCustomer: func(customer *user.Customer) {
				assert.NotZero(t, customer.ID)
				assert.Equal(t, "John", customer.Name)
				assert.Equal(t, "Smith", customer.LastName)
				assert.Equal(t, "3349938", customer.NationalID)
				assert.Equal(t, "john.smith@test.com", customer.ContactDetails.Email)
				assert.Equal(t, "+34 76898959", customer.ContactDetails.PhoneNumber)
				assert.Equal(t, "Street DF", customer.ContactDetails.Address.Street)
				assert.Equal(t, "New York", customer.ContactDetails.Address.City)
				assert.Equal(t, "NY", customer.ContactDetails.Address.State)
				assert.Equal(t, "000394", customer.ContactDetails.Address.PostalCode)
				assert.Equal(t, "US", customer.ContactDetails.Address.Country)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			customer, err := onboarding.Customer(tc.customerReq)
			assert.Equal(t, tc.expectedErr, err)
			tc.assertCustomer(customer)
		})
	}
}

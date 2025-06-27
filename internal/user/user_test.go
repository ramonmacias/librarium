package user_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"librarium/internal/user"
)

func TestBuildCustomer(t *testing.T) {
	testCases := map[string]struct {
		name           string
		lastName       string
		nationalID     string
		contactDetails *user.ContactDetails
		expectedErr    error
		assertCustomer func(customer *user.Customer)
	}{
		"it should return an error if the customer name is missing": {
			expectedErr: errors.New("customer name field is mandatory"),
			assertCustomer: func(customer *user.Customer) {
				assert.Nil(t, customer)
			},
		},
		"it should return an error if the customer last name is missing": {
			name:        "John",
			expectedErr: errors.New("customer last name field is mandatory"),
			assertCustomer: func(customer *user.Customer) {
				assert.Nil(t, customer)
			},
		},
		"it should return an error if the customer national ID is missing": {
			name:        "John",
			lastName:    "Smith",
			expectedErr: errors.New("customer national ID field is mandatory"),
			assertCustomer: func(customer *user.Customer) {
				assert.Nil(t, customer)
			},
		},
		"it should return an error if the customer email is missing": {
			name:           "John",
			lastName:       "Smith",
			nationalID:     "45869584-M",
			contactDetails: &user.ContactDetails{},
			expectedErr:    errors.New("contact details email is mandatory"),
			assertCustomer: func(customer *user.Customer) {
				assert.Nil(t, customer)
			},
		},
		"it should return an error if the customer phone number is missing": {
			name:           "John",
			lastName:       "Smith",
			nationalID:     "45869584-M",
			contactDetails: &user.ContactDetails{Email: "john.smith@test.com"},
			expectedErr:    errors.New("contact details phone number is mandatory"),
			assertCustomer: func(customer *user.Customer) {
				assert.Nil(t, customer)
			},
		},
		"it should return an error if the customer address is missing": {
			name:       "John",
			lastName:   "Smith",
			nationalID: "45869584-M",
			contactDetails: &user.ContactDetails{
				Email:       "john.smith@test.com",
				PhoneNumber: "+34 678987564",
			},
			expectedErr: errors.New("contact details physical address is mandatory"),
			assertCustomer: func(customer *user.Customer) {
				assert.Nil(t, customer)
			},
		},
		"it should return an error if the street address is missing": {
			name:       "John",
			lastName:   "Smith",
			nationalID: "45869584-M",
			contactDetails: &user.ContactDetails{
				Email:       "john.smith@test.com",
				PhoneNumber: "+34 678987564",
				Address:     &user.Address{},
			},
			expectedErr: errors.New("address street field is mandatory"),
			assertCustomer: func(customer *user.Customer) {
				assert.Nil(t, customer)
			},
		},
		"it should return an error if the city address is missing": {
			name:       "John",
			lastName:   "Smith",
			nationalID: "45869584-M",
			contactDetails: &user.ContactDetails{
				Email:       "john.smith@test.com",
				PhoneNumber: "+34 678987564",
				Address: &user.Address{
					Street: "c/ green",
				},
			},
			expectedErr: errors.New("address city field is mandatory"),
			assertCustomer: func(customer *user.Customer) {
				assert.Nil(t, customer)
			},
		},
		"it should return an error if the state address is missing": {
			name:       "John",
			lastName:   "Smith",
			nationalID: "45869584-M",
			contactDetails: &user.ContactDetails{
				Email:       "john.smith@test.com",
				PhoneNumber: "+34 678987564",
				Address: &user.Address{
					Street: "c/ green",
					City:   "Barcelona",
				},
			},
			expectedErr: errors.New("address state field is mandatory"),
			assertCustomer: func(customer *user.Customer) {
				assert.Nil(t, customer)
			},
		},
		"it should return an error if the postal code address is missing": {
			name:       "John",
			lastName:   "Smith",
			nationalID: "45869584-M",
			contactDetails: &user.ContactDetails{
				Email:       "john.smith@test.com",
				PhoneNumber: "+34 678987564",
				Address: &user.Address{
					Street: "c/ green",
					City:   "Barcelona",
					State:  "Barcelona",
				},
			},
			expectedErr: errors.New("address postal code field is mandatory"),
			assertCustomer: func(customer *user.Customer) {
				assert.Nil(t, customer)
			},
		},
		"it should return an error if the country address is missing": {
			name:       "John",
			lastName:   "Smith",
			nationalID: "45869584-M",
			contactDetails: &user.ContactDetails{
				Email:       "john.smith@test.com",
				PhoneNumber: "+34 678987564",
				Address: &user.Address{
					Street:     "c/ green",
					City:       "Barcelona",
					State:      "Barcelona",
					PostalCode: "17645",
				},
			},
			expectedErr: errors.New("address country field is mandatory"),
			assertCustomer: func(customer *user.Customer) {
				assert.Nil(t, customer)
			},
		},
		"it should return no error and the customer created": {
			name:       "John",
			lastName:   "Smith",
			nationalID: "45869584-M",
			contactDetails: &user.ContactDetails{
				Email:       "john.smith@test.com",
				PhoneNumber: "+34 678987564",
				Address: &user.Address{
					Street:     "c/ green",
					City:       "Barcelona",
					State:      "Barcelona",
					PostalCode: "17645",
					Country:    "ES",
				},
			},
			assertCustomer: func(customer *user.Customer) {
				assert.NotNil(t, customer)
				assert.NotZero(t, customer.ID)
				assert.Equal(t, "John", customer.Name)
				assert.Equal(t, "Smith", customer.LastName)
				assert.Equal(t, "45869584-M", customer.NationalID)
				assert.Equal(t, "john.smith@test.com", customer.ContactDetails.Email)
				assert.Equal(t, "+34 678987564", customer.ContactDetails.PhoneNumber)
				assert.Equal(t, "c/ green", customer.ContactDetails.Address.Street)
				assert.Equal(t, "Barcelona", customer.ContactDetails.Address.City)
				assert.Equal(t, "Barcelona", customer.ContactDetails.Address.State)
				assert.Equal(t, "17645", customer.ContactDetails.Address.PostalCode)
				assert.Equal(t, "ES", customer.ContactDetails.Address.Country)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			customer, err := user.BuildCustomer(tc.name, tc.lastName, tc.nationalID, tc.contactDetails)
			assert.Equal(t, tc.expectedErr, err)
			tc.assertCustomer(customer)
		})
	}
}

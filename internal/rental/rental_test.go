package rental_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"librarium/internal/catalog"
	"librarium/internal/rental"
	"librarium/internal/user"
)

func TestRent(t *testing.T) {
	expectedCustomerID := uuid.New()
	expectedAssetID := uuid.New()

	testCases := map[string]struct {
		customer        *user.Customer
		asset           *catalog.Asset
		activeRental    *rental.Rental
		customerRentals []*rental.Rental
		expectedErr     error
		assertRental    func(re *rental.Rental)
	}{
		"it should return an error if the asset is already rented": {
			activeRental: &rental.Rental{
				CustomerID: uuid.New(),
				AssetID:    uuid.New(),
			},
			expectedErr:  errors.New("catalog asset already rented"),
			assertRental: func(re *rental.Rental) {},
		},
		"it should return an error if we reach the mas number of rentals per customer": {
			customerRentals: []*rental.Rental{
				{
					CustomerID: expectedCustomerID,
					AssetID:    uuid.New(),
				},
				{
					CustomerID: expectedCustomerID,
					AssetID:    uuid.New(),
				},
				{
					CustomerID: expectedCustomerID,
					AssetID:    uuid.New(),
				},
				{
					CustomerID: expectedCustomerID,
					AssetID:    uuid.New(),
				},
				{
					CustomerID: expectedCustomerID,
					AssetID:    uuid.New(),
				},
			},
			expectedErr:  errors.New("customer max number of rentals reached"),
			assertRental: func(re *rental.Rental) {},
		},
		"it should return an error if the customer has some rental on overdue": {
			customerRentals: []*rental.Rental{
				{
					CustomerID: expectedCustomerID,
					AssetID:    uuid.New(),
					Status:     rental.RentalStatusOverdue,
				},
			},
			expectedErr:  errors.New("the customer has already a rental in overdue"),
			assertRental: func(re *rental.Rental) {},
		},
		"it should return an error if the customer is suspended": {
			customer: &user.Customer{
				ID:     expectedCustomerID,
				Status: user.CustomerStatusSuspended,
			},
			expectedErr:  errors.New("cannot rent the asset, customer is suspended"),
			assertRental: func(re *rental.Rental) {},
		},
		"it should create a rental between the customer an asset provided": {
			customer: &user.Customer{
				ID:     expectedCustomerID,
				Status: user.CustomerStatusActive,
			},
			asset: &catalog.Asset{
				ID: expectedAssetID,
			},
			assertRental: func(re *rental.Rental) {
				assert.NotZero(t, re.ID)
				assert.Equal(t, expectedAssetID, re.AssetID)
				assert.Equal(t, expectedCustomerID, re.CustomerID)
				assert.WithinDuration(t, time.Now().UTC(), re.RentedAt, 1*time.Second)
				assert.Nil(t, re.ReturnedAt)
				assert.Equal(t, time.Now().UTC().AddDate(0, 1, 0).Month(), re.DueAt.Month())
				assert.Equal(t, rental.RentalStatusActive, re.Status)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			re, err := rental.Rent(tc.customer, tc.asset, tc.activeRental, tc.customerRentals)
			assert.Equal(t, tc.expectedErr, err)
			tc.assertRental(re)
		})
	}
}

func TestRentalReturn(t *testing.T) {
	testCases := map[string]struct {
		rental       *rental.Rental
		expectedErr  error
		assertRental func(re *rental.Rental)
	}{
		"it should return an error if the rental is already returned": {
			rental: &rental.Rental{
				Status: rental.RentalStatusReturned,
			},
			expectedErr:  errors.New("the rental is already returned"),
			assertRental: func(re *rental.Rental) {},
		},
		"it should return the rental updated and marked as returned": {
			rental: &rental.Rental{
				Status: rental.RentalStatusActive,
			},
			assertRental: func(re *rental.Rental) {
				assert.Equal(t, rental.RentalStatusReturned, re.Status)
				assert.WithinDuration(t, *re.ReturnedAt, time.Now().UTC(), 1*time.Second)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			re, err := rental.Return(tc.rental)
			assert.Equal(t, tc.expectedErr, err)
			tc.assertRental(re)
		})
	}
}

func TestRentalExtend(t *testing.T) {
	testCases := map[string]struct {
		rental       *rental.Rental
		expectedErr  error
		assertRental func(re *rental.Rental)
	}{
		"it should return an error if the rental is already returned": {
			rental: &rental.Rental{
				Status: rental.RentalStatusReturned,
			},
			expectedErr:  errors.New("the rental is already returned"),
			assertRental: func(re *rental.Rental) {},
		},
		"it should return an error if we try to extend over the max rental months": {
			rental: &rental.Rental{
				Status:   rental.RentalStatusActive,
				RentedAt: time.Now().UTC(),
				DueAt:    time.Now().UTC().AddDate(0, 5, 0),
			},
			expectedErr:  errors.New("extend max months reached"),
			assertRental: func(re *rental.Rental) {},
		},
		"it should extend by 1 month the rental provided": {
			rental: &rental.Rental{
				Status:   rental.RentalStatusActive,
				RentedAt: time.Now().UTC(),
				DueAt:    time.Now().UTC().AddDate(0, 1, 0),
			},
			assertRental: func(re *rental.Rental) {
				assert.Equal(t, rental.RentalStatusActive, re.Status)
				assert.Equal(t, re.DueAt.Month(), time.Now().UTC().AddDate(0, 2, 0).Month())
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			re, err := rental.Extend(tc.rental)
			assert.Equal(t, tc.expectedErr, err)
			tc.assertRental(re)
		})
	}
}

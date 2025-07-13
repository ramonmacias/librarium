package http_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"librarium/internal/http"
)

func TestNewServer(t *testing.T) {
	testCases := map[string]struct {
		address            string
		authController     *http.AuthController
		catalogController  *http.CatalogController
		customerController *http.CustomerController
		rentalController   *http.RentalController
		expectedErr        error
		assertServer       func(srv *http.Server)
	}{
		"it should return an error if the address is missing": {
			expectedErr: errors.New("http server address is mandatory"),
			assertServer: func(srv *http.Server) {
				assert.Nil(t, srv)
			},
		},
		"it should return an error if the auth controller is missing": {
			address:     ":8080",
			expectedErr: errors.New("auth controller is mandatory"),
			assertServer: func(srv *http.Server) {
				assert.Nil(t, srv)
			},
		},
		"it should return an error if the catalog controller is missing": {
			address:        ":8080",
			authController: &http.AuthController{},
			expectedErr:    errors.New("catalog controller is mandatory"),
			assertServer: func(srv *http.Server) {
				assert.Nil(t, srv)
			},
		},
		"it should return an error if the customer controller is missing": {
			address:           ":8080",
			authController:    &http.AuthController{},
			catalogController: &http.CatalogController{},
			expectedErr:       errors.New("customer controller is mandatory"),
			assertServer: func(srv *http.Server) {
				assert.Nil(t, srv)
			},
		},
		"it should return an error if the rental controller is missing": {
			address:            ":8080",
			authController:     &http.AuthController{},
			catalogController:  &http.CatalogController{},
			customerController: &http.CustomerController{},
			expectedErr:        errors.New("rental controller is mandatory"),
			assertServer: func(srv *http.Server) {
				assert.Nil(t, srv)
			},
		},
		"it should return no error": {
			address:            ":8080",
			authController:     &http.AuthController{},
			catalogController:  &http.CatalogController{},
			customerController: &http.CustomerController{},
			rentalController:   &http.RentalController{},
			assertServer: func(srv *http.Server) {
				assert.NotNil(t, srv)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			srv, err := http.NewServer(tc.address, tc.authController, tc.catalogController, tc.customerController, tc.rentalController)
			assert.Equal(t, tc.expectedErr, err)
			tc.assertServer(srv)
		})
	}
}

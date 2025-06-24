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

package auth_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"librarium/internal/auth"
	"librarium/internal/user"
)

func TestLogin(t *testing.T) {
	t.Setenv("AUTH_SIGNING_KEY", "test_key")
	expectedLibrarianID := uuid.New()
	expectedPass, err := auth.HashPassword("expected_password")
	assert.Nil(t, err)

	testCases := map[string]struct {
		loginReq      *auth.LoginRequest
		librarian     *user.Librarian
		expectedErr   error
		assertSession func(session *auth.Session)
	}{
		"it should return an error if the provided email doesn't matche the librarian's email": {
			loginReq: &auth.LoginRequest{
				Email: "john@test.com",
			},
			librarian: &user.Librarian{
				Email: "test@test.com",
			},
			expectedErr:   errors.New("login bad credentials"),
			assertSession: func(session *auth.Session) {},
		},
		"it should return an error if we can't match the passwords": {
			loginReq: &auth.LoginRequest{
				Email:    "john@test.com",
				Password: "000433929",
			},
			librarian: &user.Librarian{
				Email:    "john@test.com",
				Password: "123Test",
			},
			expectedErr:   errors.New("login bad credentials"),
			assertSession: func(session *auth.Session) {},
		},
		"it should return no error and expected session created": {
			loginReq: &auth.LoginRequest{
				Email:    "john@test.com",
				Password: "expected_password",
			},
			librarian: &user.Librarian{
				ID:       expectedLibrarianID,
				Email:    "john@test.com",
				Password: expectedPass,
			},
			assertSession: func(session *auth.Session) {
				assert.Equal(t, expectedLibrarianID, session.LibrarianID)
				assert.NotZero(t, session.Token)
				expectedExpired := time.Now().UTC().Add(4 * time.Hour)
				assert.Equal(t, expectedExpired.Hour(), session.ExpiresAt.Hour())
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			session, err := auth.Login(tc.loginReq, tc.librarian)
			assert.Equal(t, tc.expectedErr, err)
			tc.assertSession(session)
		})
	}
}

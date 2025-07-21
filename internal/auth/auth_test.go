package auth_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"librarium/internal/auth"
	"librarium/internal/user"
)

func TestLogin(t *testing.T) {
	expectedLibrarianID := uuid.New()
	expectedPass, err := auth.HashPassword("expected_password")
	assert.Nil(t, err)

	testCases := map[string]struct {
		loginReq      *auth.LoginRequest
		librarian     *user.Librarian
		signingKey    string
		expectedErr   error
		assertSession func(session *auth.Session)
	}{
		"it should return an error if the provided email doesn't matche the librarian's email": {
			signingKey: "test_key",
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
			signingKey: "test_key",
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
		"it should return an error if we fail while signing the token": {
			signingKey: "",
			loginReq: &auth.LoginRequest{
				Email:    "john@test.com",
				Password: "expected_password",
			},
			librarian: &user.Librarian{
				ID:       expectedLibrarianID,
				Email:    "john@test.com",
				Password: expectedPass,
			},
			expectedErr:   errors.New("error signing token"),
			assertSession: func(session *auth.Session) {},
		},
		"it should return no error and expected session created": {
			signingKey: "test_key",
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
			t.Setenv("AUTH_SIGNING_KEY", tc.signingKey)
			session, err := auth.Login(tc.loginReq, tc.librarian)
			assert.Equal(t, tc.expectedErr, err)
			tc.assertSession(session)
		})
	}
}

func TestDecodeAndValidate(t *testing.T) {
	t.Setenv("AUTH_SIGNING_KEY", "test_key")
	expectedUserID := uuid.New()

	testCases := map[string]struct {
		token          func() string
		expectedErr    string
		expectedUserID uuid.UUID
	}{
		"it should return an error if the token is expired": {
			token: func() string {
				tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
					Subject:   expectedUserID.String(),
					Issuer:    "librarium",
					IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
					ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(-2 * time.Hour)),
				})
				signedtok, err := tok.SignedString([]byte("test_key"))
				assert.Nil(t, err)
				return signedtok
			},
			expectedErr: fmt.Sprintf("error parsing token %s: %s", jwt.ErrTokenInvalidClaims, jwt.ErrTokenExpired),
		},
		"it should return an error if we don't have a subject claim": {
			token: func() string {
				tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
					Subject:   "",
					Issuer:    "librarium",
					IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
					ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(2 * time.Hour)),
				})
				signedtok, err := tok.SignedString([]byte("test_key"))
				assert.Nil(t, err)
				return signedtok
			},
			expectedErr: "missing subject claim while decoding token",
		},
		"it should decode the token and return the expected user ID": {
			token: func() string {
				tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
					Subject:   expectedUserID.String(),
					Issuer:    "librarium",
					IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
					ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(2 * time.Hour)),
				})
				signedtok, err := tok.SignedString([]byte("test_key"))
				assert.Nil(t, err)
				return signedtok
			},
			expectedUserID: expectedUserID,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			userID, err := auth.DecodeAndValidate(tc.token())
			if tc.expectedErr != "" {
				assert.EqualError(t, err, tc.expectedErr)
			}
			assert.Equal(t, tc.expectedUserID, userID)
		})
	}
}

func TestHashPassword(t *testing.T) {
	expectedPassword := "test-password"

	testCases := map[string]struct {
		password         string
		expectedErr      error
		assertHashedPass func(hashedPass string)
	}{
		"it should return an error if the password is empty": {
			expectedErr:      errors.New("cannot hash an empty password"),
			assertHashedPass: func(hashedPass string) {},
		},
		"it should return an error if the password length is bigger than 72": {
			password:         "this_is_a_very_long_password_that_exceeds_seventy_two_characters_for_testing_purposes_123",
			expectedErr:      bcrypt.ErrPasswordTooLong,
			assertHashedPass: func(hashedPass string) {},
		},
		"it should return the expected hashed password": {
			password: expectedPassword,
			assertHashedPass: func(hashedPass string) {
				err := bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte(expectedPassword))
				assert.Nil(t, err)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			hashedPass, err := auth.HashPassword(tc.password)
			assert.Equal(t, tc.expectedErr, err)
			tc.assertHashedPass(hashedPass)
		})
	}
}

func TestCheckPassword(t *testing.T) {
	testCases := map[string]struct {
		hashedPass  func() string
		plainPass   string
		expectedErr error
	}{
		"it should return an error if the passwords are not the same": {
			plainPass: "test-password",
			hashedPass: func() string {
				hashBytes, err := bcrypt.GenerateFromPassword([]byte("a-different-test-password"), bcrypt.DefaultCost)
				assert.Nil(t, err)
				return string(hashBytes)
			},
			expectedErr: bcrypt.ErrMismatchedHashAndPassword,
		},
		"it should return no error if both passwords are equal": {
			plainPass: "test-password",
			hashedPass: func() string {
				hashBytes, err := bcrypt.GenerateFromPassword([]byte("test-password"), bcrypt.DefaultCost)
				assert.Nil(t, err)
				return string(hashBytes)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := auth.CheckPassword(tc.hashedPass(), tc.plainPass)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}

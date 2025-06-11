// Package auth provides authentication utilities for the Librarium system,
// including JWT-based session generation and validation for librarian users.
//
// It includes the following key features:
//   - Login flow that issues a signed JWT token containing librarian identity.
//   - Session representation for tracking login state and expiration.
//   - Secure password hashing and validation using bcrypt.
//   - Token decoding and validation utilities to authenticate client requests.
//
// The JWT token is signed using a secret defined in the AUTH_SIGNING_KEY environment variable,
// and includes standard claims such as subject (librarian ID), issuer, issued at, and expiration time.
package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"librarium/internal/user"
)

const (
	expiryDuration = 4 * time.Hour
)

// LoginRequest defines the json payload needed to receive from
// the client in order to trigger the login flow.
type LoginRequest struct {
	Email    string `json:"email"`    // Email to be used as authentication
	Password string `json:"password"` // Password linked to the email to validate the auth
}

// Session holds the information that represents an auth session in the system.
// The token value will be used to validate the interaction between server and client
// and to hold some basic information in the claims.
// The token is baked using the JWT protocol.
// This struct will be used as well as the login endpoint response.
type Session struct {
	LibrarianID uuid.UUID `json:"librarian_id"` // Unique librarian identifier
	Token       string    `json:"token"`        // Generate JWT token for handle auth
	ExpiresAt   time.Time `json:"expires_at"`   // Moment when the token will be invalid
}

// Login runs the login flow for the provided librarian, the function
// will generate a session containing a jwt encrypted using the AUTH_SIGNING_KEY
// envvar.
// It returns an error if we fail while signing the token.
func Login(loginReq *LoginRequest, librarian *user.Librarian) (s *Session, err error) {
	if loginReq.Email != librarian.Email || CheckPassword(librarian.Password, loginReq.Password) != nil {
		return nil, errors.New("login bad credentials")
	}

	s = &Session{
		LibrarianID: librarian.ID,
		ExpiresAt:   time.Now().UTC().Add(expiryDuration),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   librarian.ID.String(),
		Issuer:    "librarium",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(s.ExpiresAt),
	})
	s.Token, err = tok.SignedString([]byte(os.Getenv("AUTH_SIGNING_KEY")))
	if err != nil {
		return nil, fmt.Errorf("error signing token %w", err)
	}
	return s, nil
}

// DecodeAndValidate receives an encrypted jwt and it runs the parsing
// mecanism so we can validate and decode the claims.
// The function returns the librarian ID.
// The function returns an error if any validation fails in relation with the provided token.
func DecodeAndValidate(tokenStr string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("AUTH_SIGNING_KEY")), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return uuid.Nil, fmt.Errorf("error parsing token %w", err)
	}

	if claims, ok := token.Claims.(jwt.RegisteredClaims); ok {
		return uuid.MustParse(claims.Subject), nil
	}
	return uuid.Nil, errors.New("error while parsing jwt registered claims")
}

// HashPassword hashes a plain password using bcrypt.
func HashPassword(password string) (string, error) {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashBytes), err
}

// CheckPassword compares a plain password with the hashed password.
func CheckPassword(hashedPwd, plainPwd string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(plainPwd))
}

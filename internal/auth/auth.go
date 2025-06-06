package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	expiryDuration = 2 * time.Hour
)

// LoginRequest defines the json payload needed to receive from
// the client in order to trigger the login flow.
type LoginRequest struct {
	LibrarianID uuid.UUID `json:"librarian_id"`
	Email       string    `json:"email"`
	Password    string    `json:"password"`
}

// Session holds the information that represents an auth session in the system.
// The token value will be used to validate the interaction between server and client
// and to hold some basic information in the claims.
// The token is baked using the JWT protocol.
// This struct will be used as well as the login endpoint response.
type Session struct {
	LibrarianID uuid.UUID `json:"librarian_id"`
	Token       string    `json:"token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// Login runs the login flow for the provided librarian, the function
// will generate a session containing a jwt encrypted using the AUTH_SIGNING_KEY
// envvar.
// It returns an error if we fail while signing the token.
func Login(librarianID uuid.UUID) (s *Session, err error) {
	s = &Session{
		LibrarianID: librarianID,
		ExpiresAt:   time.Now().UTC().Add(expiryDuration),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   librarianID.String(),
		Issuer:    "librarium",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(s.ExpiresAt),
	})
	s.Token, err = tok.SignedString(os.Getenv("AUTH_SIGNING_KEY"))
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

package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	stdHttp "net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"librarium/internal/auth"
	"librarium/internal/http"
	"librarium/internal/mocks"
	"librarium/internal/onboarding"
	"librarium/internal/user"
)

func TestNewAuthController(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepo := mocks.NewMockUserRepository(ctrl)

	testCases := map[string]struct {
		userRepo         user.Repository
		expectedErr      error
		assertController func(controller *http.AuthController)
	}{
		"it should return an error if the user repository is missing": {
			expectedErr: errors.New("user repository is mandatory for auth controller"),
			assertController: func(controller *http.AuthController) {
				assert.Nil(t, controller)
			},
		},
		"it should return the controller created and no error": {
			userRepo: userRepo,
			assertController: func(controller *http.AuthController) {
				assert.NotNil(t, controller)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			controller, err := http.NewAuthController(tc.userRepo)
			assert.Equal(t, tc.expectedErr, err)
			tc.assertController(controller)
		})
	}
}

func TestAuthControllerLogin(t *testing.T) {
	t.Setenv("AUTH_SIGNING_KEY", "test_key")
	ctrl := gomock.NewController(t)
	userRepo := mocks.NewMockUserRepository(ctrl)
	controller, err := http.NewAuthController(userRepo)
	assert.Nil(t, err)
	assert.NotNil(t, controller)

	testCases := map[string]struct {
		mocks              func()
		request            func() *stdHttp.Request
		expectedStatusCode int
		assertBody         func(body io.Reader)
	}{
		"it should return an error if the json decoding fails": {
			mocks: func() {},
			request: func() *stdHttp.Request {
				return httptest.NewRequest(stdHttp.MethodPost, "/login", bytes.NewReader([]byte("wrong input")))
			},
			expectedStatusCode: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				errorMsg := struct {
					Error string `json:"error"`
				}{}
				err := json.NewDecoder(body).Decode(&errorMsg)
				assert.Nil(t, err)
				assert.Equal(t, "error decoding login request", errorMsg.Error)
			},
		},
		"it should return an error if getting the librarian repo call fails": {
			mocks: func() {
				userRepo.EXPECT().GetLibrarianByEmail("john.doe@test.com").Return(nil, errors.New("expected error"))
			},
			request: func() *stdHttp.Request {
				loginReq := auth.LoginRequest{Email: "john.doe@test.com", Password: "strong-pass"}
				buf, err := json.Marshal(&loginReq)
				assert.Nil(t, err)
				return httptest.NewRequest(stdHttp.MethodPost, "/login", bytes.NewReader(buf))
			},
			expectedStatusCode: stdHttp.StatusInternalServerError,
			assertBody: func(body io.Reader) {
				errorMsg := struct {
					Error string `json:"error"`
				}{}
				err := json.NewDecoder(body).Decode(&errorMsg)
				assert.Nil(t, err)
				assert.Equal(t, "error getting librarian", errorMsg.Error)
			},
		},
		"it should return an error if we cannot found the librarian": {
			mocks: func() {
				userRepo.EXPECT().GetLibrarianByEmail("john.doe@test.com").Return(nil, nil)
			},
			request: func() *stdHttp.Request {
				loginReq := auth.LoginRequest{Email: "john.doe@test.com", Password: "strong-pass"}
				buf, err := json.Marshal(&loginReq)
				assert.Nil(t, err)
				return httptest.NewRequest(stdHttp.MethodPost, "/login", bytes.NewReader(buf))
			},
			expectedStatusCode: stdHttp.StatusNotFound,
			assertBody: func(body io.Reader) {
				errorMsg := struct {
					Error string `json:"error"`
				}{}
				err := json.NewDecoder(body).Decode(&errorMsg)
				assert.Nil(t, err)
				assert.Equal(t, "librarian not found", errorMsg.Error)
			},
		},
		"it should return an error if we cannot login": {
			mocks: func() {
				userRepo.EXPECT().GetLibrarianByEmail("john.doe@test.com").Return(&user.Librarian{Password: "awesome-password"}, nil)
			},
			request: func() *stdHttp.Request {
				loginReq := auth.LoginRequest{Email: "john.doe@test.com", Password: "strong-pass"}
				buf, err := json.Marshal(&loginReq)
				assert.Nil(t, err)
				return httptest.NewRequest(stdHttp.MethodPost, "/login", bytes.NewReader(buf))
			},
			expectedStatusCode: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				errorMsg := struct {
					Error string `json:"error"`
				}{}
				err := json.NewDecoder(body).Decode(&errorMsg)
				assert.Nil(t, err)
				assert.Equal(t, "login bad credentials", errorMsg.Error)
			},
		},
		"it should return no error and the session generated": {
			mocks: func() {
				hashedPass, err := auth.HashPassword("awesome-password")
				assert.Nil(t, err)
				userRepo.EXPECT().GetLibrarianByEmail("john.doe@test.com").Return(&user.Librarian{ID: uuid.New(), Email: "john.doe@test.com", Password: hashedPass}, nil)
			},
			request: func() *stdHttp.Request {
				loginReq := auth.LoginRequest{Email: "john.doe@test.com", Password: "awesome-password"}
				buf, err := json.Marshal(&loginReq)
				assert.Nil(t, err)
				return httptest.NewRequest(stdHttp.MethodPost, "/login", bytes.NewReader(buf))
			},
			expectedStatusCode: stdHttp.StatusOK,
			assertBody: func(body io.Reader) {
				session := &auth.Session{}
				err := json.NewDecoder(body).Decode(&session)
				assert.Nil(t, err)
				assert.NotZero(t, session.Token)
				assert.NotZero(t, session.LibrarianID)
				assert.NotZero(t, session.ExpiresAt)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			tc.mocks()
			rec := httptest.NewRecorder()
			controller.Login(rec, tc.request())

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatusCode, res.StatusCode)
			tc.assertBody(res.Body)
		})
	}
}

func TestAuthSignup(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepo := mocks.NewMockUserRepository(ctrl)
	controller, err := http.NewAuthController(userRepo)
	assert.Nil(t, err)
	assert.NotNil(t, controller)

	testCases := map[string]struct {
		mocks              func()
		request            func() *stdHttp.Request
		expectedStatusCode int
		assertBody         func(body io.Reader)
	}{
		"it should return an error if the json decoding fails": {
			mocks: func() {},
			request: func() *stdHttp.Request {
				return httptest.NewRequest(stdHttp.MethodPost, "/signup", bytes.NewReader([]byte("wrong input")))
			},
			expectedStatusCode: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				errorMsg := struct {
					Error string `json:"error"`
				}{}
				err := json.NewDecoder(body).Decode(&errorMsg)
				assert.Nil(t, err)
				assert.Equal(t, "error decoding signup request", errorMsg.Error)
			},
		},
		"it should return an error if getting the librarian repo call fails": {
			mocks: func() {
				userRepo.EXPECT().GetLibrarianByEmail("john.doe@test.com").Return(nil, errors.New("expected error"))
			},
			request: func() *stdHttp.Request {
				onboardingReq := onboarding.LibrarianRequest{Name: "John Doe", Email: "john.doe@test.com", Password: "strong-pass"}
				buf, err := json.Marshal(&onboardingReq)
				assert.Nil(t, err)
				return httptest.NewRequest(stdHttp.MethodPost, "/signup", bytes.NewReader(buf))
			},
			expectedStatusCode: stdHttp.StatusInternalServerError,
			assertBody: func(body io.Reader) {
				errorMsg := struct {
					Error string `json:"error"`
				}{}
				err := json.NewDecoder(body).Decode(&errorMsg)
				assert.Nil(t, err)
				assert.Equal(t, "error getting librarian", errorMsg.Error)
			},
		},
		"it should return an error if the librarian's email already exist": {
			mocks: func() {
				userRepo.EXPECT().GetLibrarianByEmail("john.doe@test.com").Return(&user.Librarian{Name: "John Doe", Email: "john.doe@test.com"}, nil)
			},
			request: func() *stdHttp.Request {
				onboardingReq := onboarding.LibrarianRequest{Name: "John Doe", Email: "john.doe@test.com", Password: "strong-pass"}
				buf, err := json.Marshal(&onboardingReq)
				assert.Nil(t, err)
				return httptest.NewRequest(stdHttp.MethodPost, "/signup", bytes.NewReader(buf))
			},
			expectedStatusCode: stdHttp.StatusConflict,
			assertBody: func(body io.Reader) {
				errorMsg := struct {
					Error string `json:"error"`
				}{}
				err := json.NewDecoder(body).Decode(&errorMsg)
				assert.Nil(t, err)
				assert.Equal(t, "email already registered", errorMsg.Error)
			},
		},
		"it should return an error if the onboarding librarian domain function fails": {
			mocks: func() {
				userRepo.EXPECT().GetLibrarianByEmail("john.doe@test.com").Return(nil, nil)
			},
			request: func() *stdHttp.Request {
				onboardingReq := onboarding.LibrarianRequest{Name: "John Doe", Email: "john.doe@test.com", Password: ""}
				buf, err := json.Marshal(&onboardingReq)
				assert.Nil(t, err)
				return httptest.NewRequest(stdHttp.MethodPost, "/signup", bytes.NewReader(buf))
			},
			expectedStatusCode: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				errorMsg := struct {
					Error string `json:"error"`
				}{}
				err := json.NewDecoder(body).Decode(&errorMsg)
				assert.Nil(t, err)
				assert.Equal(t, "cannot hash an empty password", errorMsg.Error)
			},
		},
		"it should return an error if calling create librarian repo call fails": {
			mocks: func() {
				userRepo.EXPECT().GetLibrarianByEmail("john.doe@test.com").Return(nil, nil)
				librarianMatch := mocks.CustomMatcher{
					Match: func(x any) bool {
						librarian, ok := x.(*user.Librarian)
						if !ok {
							return false
						}

						return assert.Equal(t, "John Doe", librarian.Name) &&
							assert.Equal(t, "john.doe@test.com", librarian.Email) &&
							assert.NotZero(t, librarian.ID) &&
							assert.NotZero(t, librarian.Password)
					},
				}
				userRepo.EXPECT().CreateLibrarian(librarianMatch).Return(errors.New("expected error"))
			},
			request: func() *stdHttp.Request {
				onboardingReq := onboarding.LibrarianRequest{Name: "John Doe", Email: "john.doe@test.com", Password: "strong-password"}
				buf, err := json.Marshal(&onboardingReq)
				assert.Nil(t, err)
				return httptest.NewRequest(stdHttp.MethodPost, "/signup", bytes.NewReader(buf))
			},
			expectedStatusCode: stdHttp.StatusInternalServerError,
			assertBody: func(body io.Reader) {
				errorMsg := struct {
					Error string `json:"error"`
				}{}
				err := json.NewDecoder(body).Decode(&errorMsg)
				assert.Nil(t, err)
				assert.Equal(t, "error creating librarian", errorMsg.Error)
			},
		},
		"it should return no error and the librarian onboarded": {
			mocks: func() {
				userRepo.EXPECT().GetLibrarianByEmail("john.doe@test.com").Return(nil, nil)
				librarianMatch := mocks.CustomMatcher{
					Match: func(x any) bool {
						librarian, ok := x.(*user.Librarian)
						if !ok {
							return false
						}

						return assert.Equal(t, "John Doe", librarian.Name) &&
							assert.Equal(t, "john.doe@test.com", librarian.Email) &&
							assert.NotZero(t, librarian.ID) &&
							assert.NotZero(t, librarian.Password)
					},
				}
				userRepo.EXPECT().CreateLibrarian(librarianMatch).Return(nil)
			},
			request: func() *stdHttp.Request {
				onboardingReq := onboarding.LibrarianRequest{Name: "John Doe", Email: "john.doe@test.com", Password: "strong-password"}
				buf, err := json.Marshal(&onboardingReq)
				assert.Nil(t, err)
				return httptest.NewRequest(stdHttp.MethodPost, "/signup", bytes.NewReader(buf))
			},
			expectedStatusCode: stdHttp.StatusOK,
			assertBody: func(body io.Reader) {
				librarianOnboardedMsg := struct {
					ID string `json:"id"`
				}{}
				err := json.NewDecoder(body).Decode(&librarianOnboardedMsg)
				assert.Nil(t, err)
				assert.NotZero(t, librarianOnboardedMsg.ID)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			tc.mocks()
			rec := httptest.NewRecorder()
			controller.Signup(rec, tc.request())

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatusCode, res.StatusCode)
			tc.assertBody(res.Body)
		})
	}
}

func TestAuthMiddleware(t *testing.T) {
	t.Setenv("AUTH_SIGNING_KEY", "awesome-siging-key")
	mux := stdHttp.NewServeMux()
	mux.HandleFunc("/librarian", func(w stdHttp.ResponseWriter, r *stdHttp.Request) {
		w.WriteHeader(stdHttp.StatusOK)
	})
	mux.HandleFunc("/login", func(w stdHttp.ResponseWriter, r *stdHttp.Request) {
		w.WriteHeader(stdHttp.StatusOK)
	})
	mux.HandleFunc("/signup", func(w stdHttp.ResponseWriter, r *stdHttp.Request) {
		w.WriteHeader(stdHttp.StatusOK)
	})
	ts := httptest.NewServer(http.AuthMiddleware(mux))
	defer ts.Close()

	testCases := map[string]struct {
		request        func() *stdHttp.Request
		assertResponse func(rsp *stdHttp.Response)
	}{
		"it should return no error if call the signup endpoint without Authorization header": {
			request: func() *stdHttp.Request {
				req, err := stdHttp.NewRequest(stdHttp.MethodGet, ts.URL+"/signup", stdHttp.NoBody)
				assert.Nil(t, err)
				return req
			},
			assertResponse: func(rsp *stdHttp.Response) {
				assert.NotNil(t, rsp)
				assert.Equal(t, stdHttp.StatusOK, rsp.StatusCode)
			},
		},
		"it should return no error if call the login endpoint without Authorization header": {
			request: func() *stdHttp.Request {
				req, err := stdHttp.NewRequest(stdHttp.MethodGet, ts.URL+"/login", stdHttp.NoBody)
				assert.Nil(t, err)
				return req
			},
			assertResponse: func(rsp *stdHttp.Response) {
				assert.NotNil(t, rsp)
				assert.Equal(t, stdHttp.StatusOK, rsp.StatusCode)
			},
		},
		"it should return an error if the Authorization header is missing": {
			request: func() *stdHttp.Request {
				req, err := stdHttp.NewRequest(stdHttp.MethodGet, ts.URL+"/librarian", stdHttp.NoBody)
				assert.Nil(t, err)
				return req
			},
			assertResponse: func(rsp *stdHttp.Response) {
				assert.NotNil(t, rsp)
				assert.Equal(t, stdHttp.StatusUnauthorized, rsp.StatusCode)
				errorMsg := struct {
					Error string `json:"error"`
				}{}
				err := json.NewDecoder(rsp.Body).Decode(&errorMsg)
				assert.Nil(t, err)
				err = rsp.Body.Close()
				assert.Nil(t, err)
				assert.Equal(t, "unauthorized", errorMsg.Error)
			},
		},
		"it should return an error if the Authorization header is given but with a wrong token": {
			request: func() *stdHttp.Request {
				req, err := stdHttp.NewRequest(stdHttp.MethodGet, ts.URL+"/librarian", stdHttp.NoBody)
				assert.Nil(t, err)
				req.Header.Add("Authorization", "awesome-token")
				return req
			},
			assertResponse: func(rsp *stdHttp.Response) {
				assert.NotNil(t, rsp)
				assert.Equal(t, stdHttp.StatusUnauthorized, rsp.StatusCode)
				errorMsg := struct {
					Error string `json:"error"`
				}{}
				err := json.NewDecoder(rsp.Body).Decode(&errorMsg)
				assert.Nil(t, err)
				err = rsp.Body.Close()
				assert.Nil(t, err)
				assert.Equal(t, "unauthorized", errorMsg.Error)
			},
		},
		"it should reach the handle if the provided token is valid": {
			request: func() *stdHttp.Request {
				tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
					Subject:   uuid.New().String(),
					Issuer:    "librarium",
					IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
					ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(2 * time.Hour)),
				})
				token, err := tok.SignedString([]byte(os.Getenv("AUTH_SIGNING_KEY")))
				assert.Nil(t, err)
				req, err := stdHttp.NewRequest(stdHttp.MethodGet, ts.URL+"/librarian", stdHttp.NoBody)
				assert.Nil(t, err)
				req.Header.Add("Authorization", token)
				return req
			},
			assertResponse: func(rsp *stdHttp.Response) {
				assert.NotNil(t, rsp)
				assert.Equal(t, stdHttp.StatusOK, rsp.StatusCode)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			rsp, err := stdHttp.DefaultClient.Do(tc.request())
			assert.Nil(t, err)
			tc.assertResponse(rsp)
		})
	}
}

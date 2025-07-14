package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	stdHttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"librarium/internal/auth"
	"librarium/internal/catalog"
	"librarium/internal/http"
	"librarium/internal/mocks"
	"librarium/internal/onboarding"
	"librarium/internal/user"
)

type testDependencies struct {
	server            *http.Server
	httpServer        *httptest.Server
	serverCl          *stdHttp.Client
	userRepository    *mocks.MockUserRepository
	catalogReposiotry *mocks.MockCatalogRepository
	rentalRepository  *mocks.MockRentalRepository
}

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

func TestRoutingAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ts := buildTestServer(t)
	defer ts.httpServer.Close()

	testCases := map[string]struct {
		path           string
		method         string
		body           func() io.Reader
		mocks          func()
		assertResponse func(rsp *stdHttp.Response)
	}{
		"it should route to signup endpoint": {
			path:   "/signup",
			method: stdHttp.MethodPost,
			body: func() io.Reader {
				onboardingReq := onboarding.LibrarianRequest{Name: "John Doe", Email: "john.doe@test.com", Password: "strong-pass"}
				buf, err := json.Marshal(&onboardingReq)
				assert.Nil(t, err)
				return bytes.NewReader(buf)
			},
			mocks: func() {
				ts.userRepository.EXPECT().GetLibrarianByEmail("john.doe@test.com").Return(&user.Librarian{Email: "john.doe@test.com"}, nil)
			},
			assertResponse: func(rsp *stdHttp.Response) {
				assert.Equal(t, stdHttp.StatusConflict, rsp.StatusCode)
			},
		},
		"it should route to login endpoint": {
			path:   "/login",
			method: stdHttp.MethodPost,
			body: func() io.Reader {
				loginReq := auth.LoginRequest{Email: "john.doe@test.com", Password: "strong-pass"}
				buf, err := json.Marshal(&loginReq)
				assert.Nil(t, err)
				return bytes.NewReader(buf)
			},
			mocks: func() {
				ts.userRepository.EXPECT().GetLibrarianByEmail("john.doe@test.com").Return(nil, nil)
			},
			assertResponse: func(rsp *stdHttp.Response) {
				assert.Equal(t, stdHttp.StatusNotFound, rsp.StatusCode)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			req := buildServerWithoutAuthRequest(t, tc.method, ts.httpServer.URL, tc.path, tc.body())

			tc.mocks()
			rsp, err := ts.serverCl.Do(req)
			assert.Nil(t, err)
			tc.assertResponse(rsp)
		})
	}
}

func TestRoutingCatalogAsset(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ts := buildTestServer(t)
	defer ts.httpServer.Close()

	testCases := map[string]struct {
		path           string
		method         string
		body           func() io.Reader
		mocks          func()
		assertResponse func(rsp *stdHttp.Response)
	}{
		"it should route to catalog asset creation": {
			path:   "/catalog/assets",
			method: stdHttp.MethodPost,
			body: func() io.Reader {
				createAssetReq := catalog.CreateAssetRequest{
					Category: catalog.AssetCategoryBook,
					Asset: &catalog.Book{
						Title:       "The Lord Of The Rings The Two Towers",
						Author:      "J.R.R Tolkien",
						Publisher:   "George Allen & Unwin",
						ISBN:        "978-0261102385",
						PageCount:   352,
						PublishedAt: time.Date(1954, time.November, 11, 0, 0, 0, 0, time.UTC),
					},
				}
				buf, err := json.Marshal(&createAssetReq)
				assert.Nil(t, err)
				return bytes.NewReader(buf)
			},
			mocks: func() {
				ts.catalogReposiotry.EXPECT().CreateAsset(gomock.Any()).Return(nil)
			},
			assertResponse: func(rsp *stdHttp.Response) {
				assert.Equal(t, stdHttp.StatusOK, rsp.StatusCode)
			},
		},
		"it should route to catalog asset deletion": {
			path:   "/catalog/assets/" + uuid.NewString(),
			method: stdHttp.MethodDelete,
			body: func() io.Reader {
				return stdHttp.NoBody
			},
			mocks: func() {
				ts.catalogReposiotry.EXPECT().GetAsset(gomock.Any()).Return(nil, nil)
			},
			assertResponse: func(rsp *stdHttp.Response) {
				assert.Equal(t, stdHttp.StatusNotFound, rsp.StatusCode)
			},
		},
		"it should route to find catalog assets": {
			path:   "/catalog/assets",
			method: stdHttp.MethodGet,
			body: func() io.Reader {
				return stdHttp.NoBody
			},
			mocks: func() {
				ts.catalogReposiotry.EXPECT().FindAssets(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*catalog.Asset{}, nil)
			},
			assertResponse: func(rsp *stdHttp.Response) {
				assert.Equal(t, stdHttp.StatusOK, rsp.StatusCode)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			req := buildServerWithAuthRequest(t, tc.method, ts.httpServer.URL, tc.path, tc.body())

			tc.mocks()
			rsp, err := ts.serverCl.Do(req)
			assert.Nil(t, err)
			tc.assertResponse(rsp)
		})
	}
}

func buildTestServer(t *testing.T) testDependencies {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userRepository := mocks.NewMockUserRepository(ctrl)
	catalogRepository := mocks.NewMockCatalogRepository(ctrl)
	rentalRepository := mocks.NewMockRentalRepository(ctrl)

	address := ":8080"

	authController, err := http.NewAuthController(userRepository)
	assert.Nil(t, err)
	catalogController, err := http.NewCatalogController(catalogRepository)
	assert.Nil(t, err)
	customerController, err := http.NewCustomerController(userRepository)
	assert.Nil(t, err)
	rentalController, err := http.NewRentalController(rentalRepository, userRepository, catalogRepository)
	assert.Nil(t, err)

	srv, err := http.NewServer(address, authController, catalogController, customerController, rentalController)
	assert.Nil(t, err)

	serverTest := httptest.NewServer(srv.Handler)

	return testDependencies{
		server:            srv,
		serverCl:          serverTest.Client(),
		httpServer:        serverTest,
		userRepository:    userRepository,
		rentalRepository:  rentalRepository,
		catalogReposiotry: catalogRepository,
	}
}

func buildServerWithoutAuthRequest(t *testing.T, method, url, path string, body io.Reader) *stdHttp.Request {
	req, err := stdHttp.NewRequest(method, url+path, body)
	assert.Nil(t, err)
	if body != stdHttp.NoBody {
		req.Header.Set("Content-Type", "application/json")
	}

	return req
}

func buildServerWithAuthRequest(t *testing.T, method, url, path string, body io.Reader) *stdHttp.Request {
	t.Setenv("AUTH_SIGNING_KEY", "test_key")

	req, err := stdHttp.NewRequest(method, url+path, body)
	assert.Nil(t, err)
	if body != stdHttp.NoBody {
		req.Header.Set("Content-Type", "application/json")
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   uuid.NewString(),
		Issuer:    "librarium",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(2 * time.Hour)),
	})
	signedtok, err := tok.SignedString([]byte("test_key"))
	assert.Nil(t, err)

	req.Header.Set("Authorization", signedtok)

	return req
}

package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	stdHttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"librarium/internal/http"
	"librarium/internal/mocks"
	"librarium/internal/onboarding"
	"librarium/internal/user"
)

func TestNewCustomerController(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepository := mocks.NewMockUserRepository(ctrl)

	testCases := map[string]struct {
		userRepository   user.Repository
		expectedErr      error
		assertController func(controller *http.CustomerController)
	}{
		"it should return error when the user repository is missing": {
			expectedErr: errors.New("user repository is mandatory"),
			assertController: func(controller *http.CustomerController) {
				assert.Nil(t, controller)
			},
		},
		"it should return no error if all the dependencies are given": {
			userRepository: userRepository,
			assertController: func(controller *http.CustomerController) {
				assert.NotNil(t, controller)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			controller, err := http.NewCustomerController(tc.userRepository)
			assert.Equal(t, tc.expectedErr, err)
			tc.assertController(controller)
		})
	}
}

func TestCreateCustomer(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepository := mocks.NewMockUserRepository(ctrl)
	controller, err := http.NewCustomerController(userRepository)
	assert.NotNil(t, controller)
	assert.Nil(t, err)

	testCases := map[string]struct {
		mocks              func()
		request            func() *stdHttp.Request
		expectedStatusCode int
		assertBody         func(body io.Reader)
	}{
		"it should return bad request if JSON decoding fails": {
			mocks: func() {},
			request: func() *stdHttp.Request {
				return httptest.NewRequest(stdHttp.MethodPost, "/customers", bytes.NewReader([]byte("wrong input")))
			},
			expectedStatusCode: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				var errResp struct {
					Error string `json:"error"`
				}
				err := json.NewDecoder(body).Decode(&errResp)
				assert.Nil(t, err)
				assert.Equal(t, "error decoding customer request", errResp.Error)
			},
		},
		"it should return bad request if onboarding fails (missing fields)": {
			mocks: func() {},
			request: func() *stdHttp.Request {
				req := onboarding.CustomerRequest{
					Email: "invalid@example.com",
				}
				buf, err := json.Marshal(req)
				assert.Nil(t, err)
				return httptest.NewRequest(stdHttp.MethodPost, "/customers", bytes.NewReader(buf))
			},
			expectedStatusCode: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				var errResp struct {
					Error string `json:"error"`
				}
				err := json.NewDecoder(body).Decode(&errResp)
				assert.Nil(t, err)
				assert.Equal(t, "error onboarding customer", errResp.Error)
			},
		},
		"it should return internal server error if repository returns an error": {
			mocks: func() {
				userRepository.EXPECT().CreateCustomer(gomock.Any()).Return(errors.New("expected error"))
			},
			request: func() *stdHttp.Request {
				req := onboarding.CustomerRequest{
					Name:        "Alice",
					LastName:    "Smith",
					NationalID:  "12345678A",
					Email:       "alice@example.com",
					PhoneNumber: "+34600000000",
					Street:      "Fake Street 123",
					City:        "Barcelona",
					State:       "Catalonia",
					PostalCode:  "08001",
					Country:     "Spain",
				}
				buf, err := json.Marshal(req)
				assert.Nil(t, err)
				return httptest.NewRequest(stdHttp.MethodPost, "/customers", bytes.NewReader(buf))
			},
			expectedStatusCode: stdHttp.StatusInternalServerError,
			assertBody: func(body io.Reader) {
				var errResp struct {
					Error string `json:"error"`
				}
				err := json.NewDecoder(body).Decode(&errResp)
				assert.Nil(t, err)
				assert.Equal(t, "error creating customer", errResp.Error)
			},
		},
		"it should return created and the customer ID on success": {
			mocks: func() {
				customerMatcher := mocks.CustomMatcher{
					Match: func(x any) bool {
						c, ok := x.(*user.Customer)
						return assert.True(t, ok) &&
							assert.Equal(t, "alice@example.com", c.ContactDetails.Email) &&
							assert.Equal(t, "Alice", c.Name) &&
							assert.Equal(t, "Smith", c.LastName)
					},
				}
				userRepository.EXPECT().CreateCustomer(customerMatcher).Return(nil)
			},
			request: func() *stdHttp.Request {
				req := onboarding.CustomerRequest{
					Name:        "Alice",
					LastName:    "Smith",
					NationalID:  "12345678A",
					Email:       "alice@example.com",
					PhoneNumber: "+34600000000",
					Street:      "Fake Street 123",
					City:        "Barcelona",
					State:       "Barcelona",
					PostalCode:  "08001",
					Country:     "Spain",
				}
				buf, err := json.Marshal(req)
				assert.Nil(t, err)
				return httptest.NewRequest(stdHttp.MethodPost, "/customers", bytes.NewReader(buf))
			},
			expectedStatusCode: stdHttp.StatusCreated,
			assertBody: func(body io.Reader) {
				var resp struct {
					ID string `json:"id"`
				}
				err := json.NewDecoder(body).Decode(&resp)
				assert.Nil(t, err)
				assert.NotZero(t, resp.ID)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			tc.mocks()
			rec := httptest.NewRecorder()
			controller.Create(rec, tc.request())

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatusCode, res.StatusCode)
			tc.assertBody(res.Body)
		})
	}
}

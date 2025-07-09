package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	stdHttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
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

func TestFindCustomers(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepository := mocks.NewMockUserRepository(ctrl)
	controller, err := http.NewCustomerController(userRepository)
	assert.Nil(t, err)
	assert.NotNil(t, controller)

	testCases := map[string]struct {
		mocks              func()
		request            func() *stdHttp.Request
		expectedStatusCode int
		assertBody         func(body io.Reader)
	}{
		"should return an error if pagination params are invalid": {
			mocks: func() {},
			request: func() *stdHttp.Request {
				req := httptest.NewRequest(stdHttp.MethodGet, "/customers?limit=invalid", nil)
				return req
			},
			expectedStatusCode: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				var errMsg struct {
					Error string `json:"error"`
				}
				err := json.NewDecoder(body).Decode(&errMsg)
				assert.Nil(t, err)
				assert.Contains(t, errMsg.Error, "error getting limit")
			},
		},
		"should return an error if repository returns an error": {
			mocks: func() {
				userRepository.EXPECT().FindCustomers(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("expected error"))
			},
			request: func() *stdHttp.Request {
				req := httptest.NewRequest(stdHttp.MethodGet, "/customers?limit=10&offset=0", nil)
				return req
			},
			expectedStatusCode: stdHttp.StatusInternalServerError,
			assertBody: func(body io.Reader) {
				var errMsg struct {
					Error string `json:"error"`
				}
				err := json.NewDecoder(body).Decode(&errMsg)
				assert.Nil(t, err)
				assert.Equal(t, "error finding customers", errMsg.Error)
			},
		},
		"should return 200 and list of customers": {
			mocks: func() {
				customers := []*user.Customer{
					{
						ID:         uuid.New(),
						Name:       "Alice",
						LastName:   "Smith",
						NationalID: "12345678A",
						ContactDetails: &user.ContactDetails{
							Email: "alice@example.com",
							Address: &user.Address{
								City: "Madrid",
							},
						},
					},
				}
				userRepository.EXPECT().FindCustomers(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(customers, nil)
			},
			request: func() *stdHttp.Request {
				req := httptest.NewRequest(stdHttp.MethodGet, "/customers?limit=10&offset=0", nil)
				return req
			},
			expectedStatusCode: stdHttp.StatusOK,
			assertBody: func(body io.Reader) {
				var result []*user.Customer
				err := json.NewDecoder(body).Decode(&result)
				assert.Nil(t, err)
				assert.Len(t, result, 1)
				assert.Equal(t, "Alice", result[0].Name)
				assert.Equal(t, "Smith", result[0].LastName)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			tc.mocks()
			rec := httptest.NewRecorder()
			controller.Find(rec, tc.request())

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatusCode, res.StatusCode)
			tc.assertBody(res.Body)
		})
	}
}

func TestSuspendCustomer(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepository := mocks.NewMockUserRepository(ctrl)
	controller, err := http.NewCustomerController(userRepository)
	assert.Nil(t, err)
	assert.NotNil(t, controller)
	expectedCustomerID := uuid.New()

	testCases := map[string]struct {
		mocks              func()
		request            func() *stdHttp.Request
		expectedStatusCode int
		assertBody         func(body io.Reader)
	}{
		"it should return an error if path is invalid": {
			mocks: func() {},
			request: func() *stdHttp.Request {
				return httptest.NewRequest(stdHttp.MethodPost, "/customers", nil)
			},
			expectedStatusCode: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				var res struct {
					Error string `json:"error"`
				}
				err := json.NewDecoder(body).Decode(&res)
				assert.Nil(t, err)
				assert.Equal(t, "invalid expected path", res.Error)
			},
		},
		"it should return an error if the id is not a valid UUID": {
			mocks: func() {},
			request: func() *stdHttp.Request {
				return httptest.NewRequest(stdHttp.MethodPost, "/customers/not-a-uuid/suspend", nil)
			},
			expectedStatusCode: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				var res struct {
					Error string `json:"error"`
				}
				err := json.NewDecoder(body).Decode(&res)
				assert.Nil(t, err)
				assert.Equal(t, "invalid customer ID format, expected UUID", res.Error)
			},
		},
		"it should return an error if repository fails to get customer": {
			mocks: func() {
				userRepository.EXPECT().GetCustomer(gomock.Any()).Return(nil, errors.New("repo error"))
			},
			request: func() *stdHttp.Request {
				return httptest.NewRequest(stdHttp.MethodPost, fmt.Sprintf("/customers/%s/suspend", expectedCustomerID), nil)
			},
			expectedStatusCode: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				var res struct {
					Error string `json:"error"`
				}
				err := json.NewDecoder(body).Decode(&res)
				assert.Nil(t, err)
				assert.Equal(t, "error getting customer", res.Error)
			},
		},
		"it should return an error if suspend fails": {
			mocks: func() {
				userRepository.EXPECT().GetCustomer(expectedCustomerID).Return(&user.Customer{
					ID:     expectedCustomerID,
					Status: user.CustomerStatusSuspended,
				}, nil)
			},
			request: func() *stdHttp.Request {
				req := httptest.NewRequest(stdHttp.MethodPost, "/customers/"+expectedCustomerID.String()+"/suspend", nil)
				return req
			},
			expectedStatusCode: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				var res struct {
					Error string `json:"error"`
				}
				err := json.NewDecoder(body).Decode(&res)
				assert.Nil(t, err)
				assert.Equal(t, "error suspending customer", res.Error)
			},
		},
		"it should return an error if update fails": {
			mocks: func() {
				userRepository.EXPECT().GetCustomer(expectedCustomerID).Return(&user.Customer{
					ID:     expectedCustomerID,
					Status: user.CustomerStatusActive,
				}, nil)
				userRepository.EXPECT().UpdateCustomer(&user.Customer{
					ID:     expectedCustomerID,
					Status: user.CustomerStatusSuspended,
				}).Return(errors.New("expected error"))
			},
			request: func() *stdHttp.Request {
				req := httptest.NewRequest(stdHttp.MethodPost, "/customers/"+expectedCustomerID.String()+"/suspend", nil)
				return req
			},
			expectedStatusCode: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				var res struct {
					Error string `json:"error"`
				}
				err := json.NewDecoder(body).Decode(&res)
				assert.Nil(t, err)
				assert.Equal(t, "error updating customer", res.Error)
			},
		},
		"it should return no error if suspend and update succeed": {
			mocks: func() {
				userRepository.EXPECT().GetCustomer(expectedCustomerID).Return(&user.Customer{
					ID:     expectedCustomerID,
					Status: user.CustomerStatusActive,
				}, nil)
				userRepository.EXPECT().UpdateCustomer(&user.Customer{
					ID:     expectedCustomerID,
					Status: user.CustomerStatusSuspended,
				}).Return(nil)
			},
			request: func() *stdHttp.Request {
				req := httptest.NewRequest(stdHttp.MethodPost, "/customers/"+expectedCustomerID.String()+"/suspend", nil)
				return req
			},
			expectedStatusCode: stdHttp.StatusNoContent,
			assertBody: func(body io.Reader) {
				data, err := io.ReadAll(body)
				assert.Nil(t, err)
				assert.Len(t, data, 0)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			tc.mocks()
			rec := httptest.NewRecorder()
			controller.Suspend(rec, tc.request())

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatusCode, res.StatusCode)
			tc.assertBody(res.Body)
		})
	}
}

func TestUnSuspendCustomer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepository := mocks.NewMockUserRepository(ctrl)
	controller, err := http.NewCustomerController(userRepository)
	assert.Nil(t, err)
	assert.NotNil(t, controller)
	expectedCustomerID := uuid.New()

	testCases := map[string]struct {
		mocks          func()
		request        func() *stdHttp.Request
		expectedStatus int
		assertBody     func(body io.Reader)
	}{
		"it should return an error if path is invalid": {
			mocks: func() {},
			request: func() *stdHttp.Request {
				return httptest.NewRequest(stdHttp.MethodPost, "/customers", nil)
			},
			expectedStatus: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				var res struct{ Error string }
				_ = json.NewDecoder(body).Decode(&res)
				assert.Equal(t, "invalid expected path", res.Error)
			},
		},
		"it should return an error if the id is not a valid UUID": {
			mocks: func() {},
			request: func() *stdHttp.Request {
				return httptest.NewRequest(stdHttp.MethodPost, "/customers/invalid-uuid/unsuspend", nil)
			},
			expectedStatus: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				var res struct{ Error string }
				_ = json.NewDecoder(body).Decode(&res)
				assert.Equal(t, "invalid customer ID format, expected UUID", res.Error)
			},
		},
		"it should return an error if getting customer repo call fails": {
			mocks: func() {
				userRepository.EXPECT().GetCustomer(expectedCustomerID).Return(nil, errors.New("repo error"))
			},
			request: func() *stdHttp.Request {
				return httptest.NewRequest(stdHttp.MethodPost, "/customers/"+expectedCustomerID.String()+"/unsuspend", nil)
			},
			expectedStatus: stdHttp.StatusInternalServerError,
			assertBody: func(body io.Reader) {
				var res struct{ Error string }
				_ = json.NewDecoder(body).Decode(&res)
				assert.Equal(t, "error getting customer", res.Error)
			},
		},
		"it should return an error if we try to unsuspend the customer": {
			mocks: func() {
				userRepository.EXPECT().GetCustomer(expectedCustomerID).Return(&user.Customer{
					ID:     expectedCustomerID,
					Status: user.CustomerStatusActive,
				}, nil)
			},
			request: func() *stdHttp.Request {
				return httptest.NewRequest(stdHttp.MethodPost, "/customers/"+expectedCustomerID.String()+"/unsuspend", nil)
			},
			expectedStatus: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				var res struct{ Error string }
				_ = json.NewDecoder(body).Decode(&res)
				assert.Equal(t, "customer should be suspended to be unsuspend", res.Error)
			},
		},
		"it should return an error if the update repo call fails": {
			mocks: func() {
				userRepository.EXPECT().GetCustomer(expectedCustomerID).Return(&user.Customer{
					ID:     expectedCustomerID,
					Status: user.CustomerStatusSuspended,
				}, nil)
				userRepository.EXPECT().UpdateCustomer(&user.Customer{
					ID:     expectedCustomerID,
					Status: user.CustomerStatusActive,
				}).Return(errors.New("expected error"))
			},
			request: func() *stdHttp.Request {
				return httptest.NewRequest(stdHttp.MethodPost, "/customers/"+expectedCustomerID.String()+"/unsuspend", nil)
			},
			expectedStatus: stdHttp.StatusInternalServerError,
			assertBody: func(body io.Reader) {
				var res struct{ Error string }
				_ = json.NewDecoder(body).Decode(&res)
				assert.Equal(t, "error updating customer", res.Error)
			},
		},
		"it should return no error and unsuspend the customer": {
			mocks: func() {
				userRepository.EXPECT().GetCustomer(expectedCustomerID).Return(&user.Customer{
					ID:     expectedCustomerID,
					Status: user.CustomerStatusSuspended,
				}, nil)
				userRepository.EXPECT().UpdateCustomer(&user.Customer{
					ID:     expectedCustomerID,
					Status: user.CustomerStatusActive,
				}).Return(nil)
			},
			request: func() *stdHttp.Request {
				return httptest.NewRequest(stdHttp.MethodPost, "/customers/"+expectedCustomerID.String()+"/unsuspend", nil)
			},
			expectedStatus: stdHttp.StatusNoContent,
			assertBody: func(body io.Reader) {
				data, _ := io.ReadAll(body)
				assert.Len(t, data, 0)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			tc.mocks()

			rec := httptest.NewRecorder()
			req := tc.request()
			controller.UnSuspend(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatus, res.StatusCode)
			tc.assertBody(res.Body)
		})
	}
}

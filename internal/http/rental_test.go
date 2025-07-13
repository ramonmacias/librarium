package http_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	stdHttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"librarium/internal/catalog"
	"librarium/internal/http"
	"librarium/internal/mocks"
	"librarium/internal/rental"
	"librarium/internal/user"
)

func TestNewRentalController(t *testing.T) {
	ctrl := gomock.NewController(t)
	rentalRepository := mocks.NewMockRentalRepository(ctrl)
	userRepository := mocks.NewMockUserRepository(ctrl)
	catalogRepository := mocks.NewMockCatalogRepository(ctrl)

	testCases := map[string]struct {
		rentalRepository  rental.Repository
		userRepository    user.Repository
		catalogRepository catalog.Repository
		expectedErr       error
		assertController  func(controller *http.RentalController)
	}{
		"it should return error when the rental repository is missing": {
			expectedErr: errors.New("rental repository is mandatory"),
			assertController: func(controller *http.RentalController) {
				assert.Nil(t, controller)
			},
		},
		"it should return error when the user repository is missing": {
			rentalRepository: rentalRepository,
			expectedErr:      errors.New("user repository is mandatory"),
			assertController: func(controller *http.RentalController) {
				assert.Nil(t, controller)
			},
		},
		"it should return error when the catalog repository is missing": {
			rentalRepository: rentalRepository,
			userRepository:   userRepository,
			expectedErr:      errors.New("catalog repository is mandatory"),
			assertController: func(controller *http.RentalController) {
				assert.Nil(t, controller)
			},
		},
		"it should return no error if all the dependencies are given": {
			userRepository:    userRepository,
			catalogRepository: catalogRepository,
			rentalRepository:  rentalRepository,
			assertController: func(controller *http.RentalController) {
				assert.NotNil(t, controller)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			controller, err := http.NewRentalController(tc.rentalRepository, tc.userRepository, tc.catalogRepository)
			assert.Equal(t, tc.expectedErr, err)
			tc.assertController(controller)
		})
	}
}

func TestFindRentals(t *testing.T) {
	ctrl := gomock.NewController(t)
	rentalRepository := mocks.NewMockRentalRepository(ctrl)
	userRepository := mocks.NewMockUserRepository(ctrl)
	catalogRepository := mocks.NewMockCatalogRepository(ctrl)
	controller, err := http.NewRentalController(rentalRepository, userRepository, catalogRepository)
	assert.Nil(t, err)
	assert.NotNil(t, controller)

	testCases := map[string]struct {
		mocks              func()
		request            func() *stdHttp.Request
		expectedStatusCode int
		assertBody         func(body io.Reader)
	}{
		"it should return bad request if limit param is invalid": {
			mocks: func() {},
			request: func() *stdHttp.Request {
				return httptest.NewRequest(stdHttp.MethodGet, "/rentals?limit=notanint", nil)
			},
			expectedStatusCode: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				var errResp struct {
					Error string `json:"error"`
				}
				err := json.NewDecoder(body).Decode(&errResp)
				assert.NoError(t, err)
				assert.Contains(t, errResp.Error, "error getting limit")
			},
		},
		"it should return bad request if offset param is invalid": {
			mocks: func() {},
			request: func() *stdHttp.Request {
				return httptest.NewRequest(stdHttp.MethodGet, "/rentals?offset=abc", nil)
			},
			expectedStatusCode: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				var errResp struct {
					Error string `json:"error"`
				}
				err := json.NewDecoder(body).Decode(&errResp)
				assert.NoError(t, err)
				assert.Contains(t, errResp.Error, "error getting offset")
			},
		},
		"it should return internal server error if repository fails": {
			mocks: func() {
				rentalRepository.EXPECT().
					FindRentals(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("expected error"))
			},
			request: func() *stdHttp.Request {
				return httptest.NewRequest(stdHttp.MethodGet, "/rentals?limit=10&offset=0", nil)
			},
			expectedStatusCode: stdHttp.StatusInternalServerError,
			assertBody: func(body io.Reader) {
				var errResp struct {
					Error string `json:"error"`
				}
				err := json.NewDecoder(body).Decode(&errResp)
				assert.NoError(t, err)
				assert.Equal(t, "error finding rentals", errResp.Error)
			},
		},
		"it should return no errors and the list of rentals": {
			mocks: func() {
				re := &rental.Rental{
					ID:         uuid.New(),
					CustomerID: uuid.New(),
					AssetID:    uuid.New(),
					RentedAt:   time.Now(),
					Status:     rental.RentalStatusActive,
				}
				rentalRepository.EXPECT().
					FindRentals(gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]*rental.Rental{re}, nil)
			},
			request: func() *stdHttp.Request {
				return httptest.NewRequest(stdHttp.MethodGet, "/rentals?limit=10&offset=0", nil)
			},
			expectedStatusCode: stdHttp.StatusOK,
			assertBody: func(body io.Reader) {
				var rentals []*rental.Rental
				err := json.NewDecoder(body).Decode(&rentals)
				assert.NoError(t, err)
				assert.Len(t, rentals, 1)
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

func TestCreateRental(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepository := mocks.NewMockUserRepository(ctrl)
	catalogRepository := mocks.NewMockCatalogRepository(ctrl)
	rentalRepository := mocks.NewMockRentalRepository(ctrl)
	controller, err := http.NewRentalController(rentalRepository, userRepository, catalogRepository)
	assert.Nil(t, err)
	assert.NotNil(t, controller)

	customerID := uuid.New()
	assetID := uuid.New()

	testCases := map[string]struct {
		body               string
		mocks              func()
		expectedStatusCode int
		assertBody         func(body io.Reader)
	}{
		"it should return an error while decoding the json": {
			body:               `invalid-json`,
			mocks:              func() {},
			expectedStatusCode: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				var resp struct {
					Error string `json:"error"`
				}
				err = json.NewDecoder(body).Decode(&resp)
				assert.Nil(t, err)
				assert.Contains(t, resp.Error, "error decoding rental request")
			},
		},
		"it should return an error if getting customer repo call fails": {
			body: fmt.Sprintf(`{"customer_id": "%s", "asset_id": "%s"}`, customerID, assetID),
			mocks: func() {
				userRepository.EXPECT().GetCustomer(customerID).Return(nil, errors.New("expected error"))
			},
			expectedStatusCode: stdHttp.StatusInternalServerError,
			assertBody: func(body io.Reader) {
				var resp struct {
					Error string `json:"error"`
				}
				err = json.NewDecoder(body).Decode(&resp)
				assert.Nil(t, err)
				assert.Equal(t, "error getting customer", resp.Error)
			},
		},
		"it should return an error if cannot find the customer": {
			body: fmt.Sprintf(`{"customer_id": "%s", "asset_id": "%s"}`, customerID, assetID),
			mocks: func() {
				userRepository.EXPECT().GetCustomer(customerID).Return(nil, nil)
			},
			expectedStatusCode: stdHttp.StatusNotFound,
			assertBody: func(body io.Reader) {
				var resp struct {
					Error string `json:"error"`
				}
				err = json.NewDecoder(body).Decode(&resp)
				assert.Nil(t, err)
				assert.Equal(t, "customer not found", resp.Error)
			},
		},
		"it should return an error if the get asset repo call fails": {
			body: fmt.Sprintf(`{"customer_id": "%s", "asset_id": "%s"}`, customerID, assetID),
			mocks: func() {
				userRepository.EXPECT().GetCustomer(customerID).Return(&user.Customer{ID: customerID}, nil)
				catalogRepository.EXPECT().GetAsset(assetID).Return(nil, errors.New("expected error"))
			},
			expectedStatusCode: stdHttp.StatusInternalServerError,
			assertBody: func(body io.Reader) {
				var resp struct {
					Error string `json:"error"`
				}
				err = json.NewDecoder(body).Decode(&resp)
				assert.Nil(t, err)
				assert.Equal(t, "error getting asset", resp.Error)
			},
		},
		"it should return an error if cannot find the asset": {
			body: fmt.Sprintf(`{"customer_id": "%s", "asset_id": "%s"}`, customerID, assetID),
			mocks: func() {
				userRepository.EXPECT().GetCustomer(customerID).Return(&user.Customer{ID: customerID}, nil)
				catalogRepository.EXPECT().GetAsset(assetID).Return(nil, nil)
			},
			expectedStatusCode: stdHttp.StatusNotFound,
			assertBody: func(body io.Reader) {
				var resp struct {
					Error string `json:"error"`
				}
				err = json.NewDecoder(body).Decode(&resp)
				assert.Nil(t, err)
				assert.Equal(t, "asset not found", resp.Error)
			},
		},
		"it should return an error if active rental repo call fails": {
			body: fmt.Sprintf(`{"customer_id": "%s", "asset_id": "%s"}`, customerID, assetID),
			mocks: func() {
				userRepository.EXPECT().GetCustomer(customerID).Return(&user.Customer{ID: customerID}, nil)
				catalogRepository.EXPECT().GetAsset(assetID).Return(&catalog.Asset{ID: assetID}, nil)
				rentalRepository.EXPECT().GetActiveRental(customerID, assetID).Return(nil, errors.New("expected error"))
			},
			expectedStatusCode: stdHttp.StatusInternalServerError,
			assertBody: func(body io.Reader) {
				var resp struct {
					Error string `json:"error"`
				}
				err = json.NewDecoder(body).Decode(&resp)
				assert.Nil(t, err)
				assert.Equal(t, "error getting active rental", resp.Error)
			},
		},
		"it should return an error if find customer rentals repo call fails": {
			body: fmt.Sprintf(`{"customer_id": "%s", "asset_id": "%s"}`, customerID, assetID),
			mocks: func() {
				userRepository.EXPECT().GetCustomer(customerID).Return(&user.Customer{ID: customerID}, nil)
				catalogRepository.EXPECT().GetAsset(assetID).Return(&catalog.Asset{ID: assetID}, nil)
				rentalRepository.EXPECT().GetActiveRental(customerID, assetID).Return(nil, nil)
				rentalRepository.EXPECT().FindRentals(gomock.Any(), gomock.Nil(), gomock.Nil()).Return(nil, errors.New("expected error"))
			},
			expectedStatusCode: stdHttp.StatusInternalServerError,
			assertBody: func(body io.Reader) {
				var resp struct {
					Error string `json:"error"`
				}
				err = json.NewDecoder(body).Decode(&resp)
				assert.Nil(t, err)
				assert.Equal(t, "error getting customer rentals", resp.Error)
			},
		},
		"it should return an error if the rental domain function fails": {
			body: fmt.Sprintf(`{"customer_id": "%s", "asset_id": "%s"}`, customerID, assetID),
			mocks: func() {
				userRepository.EXPECT().GetCustomer(customerID).Return(&user.Customer{ID: customerID}, nil)
				catalogRepository.EXPECT().GetAsset(assetID).Return(&catalog.Asset{ID: assetID}, nil)
				rentalRepository.EXPECT().GetActiveRental(customerID, assetID).Return(&rental.Rental{
					AssetID:    assetID,
					CustomerID: customerID,
				}, nil)
				rentalRepository.EXPECT().FindRentals(gomock.Any(), gomock.Nil(), gomock.Nil()).Return(nil, nil)
			},
			expectedStatusCode: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				var resp struct {
					Error string `json:"error"`
				}
				err = json.NewDecoder(body).Decode(&resp)
				assert.Nil(t, err)
				assert.Equal(t, "catalog asset already rented", resp.Error)
			},
		},
		"it should return an error if creating the rental repo call fails": {
			body: fmt.Sprintf(`{"customer_id": "%s", "asset_id": "%s"}`, customerID, assetID),
			mocks: func() {
				userRepository.EXPECT().GetCustomer(customerID).Return(&user.Customer{ID: customerID}, nil)
				catalogRepository.EXPECT().GetAsset(assetID).Return(&catalog.Asset{ID: assetID}, nil)
				rentalRepository.EXPECT().GetActiveRental(customerID, assetID).Return(nil, nil)
				rentalRepository.EXPECT().FindRentals(gomock.Any(), gomock.Nil(), gomock.Nil()).Return([]*rental.Rental{}, nil)
				rentalRepository.EXPECT().CreateRental(gomock.Any()).Return(errors.New("expected error"))
			},
			expectedStatusCode: stdHttp.StatusInternalServerError,
			assertBody: func(body io.Reader) {
				var resp struct {
					Error string `json:"error"`
				}
				err = json.NewDecoder(body).Decode(&resp)
				assert.Nil(t, err)
				assert.Equal(t, "error creating rental", resp.Error)
			},
		},
		"it should return no error and the rental created": {
			body: fmt.Sprintf(`{"customer_id": "%s", "asset_id": "%s"}`, customerID, assetID),
			mocks: func() {
				userRepository.EXPECT().GetCustomer(customerID).Return(&user.Customer{ID: customerID}, nil)
				catalogRepository.EXPECT().GetAsset(assetID).Return(&catalog.Asset{ID: assetID}, nil)
				rentalRepository.EXPECT().GetActiveRental(customerID, assetID).Return(nil, nil)
				rentalRepository.EXPECT().FindRentals(gomock.Any(), gomock.Nil(), gomock.Nil()).Return([]*rental.Rental{}, nil)
				rentalRepository.EXPECT().CreateRental(gomock.Any()).Return(nil)
			},
			expectedStatusCode: stdHttp.StatusOK,
			assertBody: func(body io.Reader) {
				var resp struct {
					ID string `json:"id"`
				}
				err = json.NewDecoder(body).Decode(&resp)
				assert.Nil(t, err)
				assert.NotZero(t, resp.ID)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			tc.mocks()
			req := httptest.NewRequest(stdHttp.MethodPost, "/rentals", strings.NewReader(tc.body))
			rec := httptest.NewRecorder()
			controller.Create(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatusCode, res.StatusCode)
			tc.assertBody(res.Body)
		})
	}
}

func TestReturnRental(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepository := mocks.NewMockUserRepository(ctrl)
	catalogRepository := mocks.NewMockCatalogRepository(ctrl)
	rentalRepository := mocks.NewMockRentalRepository(ctrl)
	controller, err := http.NewRentalController(rentalRepository, userRepository, catalogRepository)
	assert.Nil(t, err)
	assert.NotNil(t, controller)

	rentalID := uuid.New()

	testCases := map[string]struct {
		path               string
		mocks              func()
		expectedStatusCode int
		assertBody         func(body io.Reader)
	}{
		"it should return an error on invalid path structure": {
			path:               "/rentals/return",
			mocks:              func() {},
			expectedStatusCode: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				var resp struct{ Error string }
				err = json.NewDecoder(body).Decode(&resp)
				assert.Nil(t, err)
				assert.Equal(t, "invalid expected path", resp.Error)
			},
		},
		"it should return an error on invalid UUID": {
			path:               "/rentals/not-a-uuid/return",
			mocks:              func() {},
			expectedStatusCode: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				var resp struct{ Error string }
				err = json.NewDecoder(body).Decode(&resp)
				assert.Nil(t, err)
				assert.Equal(t, "invalid rental ID format, expected UUID", resp.Error)
			},
		},
		"it should return an error if getting rental repo call fails": {
			path: fmt.Sprintf("/rentals/%s/return", rentalID.String()),
			mocks: func() {
				rentalRepository.EXPECT().
					GetRental(rentalID).
					Return(nil, errors.New("expected error"))
			},
			expectedStatusCode: stdHttp.StatusInternalServerError,
			assertBody: func(body io.Reader) {
				var resp struct{ Error string }
				err = json.NewDecoder(body).Decode(&resp)
				assert.Nil(t, err)
				assert.Equal(t, "error getting rental", resp.Error)
			},
		},
		"it should return a not found error if cannot find the rental": {
			path: fmt.Sprintf("/rentals/%s/return", rentalID.String()),
			mocks: func() {
				rentalRepository.EXPECT().
					GetRental(rentalID).
					Return(nil, nil)
			},
			expectedStatusCode: stdHttp.StatusNotFound,
			assertBody: func(body io.Reader) {
				var resp struct{ Error string }
				err = json.NewDecoder(body).Decode(&resp)
				assert.Nil(t, err)
				assert.Equal(t, "rental not found", resp.Error)
			},
		},
		"it should return an error if we cannot return the rental": {
			path: fmt.Sprintf("/rentals/%s/return", rentalID.String()),
			mocks: func() {
				rentalRepository.EXPECT().
					GetRental(rentalID).
					Return(&rental.Rental{ID: rentalID, Status: rental.RentalStatusReturned}, nil)
			},
			expectedStatusCode: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				var resp struct{ Error string }
				err = json.NewDecoder(body).Decode(&resp)
				assert.Nil(t, err)
				assert.Equal(t, "the rental is already returned", resp.Error)
			},
		},
		"it should return an error if update rental repo call fails": {
			path: fmt.Sprintf("/rentals/%s/return", rentalID.String()),
			mocks: func() {
				rentalRepository.EXPECT().
					GetRental(rentalID).
					Return(&rental.Rental{ID: rentalID, Status: rental.RentalStatusActive}, nil)
				rentalMatcher := mocks.CustomMatcher{
					Match: func(x any) bool {
						re, ok := x.(*rental.Rental)
						assert.True(t, ok)

						return assert.Equal(t, rentalID, re.ID) && assert.Equal(t, rental.RentalStatusReturned, re.Status) &&
							assert.WithinDuration(t, time.Now().UTC(), *re.ReturnedAt, 2*time.Minute)
					},
				}
				rentalRepository.EXPECT().
					UpdateRental(rentalMatcher).
					Return(errors.New("expected error"))
			},
			expectedStatusCode: stdHttp.StatusInternalServerError,
			assertBody: func(body io.Reader) {
				var resp struct{ Error string }
				err = json.NewDecoder(body).Decode(&resp)
				assert.Nil(t, err)
				assert.Equal(t, "error updating rental", resp.Error)
			},
		},
		"it should return no error if success": {
			path: fmt.Sprintf("/rentals/%s/return", rentalID.String()),
			mocks: func() {
				rentalRepository.EXPECT().
					GetRental(rentalID).
					Return(&rental.Rental{ID: rentalID, Status: rental.RentalStatusActive}, nil)
				rentalMatcher := mocks.CustomMatcher{
					Match: func(x any) bool {
						re, ok := x.(*rental.Rental)
						assert.True(t, ok)

						return assert.Equal(t, rentalID, re.ID) && assert.Equal(t, rental.RentalStatusReturned, re.Status) &&
							assert.WithinDuration(t, time.Now().UTC(), *re.ReturnedAt, 2*time.Minute)
					},
				}
				rentalRepository.EXPECT().
					UpdateRental(rentalMatcher).
					Return(nil)
			},
			expectedStatusCode: stdHttp.StatusNoContent,
			assertBody: func(body io.Reader) {
				data, err := io.ReadAll(body)
				assert.Nil(t, err)
				assert.Empty(t, data)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			tc.mocks()

			req := httptest.NewRequest(stdHttp.MethodPost, tc.path, nil)
			rec := httptest.NewRecorder()

			controller.Return(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatusCode, res.StatusCode)
			tc.assertBody(res.Body)
		})
	}
}

func TestExtendRental(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepository := mocks.NewMockUserRepository(ctrl)
	catalogRepository := mocks.NewMockCatalogRepository(ctrl)
	rentalRepository := mocks.NewMockRentalRepository(ctrl)
	controller, err := http.NewRentalController(rentalRepository, userRepository, catalogRepository)
	assert.Nil(t, err)
	assert.NotNil(t, controller)

	rentalID := uuid.New()

	testCases := map[string]struct {
		path               string
		mocks              func()
		expectedStatusCode int
		assertBody         func(body io.Reader)
	}{
		"it should return an error for invalid path structure": {
			path:               "/rentals/extend",
			mocks:              func() {},
			expectedStatusCode: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				var resp struct{ Error string }
				err = json.NewDecoder(body).Decode(&resp)
				assert.Nil(t, err)
				assert.Equal(t, "invalid expected path", resp.Error)
			},
		},
		"it should return an error for invalid UUID format": {
			path:               "/rentals/not-a-uuid/extend",
			mocks:              func() {},
			expectedStatusCode: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				var resp struct{ Error string }
				err = json.NewDecoder(body).Decode(&resp)
				assert.Nil(t, err)
				assert.Equal(t, "invalid rental ID format, expected UUID", resp.Error)
			},
		},
		"it should return and error if getting rental repo call fails": {
			path: fmt.Sprintf("/rentals/%s/extend", rentalID),
			mocks: func() {
				rentalRepository.EXPECT().
					GetRental(rentalID).
					Return(nil, errors.New("expected error"))
			},
			expectedStatusCode: stdHttp.StatusInternalServerError,
			assertBody: func(body io.Reader) {
				var resp struct{ Error string }
				err = json.NewDecoder(body).Decode(&resp)
				assert.Nil(t, err)
				assert.Equal(t, "error getting rental", resp.Error)
			},
		},
		"it should return an error if rental cannot be find": {
			path: fmt.Sprintf("/rentals/%s/extend", rentalID),
			mocks: func() {
				rentalRepository.EXPECT().
					GetRental(rentalID).
					Return(nil, nil)
			},
			expectedStatusCode: stdHttp.StatusNotFound,
			assertBody: func(body io.Reader) {
				var resp struct{ Error string }
				err = json.NewDecoder(body).Decode(&resp)
				assert.Nil(t, err)
				assert.Equal(t, "rental not found", resp.Error)
			},
		},
		"it should return an error if we cannot extend the rental": {
			path: fmt.Sprintf("/rentals/%s/extend", rentalID),
			mocks: func() {
				rentalRepository.EXPECT().
					GetRental(rentalID).
					Return(&rental.Rental{
						ID:     rentalID,
						Status: rental.RentalStatusReturned,
					}, nil)
			},
			expectedStatusCode: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				var resp struct{ Error string }
				err = json.NewDecoder(body).Decode(&resp)
				assert.Nil(t, err)
				assert.Equal(t, "the rental is already returned", resp.Error)
			},
		},
		"it should return an error if updating rental repo call fails": {
			path: fmt.Sprintf("/rentals/%s/extend", rentalID),
			mocks: func() {
				rentalRepository.EXPECT().
					GetRental(rentalID).
					Return(&rental.Rental{
						ID:       rentalID,
						Status:   rental.RentalStatusActive,
						DueAt:    time.Now().AddDate(0, 1, 0).UTC(),
						RentedAt: time.Now().UTC(),
					}, nil)
				rentalMatcher := mocks.CustomMatcher{
					Match: func(x any) bool {
						re, ok := x.(*rental.Rental)
						assert.True(t, ok)

						return assert.Equal(t, rentalID, re.ID) && assert.Equal(t, rental.RentalStatusActive, re.Status) &&
							assert.WithinDuration(t, time.Now().UTC(), re.RentedAt, 2*time.Minute) &&
							assert.WithinDuration(t, time.Now().AddDate(0, 2, 0).UTC(), re.DueAt, 2*time.Minute)
					},
				}
				rentalRepository.EXPECT().
					UpdateRental(rentalMatcher).
					Return(errors.New("expected error"))
			},
			expectedStatusCode: stdHttp.StatusInternalServerError,
			assertBody: func(body io.Reader) {
				var resp struct{ Error string }
				err = json.NewDecoder(body).Decode(&resp)
				assert.Nil(t, err)
				assert.Equal(t, "error updating rental", resp.Error)
			},
		},
		"it should return no error on success": {
			path: fmt.Sprintf("/rentals/%s/extend", rentalID),
			mocks: func() {
				rentalRepository.EXPECT().
					GetRental(rentalID).
					Return(&rental.Rental{
						ID:       rentalID,
						Status:   rental.RentalStatusActive,
						DueAt:    time.Now().AddDate(0, 1, 0).UTC(),
						RentedAt: time.Now().UTC(),
					}, nil)
				rentalMatcher := mocks.CustomMatcher{
					Match: func(x any) bool {
						re, ok := x.(*rental.Rental)
						assert.True(t, ok)

						return assert.Equal(t, rentalID, re.ID) && assert.Equal(t, rental.RentalStatusActive, re.Status) &&
							assert.WithinDuration(t, time.Now().UTC(), re.RentedAt, 2*time.Minute) &&
							assert.WithinDuration(t, time.Now().AddDate(0, 2, 0).UTC(), re.DueAt, 2*time.Minute)
					},
				}
				rentalRepository.EXPECT().
					UpdateRental(rentalMatcher).
					Return(nil)
			},
			expectedStatusCode: stdHttp.StatusNoContent,
			assertBody: func(body io.Reader) {
				data, err := io.ReadAll(body)
				assert.NoError(t, err)
				assert.Empty(t, data)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			tc.mocks()

			req := httptest.NewRequest(stdHttp.MethodPost, tc.path, nil)
			rec := httptest.NewRecorder()

			controller.Extend(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatusCode, res.StatusCode)
			tc.assertBody(res.Body)
		})
	}
}

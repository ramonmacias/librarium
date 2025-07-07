package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	stdHttp "net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"librarium/internal/catalog"
	"librarium/internal/http"
	"librarium/internal/mocks"
)

func TestNewCatalogController(t *testing.T) {
	ctrl := gomock.NewController(t)
	catalogRepository := mocks.NewMockCatalogRepository(ctrl)

	testCases := map[string]struct {
		catalogRepository catalog.Repository
		expectedErr       error
		assertController  func(controller *http.CatalogController)
	}{
		"it should return error when the catalog repository is missing": {
			expectedErr: errors.New("catalog repository is mandatory"),
			assertController: func(controller *http.CatalogController) {
				assert.Nil(t, controller)
			},
		},
		"it should return no error if all the dependencies are given": {
			catalogRepository: catalogRepository,
			assertController: func(controller *http.CatalogController) {
				assert.NotNil(t, controller)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			controller, err := http.NewCatalogController(tc.catalogRepository)
			assert.Equal(t, tc.expectedErr, err)
			tc.assertController(controller)
		})
	}
}

func TestCreateCatalogItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	catalogRepository := mocks.NewMockCatalogRepository(ctrl)
	controller, err := http.NewCatalogController(catalogRepository)
	assert.NotNil(t, controller)
	assert.Nil(t, err)

	testCases := map[string]struct {
		mocks              func()
		request            func() *stdHttp.Request
		expectedStatusCode int
		assertBody         func(body io.Reader)
	}{
		"it should return an error if the json decoding fails": {
			mocks: func() {},
			request: func() *stdHttp.Request {
				return httptest.NewRequest(stdHttp.MethodPost, "/catalog/assets", bytes.NewReader([]byte("wrong input")))
			},
			expectedStatusCode: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				errorMsg := struct {
					Error string `json:"error"`
				}{}
				err := json.NewDecoder(body).Decode(&errorMsg)
				assert.Nil(t, err)
				assert.Equal(t, "error decoding asset catalog create request", errorMsg.Error)
			},
		},
		"it should return an error if the repository create asset call fails": {
			mocks: func() {
				catalogRepository.EXPECT().CreateAsset(gomock.Any()).Return(errors.New("expected error"))
			},
			request: func() *stdHttp.Request {
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
				return httptest.NewRequest(stdHttp.MethodPost, "/catalog/assets", bytes.NewReader(buf))
			},
			expectedStatusCode: stdHttp.StatusInternalServerError,
			assertBody: func(body io.Reader) {
				errorMsg := struct {
					Error string `json:"error"`
				}{}
				err := json.NewDecoder(body).Decode(&errorMsg)
				assert.Nil(t, err)
				assert.Equal(t, "error creating asset catalog", errorMsg.Error)
			},
		},
		"it should return no error and the catalgo asset created and the ID returned": {
			mocks: func() {
				assetCreateMatcher := mocks.CustomMatcher{
					Match: func(x any) bool {
						asset, ok := x.(*catalog.Asset)
						assert.True(t, ok)
						book, ok := asset.Info.(*catalog.Book)
						assert.True(t, ok)

						return assert.NotNil(t, asset) && assert.NotZero(t, asset.ID) && assert.Equal(t, catalog.AssetCategoryBook, asset.Category) &&
							assert.NotNil(t, book) && assert.Equal(t, "The Lord Of The Rings The Two Towers", book.Title) && assert.Equal(t, "J.R.R Tolkien", book.Author) &&
							assert.Equal(t, "George Allen & Unwin", book.Publisher) && assert.Equal(t, "978-0261102385", book.ISBN) && assert.Equal(t, 352, book.PageCount) &&
							assert.Equal(t, time.Date(2020, time.May, 20, 12, 30, 30, 0, time.UTC), book.PublishedAt)
					},
				}
				catalogRepository.EXPECT().CreateAsset(assetCreateMatcher).Return(nil)
			},
			request: func() *stdHttp.Request {
				buf, err := os.ReadFile("testdata/book.json")
				assert.Nil(t, err)
				return httptest.NewRequest(stdHttp.MethodPost, "/catalog/assets", bytes.NewReader(buf))
			},
			expectedStatusCode: stdHttp.StatusOK,
			assertBody: func(body io.Reader) {
				createAssetRsp := struct {
					ID string `json:"id"`
				}{}
				err := json.NewDecoder(body).Decode(&createAssetRsp)
				assert.Nil(t, err)
				assert.NotZero(t, createAssetRsp.ID)
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

func TestDeleteCatalogItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	catalogRepository := mocks.NewMockCatalogRepository(ctrl)
	controller, err := http.NewCatalogController(catalogRepository)
	assert.NotNil(t, controller)
	assert.Nil(t, err)
	expectedItemID := uuid.New()

	testCases := map[string]struct {
		mocks              func()
		request            func() *stdHttp.Request
		expectedStatusCode int
		assertBody         func(body io.Reader)
	}{
		"it should return bad request if the URL path is invalid": {
			mocks: func() {},
			request: func() *stdHttp.Request {
				return httptest.NewRequest(stdHttp.MethodDelete, "/catalog/invalid", nil)
			},
			expectedStatusCode: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				errorMsg := struct {
					Error string `json:"error"`
				}{}
				err := json.NewDecoder(body).Decode(&errorMsg)
				assert.Nil(t, err)
				assert.Equal(t, "error getting asset ID from the url", errorMsg.Error)
			},
		},
		"it should return bad request if the asset ID format is invalid": {
			mocks: func() {},
			request: func() *stdHttp.Request {
				return httptest.NewRequest(stdHttp.MethodDelete, "/catalog/assets/not-a-uuid", nil)
			},
			expectedStatusCode: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				errorMsg := struct {
					Error string `json:"error"`
				}{}
				err := json.NewDecoder(body).Decode(&errorMsg)
				assert.Nil(t, err)
				assert.Equal(t, "invalid asset ID format, expected UUID", errorMsg.Error)
			},
		},
		"it should return not found if asset does not exist": {
			mocks: func() {
				catalogRepository.EXPECT().GetAsset(expectedItemID).Return(nil, nil)
			},
			request: func() *stdHttp.Request {
				return httptest.NewRequest(stdHttp.MethodDelete, "/catalog/assets/"+expectedItemID.String(), nil)
			},
			expectedStatusCode: stdHttp.StatusNotFound,
			assertBody: func(body io.Reader) {
				errorMsg := struct {
					Error string `json:"error"`
				}{}
				err := json.NewDecoder(body).Decode(&errorMsg)
				assert.Nil(t, err)
				assert.Equal(t, "asset not found", errorMsg.Error)
			},
		},
		"it should return internal server error if get asset repo call fails": {
			mocks: func() {
				catalogRepository.EXPECT().GetAsset(expectedItemID).Return(nil, errors.New("expected error"))
			},
			request: func() *stdHttp.Request {
				return httptest.NewRequest(stdHttp.MethodDelete, "/catalog/assets/"+expectedItemID.String(), nil)
			},
			expectedStatusCode: stdHttp.StatusInternalServerError,
			assertBody: func(body io.Reader) {
				errorMsg := struct {
					Error string `json:"error"`
				}{}
				err := json.NewDecoder(body).Decode(&errorMsg)
				assert.Nil(t, err)
				assert.Equal(t, "error getting asset catalog", errorMsg.Error)
			},
		},
		"it should return internal server error if delete asset repo call fails": {
			mocks: func() {
				catalogRepository.EXPECT().GetAsset(expectedItemID).Return(&catalog.Asset{ID: expectedItemID}, nil)
				catalogRepository.EXPECT().DeleteAsset(expectedItemID).Return(errors.New("expected error"))
			},
			request: func() *stdHttp.Request {
				return httptest.NewRequest(stdHttp.MethodDelete, "/catalog/assets/"+expectedItemID.String(), nil)
			},
			expectedStatusCode: stdHttp.StatusInternalServerError,
			assertBody: func(body io.Reader) {
				errorMsg := struct {
					Error string `json:"error"`
				}{}
				err := json.NewDecoder(body).Decode(&errorMsg)
				assert.Nil(t, err)
				assert.Equal(t, "error deleting asset catalog", errorMsg.Error)
			},
		},
		"it should return no content if the asset is successfully deleted": {
			mocks: func() {
				catalogRepository.EXPECT().GetAsset(expectedItemID).Return(&catalog.Asset{ID: expectedItemID}, nil)
				catalogRepository.EXPECT().DeleteAsset(expectedItemID).Return(nil)
			},
			request: func() *stdHttp.Request {
				return httptest.NewRequest(stdHttp.MethodDelete, "/catalog/assets/"+expectedItemID.String(), nil)
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
			controller.Delete(rec, tc.request())

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatusCode, res.StatusCode)
			tc.assertBody(res.Body)
		})
	}
}

func TestFindCatalogItems(t *testing.T) {
	ctrl := gomock.NewController(t)
	catalogRepository := mocks.NewMockCatalogRepository(ctrl)
	controller, err := http.NewCatalogController(catalogRepository)
	assert.NotNil(t, controller)
	assert.Nil(t, err)

	testCases := map[string]struct {
		mocks              func()
		request            func() *stdHttp.Request
		expectedStatusCode int
		assertBody         func(body io.Reader)
	}{
		"it should return bad request if limit param is invalid": {
			mocks: func() {},
			request: func() *stdHttp.Request {
				return httptest.NewRequest(stdHttp.MethodGet, "/catalog/assets?limit=notanint", nil)
			},
			expectedStatusCode: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				var errResp struct {
					Error string `json:"error"`
				}
				err := json.NewDecoder(body).Decode(&errResp)
				assert.Nil(t, err)
				assert.Contains(t, errResp.Error, "error getting limit")
			},
		},
		"it should return bad request if offset param is invalid": {
			mocks: func() {},
			request: func() *stdHttp.Request {
				return httptest.NewRequest(stdHttp.MethodGet, "/catalog/assets?offset=oops", nil)
			},
			expectedStatusCode: stdHttp.StatusBadRequest,
			assertBody: func(body io.Reader) {
				var errResp struct {
					Error string `json:"error"`
				}
				err := json.NewDecoder(body).Decode(&errResp)
				assert.Nil(t, err)
				assert.Contains(t, errResp.Error, "error getting offset")
			},
		},
		"it should return internal server error if repository fails": {
			mocks: func() {
				catalogRepository.EXPECT().
					FindAssets(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("expected error"))
			},
			request: func() *stdHttp.Request {
				return httptest.NewRequest(stdHttp.MethodGet, "/catalog/assets?limit=10&offset=0", nil)
			},
			expectedStatusCode: stdHttp.StatusInternalServerError,
			assertBody: func(body io.Reader) {
				var errResp struct {
					Error string `json:"error"`
				}
				err := json.NewDecoder(body).Decode(&errResp)
				assert.Nil(t, err)
				assert.Equal(t, "error finding catalog assets", errResp.Error)
			},
		},
		"it should return 200 and the list of assets": {
			mocks: func() {
				book := &catalog.Book{
					Title:       "1984",
					Author:      "George Orwell",
					Publisher:   "Secker & Warburg",
					ISBN:        "978-0451524935",
					PageCount:   328,
					PublishedAt: time.Date(1949, time.June, 8, 0, 0, 0, 0, time.UTC),
				}
				asset := &catalog.Asset{
					ID:       uuid.New(),
					Category: catalog.AssetCategoryBook,
					Info:     book,
				}
				catalogRepository.EXPECT().
					FindAssets(gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]*catalog.Asset{asset}, nil)
			},
			request: func() *stdHttp.Request {
				return httptest.NewRequest(stdHttp.MethodGet, "/catalog/assets?limit=10&offset=0", nil)
			},
			expectedStatusCode: stdHttp.StatusOK,
			assertBody: func(body io.Reader) {
				var assets []*catalog.Asset
				err := json.NewDecoder(body).Decode(&assets)
				assert.Nil(t, err)
				assert.Len(t, assets, 1)
				book, ok := assets[0].Info.(map[string]any)
				assert.True(t, ok)
				log.Println(book)
				assert.Equal(t, "1984", book["title"])
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

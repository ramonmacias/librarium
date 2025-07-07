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

	"github.com/golang/mock/gomock"
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

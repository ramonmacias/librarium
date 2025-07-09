package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	stdHttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"librarium/internal/http"
)

// Sample type to use with the generic DecodeRequest
type Sample struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestDecodeRequest(t *testing.T) {
	testCases := map[string]struct {
		requestBody    io.Reader
		expectedResult *Sample
		expectedErrStr string
	}{
		"it should return error for nil body": {
			requestBody:    nil,
			expectedResult: nil,
			expectedErrStr: "empty request body",
		},
		"it should return error for invalid JSON": {
			requestBody:    bytes.NewBufferString("{invalid-json}"),
			expectedResult: nil,
			expectedErrStr: "invalid character",
		},
		"it should decode valid JSON": {
			requestBody: func() io.Reader {
				data, _ := json.Marshal(Sample{Name: "Alice", Age: 30})
				return bytes.NewReader(data)
			}(),
			expectedResult: &Sample{Name: "Alice", Age: 30},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(stdHttp.MethodPost, "/", tc.requestBody)
			result, err := http.DecodeRequest[Sample](req)
			if tc.expectedErrStr != "" {
				assert.ErrorContains(t, err, tc.expectedErrStr)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestWriteResponse(t *testing.T) {
	testCases := map[string]struct {
		statusCode        int
		response          any
		expectedStatus    int
		expectedBodyCheck func(t *testing.T, body string)
	}{
		"it should return 204 No Content for a nil response": {
			statusCode:     stdHttp.StatusOK,
			response:       nil,
			expectedStatus: stdHttp.StatusNoContent,
			expectedBodyCheck: func(t *testing.T, body string) {
				assert.Empty(t, body)
			},
		},
		"it should return json with error message": {
			statusCode:     stdHttp.StatusBadRequest,
			response:       errors.New("expected error"),
			expectedStatus: stdHttp.StatusBadRequest,
			expectedBodyCheck: func(t *testing.T, body string) {
				var resp struct {
					Error string `json:"error"`
				}
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "expected error", resp.Error)
			},
		},
		"it should return json encoded body": {
			statusCode: stdHttp.StatusOK,
			response: struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}{
				ID: "123", Name: "John",
			},
			expectedStatus: stdHttp.StatusOK,
			expectedBodyCheck: func(t *testing.T, body string) {
				var resp struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				}
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "123", resp.ID)
				assert.Equal(t, "John", resp.Name)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			http.WriteResponse(rec, tc.statusCode, tc.response)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatus, res.StatusCode)

			bodyBytes, err := io.ReadAll(res.Body)
			assert.NoError(t, err)
			tc.expectedBodyCheck(t, string(bodyBytes))

			if tc.response != nil {
				assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
			}
		})
	}
}

func TestJSONMiddleware(t *testing.T) {
	mux := stdHttp.NewServeMux()
	mux.HandleFunc("POST /librarian", func(w stdHttp.ResponseWriter, r *stdHttp.Request) {
		w.WriteHeader(stdHttp.StatusOK)
	})
	mux.HandleFunc("PUT /customer", func(w stdHttp.ResponseWriter, r *stdHttp.Request) {
		w.WriteHeader(stdHttp.StatusOK)
	})
	mux.HandleFunc("GET /catalog", func(w stdHttp.ResponseWriter, r *stdHttp.Request) {
		w.WriteHeader(stdHttp.StatusOK)
	})
	ts := httptest.NewServer(http.JsonContentTypeMiddleware(mux))
	defer ts.Close()

	testCases := map[string]struct {
		request        func() *stdHttp.Request
		assertResponse func(rsp *stdHttp.Response)
	}{
		"it should return no error non check content header if the http method is not POST or PUT": {
			request: func() *stdHttp.Request {
				req, err := stdHttp.NewRequest(stdHttp.MethodGet, ts.URL+"/catalog", stdHttp.NoBody)
				assert.Nil(t, err)
				return req
			},
			assertResponse: func(rsp *stdHttp.Response) {
				assert.NotNil(t, rsp)
				assert.Equal(t, stdHttp.StatusOK, rsp.StatusCode)
			},
		},
		"it should return an error if the http method is POST and the content type header is missing": {
			request: func() *stdHttp.Request {
				req, err := stdHttp.NewRequest(stdHttp.MethodPost, ts.URL+"/librarian", stdHttp.NoBody)
				assert.Nil(t, err)
				return req
			},
			assertResponse: func(rsp *stdHttp.Response) {
				assert.NotNil(t, rsp)
				assert.Equal(t, stdHttp.StatusBadRequest, rsp.StatusCode)
				errorMsg := struct {
					Error string `json:"error"`
				}{}
				err := json.NewDecoder(rsp.Body).Decode(&errorMsg)
				assert.Nil(t, err)
				err = rsp.Body.Close()
				assert.Nil(t, err)
				assert.Equal(t, "Content-Type must be application/json", errorMsg.Error)
			},
		},
		"it should return an error if the http method is PUT and the content type header is missing": {
			request: func() *stdHttp.Request {
				req, err := stdHttp.NewRequest(stdHttp.MethodPut, ts.URL+"/customer", stdHttp.NoBody)
				assert.Nil(t, err)
				return req
			},
			assertResponse: func(rsp *stdHttp.Response) {
				assert.NotNil(t, rsp)
				assert.Equal(t, stdHttp.StatusBadRequest, rsp.StatusCode)
				errorMsg := struct {
					Error string `json:"error"`
				}{}
				err := json.NewDecoder(rsp.Body).Decode(&errorMsg)
				assert.Nil(t, err)
				err = rsp.Body.Close()
				assert.Nil(t, err)
				assert.Equal(t, "Content-Type must be application/json", errorMsg.Error)
			},
		},
		"it should return no error if the http method is POST and the content type header is given": {
			request: func() *stdHttp.Request {
				req, err := stdHttp.NewRequest(stdHttp.MethodPost, ts.URL+"/librarian", stdHttp.NoBody)
				assert.Nil(t, err)
				req.Header.Add("Content-Type", "application/json")
				return req
			},
			assertResponse: func(rsp *stdHttp.Response) {
				assert.NotNil(t, rsp)
				assert.Equal(t, stdHttp.StatusOK, rsp.StatusCode)
			},
		},
		"it should return no error if the http method is PUT and the content type header is given": {
			request: func() *stdHttp.Request {
				req, err := stdHttp.NewRequest(stdHttp.MethodPut, ts.URL+"/customer", stdHttp.NoBody)
				assert.Nil(t, err)
				req.Header.Add("Content-Type", "application/json")
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

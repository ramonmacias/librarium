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

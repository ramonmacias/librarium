package query_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"librarium/internal/query"
)

func TestFilterFromHTTPRequest(t *testing.T) {
	testCases := map[string]struct {
		expected query.Filters
		query    func(req *http.Request)
	}{
		"it should return no filters if we pass no query params": {
			expected: query.Filters{},
			query:    func(req *http.Request) {},
		},
		"it should return no filters if we pass only pagination params and not filters": {
			expected: query.Filters{},
			query: func(req *http.Request) {
				q := req.URL.Query()
				q.Add("limit", "10")
				q.Add("offset", "0")
				q.Add("sort_by", "created_at")
				q.Add("descending", "true")
				req.URL.RawQuery = q.Encode()
			},
		},
		"it should return scalar equal filters": {
			expected: query.Filters{
				"filter_test_1": query.Filter{
					Type:  query.FilterTypeEqual,
					Value: "test_value_1",
				},
				"filter_test_2": query.Filter{
					Type:  query.FilterTypeEqual,
					Value: "test_value_2",
				},
			},
			query: func(req *http.Request) {
				q := req.URL.Query()
				q.Add("filter_test_1", "test_value_1")
				q.Add("filter_test_2", "test_value_2")
				req.URL.RawQuery = q.Encode()
			},
		},
		"it should return in filters": {
			expected: query.Filters{
				"filter_test_1": query.Filter{
					Type:  query.FilterTypeIn,
					Value: []string{"test_value_1", "test_value_2"},
				},
			},
			query: func(req *http.Request) {
				q := req.URL.Query()
				q.Add("filter_test_1[]", "test_value_1")
				q.Add("filter_test_1[]", "test_value_2")
				req.URL.RawQuery = q.Encode()
			},
		},
		"it should return scalar gte and lte filters": {
			expected: query.Filters{
				"created_at": query.Filter{
					Type: query.FilterTypeRange,
					Value: query.RangeFilter{
						From: "2022-01-02",
						To:   "2023-01-02",
					},
				},
			},
			query: func(req *http.Request) {
				q := req.URL.Query()
				q.Add("created_at_from", "2022-01-02")
				q.Add("created_at_to", "2023-01-02")
				req.URL.RawQuery = q.Encode()
			},
		},
		"it should return scalar gte and lte filter but checking the _from": {
			expected: query.Filters{
				"updated_at": query.Filter{
					Type: query.FilterTypeRange,
					Value: query.RangeFilter{
						From: "2022-01-02",
						To:   "2023-01-02",
					},
				},
			},
			query: func(req *http.Request) {
				var b strings.Builder
				b.WriteString("updated_at_to=2023-01-02")
				b.WriteString("&updated_at_from=2022-01-02")
				req.URL.RawQuery = b.String()
			},
		},
		"it should return a mixed type of filters": {
			expected: query.Filters{
				"filter_test_1": query.Filter{
					Type:  query.FilterTypeIn,
					Value: []string{"test_value_1", "test_value_2"},
				},
				"filter_test_2": query.Filter{
					Type:  query.FilterTypeEqual,
					Value: "test_value_2",
				},
				"created_at": query.Filter{
					Type:  query.FilterTypeRange,
					Value: query.RangeFilter{To: "2023-01-02"},
				},
			},
			query: func(req *http.Request) {
				q := req.URL.Query()
				q.Add("filter_test_1[]", "test_value_1")
				q.Add("filter_test_1[]", "test_value_2")
				q.Add("filter_test_2", "test_value_2")
				q.Add("created_at_to", "2023-01-02")
				req.URL.RawQuery = q.Encode()
			},
		},
		"it should return scalar not equal filters": {
			expected: query.Filters{
				"filter_test": query.Filter{
					Type:  query.FilterTypeNotEqual,
					Value: "test_value",
				},
			},
			query: func(req *http.Request) {
				q := req.URL.Query()
				q.Add("filter_test_not", "test_value")
				req.URL.RawQuery = q.Encode()
			},
		},
		"it should return like filters": {
			expected: query.Filters{
				"filter_test": query.Filter{
					Type:  query.FilterTypeLike,
					Value: "test_value",
				},
			},
			query: func(req *http.Request) {
				q := req.URL.Query()
				q.Add("filter_test_like", "test_value")
				req.URL.RawQuery = q.Encode()
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "http://www.test.io", http.NoBody)
			assert.Nil(t, err)
			tc.query(req)

			filters := query.FiltersFromHTTPRequest(req)
			assert.Equal(t, tc.expected, filters)
		})
	}
}

func TestPaginationFromHTTPRequest(t *testing.T) {
	testCases := map[string]struct {
		input       *http.Request
		expected    *query.Pagination
		expectedErr error
	}{
		"it should fail if the limit param is not a number": {
			input:       buildRequestWithParams(t, map[string]string{"limit": "not_a_number"}),
			expectedErr: errors.New(`error getting limit from http request: strconv.Atoi: parsing "not_a_number": invalid syntax`),
		},
		"it should fail if the offset param is not a number": {
			input:       buildRequestWithParams(t, map[string]string{"offset": "not_a_number"}),
			expectedErr: errors.New(`error getting offset from http request: strconv.Atoi: parsing "not_a_number": invalid syntax`),
		},
		"it should work if both limit and offset are numbers": {
			input: buildRequestWithParams(t, map[string]string{"offset": "20", "limit": "25"}),
			expected: &query.Pagination{
				Limit:  25,
				Offset: 20,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			pag, err := query.PaginationFromHTTPRequest(tc.input)
			if tc.expectedErr != nil {
				assert.Nil(t, pag)
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.expected, pag)
			}
		})
	}
}

func TestSortingFromHTTPRequest(t *testing.T) {
	testCases := map[string]struct {
		input    *http.Request
		expected *query.Sorting
	}{
		"it should return a sort by value and descending false by default": {
			input: buildRequestWithParams(t, map[string]string{"sort_by": "created_at"}),
			expected: &query.Sorting{
				SortBy:     "created_at",
				Descending: false,
			},
		},
		"it should return a sort by value and descending true": {
			input: buildRequestWithParams(t, map[string]string{"sort_by": "created_at", "descending": "true"}),
			expected: &query.Sorting{
				SortBy:     "created_at",
				Descending: true,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			sort := query.SortingFromHTTPRequest(tc.input)
			assert.Equal(t, tc.expected, sort)
		})
	}
}

func buildRequestWithParams(t *testing.T, params map[string]string) *http.Request {
	req, err := http.NewRequest(http.MethodGet, "http://www.test.io", http.NoBody)
	assert.Nil(t, err)

	q := req.URL.Query()
	for key, val := range params {
		q.Add(key, val)
	}
	req.URL.RawQuery = q.Encode()
	return req
}

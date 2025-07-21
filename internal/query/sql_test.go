package query_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"librarium/internal/query"
)

var fields = query.DatabaseFields{
	"id":               "id",
	"some_field":       "some_table.some_field",
	"some_other_field": "some_table.some_other_field",
	"custom_field":     "custom_field",
}

func TestSQLPaginateBy(t *testing.T) {
	testCases := map[string]struct {
		pagination  *query.Pagination
		expectedRes string
	}{
		"it should apply pagination to the query": {
			pagination:  query.NewPagination(100, 10),
			expectedRes: "LIMIT 10 OFFSET 100",
		},
		"it should apply no pagination to the query if it's nil": {
			expectedRes: "",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expectedRes, query.SQLPaginateBy(tc.pagination))
		})
	}
}

func TestSQLSortBy(t *testing.T) {
	sorts := []query.Sorting{
		{SortBy: "some_field", Descending: true},
	}
	testCases := map[string]struct {
		sorts       []query.Sorting
		expectedRes string
	}{
		"it should apply the expected sort to the query": {
			sorts:       sorts,
			expectedRes: "ORDER BY some_table.some_field DESC",
		},
		"it should apply multiple sorts": {
			sorts: []query.Sorting{
				{SortBy: "some_field", Descending: true},
				{SortBy: "some_other_field", Descending: false},
			},
			expectedRes: "ORDER BY some_table.some_field DESC, some_table.some_other_field ASC",
		},
		"it should apply no sorts if not within the expected fields": {
			sorts:       []query.Sorting{{SortBy: "invalid-field"}},
			expectedRes: "",
		},
		"it should apply no sorts if empty": {
			sorts:       []query.Sorting{},
			expectedRes: "",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expectedRes, query.SQLSortBy(tc.sorts, fields))
		})
	}
}

func TestSQLFilterBy(t *testing.T) {
	testCases := map[string]struct {
		filters     query.Filters
		expectedRes string
	}{
		"it should apply no filters if empty": {
			filters:     query.Filters{},
			expectedRes: "",
		},
		"it should apply no filter if the fields is not in the expected filters": {
			filters: query.Filters{
				"unexpected_field": query.Filter{Type: query.FilterTypeEqual, Value: "unexpected_value"},
			},
			expectedRes: "",
		},
		"it should apply equal filter": {
			filters: query.Filters{
				"some_field": query.Filter{Type: query.FilterTypeEqual, Value: "some-value"},
			},
			expectedRes: `some_table.some_field = 'some-value'`,
		},
		"it should apply in filter": {
			filters: query.Filters{
				"some_field": query.Filter{Type: query.FilterTypeIn, Value: []string{"some", "values"}},
			},
			expectedRes: `some_table.some_field IN ('some','values')`,
		},
		"it should apply not in filter": {
			filters: query.Filters{
				"some_field": query.Filter{Type: query.FilterTypeNotIn, Value: []string{"some", "values"}},
			},
			expectedRes: `some_table.some_field NOT IN ('some','values')`,
		},
		"it should apply greater equal filter": {
			filters: query.Filters{
				"some_field": query.Filter{Type: query.FilterTypeGreaterEqual, Value: "some-value"},
			},
			expectedRes: `some_table.some_field >= 'some-value'`,
		},
		"it should apply greater filter": {
			filters: query.Filters{
				"some_field": query.Filter{Type: query.FilterTypeGreater, Value: "some-value"},
			},
			expectedRes: `some_table.some_field > 'some-value'`,
		},
		"it should apply lower equal filter": {
			filters: query.Filters{
				"some_field": query.Filter{Type: query.FilterTypeLowerEqual, Value: "some-value"},
			},
			expectedRes: `some_table.some_field <= 'some-value'`,
		},
		"it should apply lower filter": {
			filters: query.Filters{
				"some_field": query.Filter{Type: query.FilterTypeLower, Value: "some-value"},
			},
			expectedRes: `some_table.some_field < 'some-value'`,
		},
		"it should apply range filter": {
			filters: query.Filters{
				"some_field": query.Filter{Type: query.FilterTypeRange, Value: query.RangeFilter{From: "some", To: "value"}},
			},
			expectedRes: `some_table.some_field >= 'some' AND some_table.some_field <= 'value'`,
		},
		"it should apply only tail range filter": {
			filters: query.Filters{
				"some_field": query.Filter{Type: query.FilterTypeRange, Value: query.RangeFilter{To: "value"}},
			},
			expectedRes: `some_table.some_field <= 'value'`,
		},
		"it should apply not equal filter": {
			filters: query.Filters{
				"some_field": query.Filter{Type: query.FilterTypeNotEqual, Value: "some-value"},
			},
			expectedRes: `some_table.some_field <> 'some-value'`,
		},
		"it should apply like filter": {
			filters: query.Filters{
				"some_field": query.Filter{Type: query.FilterTypeLike, Value: "some-value"},
			},
			expectedRes: `some_table.some_field LIKE '%some-value%'`,
		},
		"it should apply no filters if unknown": {
			filters: query.Filters{
				"some_field": query.Filter{Type: query.FilterTypeUnknown, Value: "some-value"},
			},
			expectedRes: ``,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expectedRes, query.SQLFilterBy(tc.filters, fields))
		})
	}
}

func TestSQLFullQuery(t *testing.T) {
	t.Run("it should create a full query with all the query options", func(t *testing.T) {
		q := fmt.Sprintf(
			`SELECT * FROM "some_table" WHERE %s %s %s`,
			query.SQLFilterBy(query.Filters{
				"id":         query.Filter{Type: query.FilterTypeEqual, Value: "some-id"},
				"some_field": query.Filter{Type: query.FilterTypeLower, Value: "some-value"},
			}, fields),
			query.SQLSortBy([]query.Sorting{{SortBy: "some_field", Descending: true}}, fields),
			query.SQLPaginateBy(query.NewPagination(10, 100)),
		)

		// Since the filters is a map we cant guarantee the order
		expectedQ1 := `SELECT * FROM "some_table" WHERE id = 'some-id' AND some_table.some_field < 'some-value' ORDER BY some_table.some_field DESC LIMIT 100 OFFSET 10`
		expectedQ2 := `SELECT * FROM "some_table" WHERE some_table.some_field < 'some-value' AND id = 'some-id' ORDER BY some_table.some_field DESC LIMIT 100 OFFSET 10`

		if expectedQ1 == q {
			assert.Equal(t, expectedQ1, q)
		} else if expectedQ2 == q {
			assert.Equal(t, expectedQ2, q)
		} else {
			assert.Fail(t, fmt.Sprintf("Query doesnt match, got %s", q))
		}
	})
}

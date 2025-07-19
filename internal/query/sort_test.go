package query_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"librarium/internal/query"
)

func TestSQLDirection(t *testing.T) {
	testCases := map[string]struct {
		sort              query.Sorting
		expectedDirection string
	}{
		"it should return an ASC direction": {
			sort:              query.Sorting{Descending: false},
			expectedDirection: "ASC",
		},
		"it should return a DESC direction": {
			sort:              query.Sorting{Descending: true},
			expectedDirection: "DESC",
		},
		"it should return an ASC if we don't setup the descending flag": {
			sort:              query.Sorting{},
			expectedDirection: "ASC",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expectedDirection, tc.sort.SQLDirection())
		})
	}
}

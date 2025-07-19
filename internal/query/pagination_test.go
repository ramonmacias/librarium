package query_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"librarium/internal/query"
)

func TestNewPagination(t *testing.T) {
	t.Run("it should return the expected pagination if values are within the limits", func(t *testing.T) {
		pag := query.NewPagination(5, 10)
		assert.Equal(t, 5, pag.Offset)
		assert.Equal(t, 10, pag.Limit)
	})
	t.Run("it should return the expected pagination if values are not within the limits", func(t *testing.T) {
		pag := query.NewPagination(-10, 300)
		assert.Equal(t, 0, pag.Offset)
		assert.Equal(t, 100, pag.Limit)
	})
}

func TestLastPage(t *testing.T) {
	t.Run("it should return true if it's the last page", func(t *testing.T) {
		pag := &query.Pagination{
			Offset:     0,
			Limit:      100,
			TotalItems: 50,
		}
		assert.True(t, pag.LastPage())
	})
	t.Run("it should return false if it's not the last page", func(t *testing.T) {
		pag := &query.Pagination{
			Offset:     50,
			Limit:      50,
			TotalItems: 150,
		}
		assert.False(t, pag.LastPage())
	})
}

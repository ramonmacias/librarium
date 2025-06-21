package query

const (
	// maxLimit represents the maximum limit allowed in the pagination.
	// If we try to create a pagination with a greater limit it will be
	// set to this value.
	maxLimit = 100
)

// NewPagination returns a pagination object from the provided page and limit.
// If the limit is higher maximum limit we force it to the maxLimit.
// If the page is invalid we apply the first page.
func NewPagination(offset, limit int) *Pagination {
	if limit > maxLimit || limit <= 0 {
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}

	return &Pagination{
		Offset: offset,
		Limit:  limit,
	}
}

// Pagination holds the information needed
// for handle a paginated request and response
type Pagination struct {
	// Offset is the field used for determine from where we need
	// to apply the limit, it is used for move between pages
	Offset int `json:"offset"`

	// Limit is the field used for determine the max number
	// items per page.
	Limit int `json:"limit"`

	// TotalItems is the total number of items that we have in our
	// db for the given asked resource
	TotalItems int `json:"total_items"`
}

// LastPage will check if the current offset is the last page
// for the given pagination
func (p *Pagination) LastPage() bool {
	return p.Offset+p.Limit >= p.TotalItems
}

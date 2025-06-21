package query

// Sorting holds the information needed for
// handle a sorted request and response
type Sorting struct {
	// Is used for determine which field is the one used
	// for sort the request, currently it only allows to
	// sort by one column
	SortBy string

	// If the descending is true then we apply a descendent
	// order into our query, otherwise we apply ascendent ordering
	Descending bool
}

// SQLDirection returns ASC if it's ascending and DESC if it's descending
func (s Sorting) SQLDirection() string {
	if s.Descending {
		return "DESC"
	}
	return "ASC"
}

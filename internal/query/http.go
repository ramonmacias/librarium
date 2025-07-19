package query

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

const (
	limitParam      = "limit"
	offsetParam     = "offset"
	descendingParam = "descending"
	sortByParam     = "sort_by"
)

// PaginationFromHTTPRequest retrieves the limit and the offset values
// from the given req *http.Request by looking into the expected
// limit and offset query parameters, it returns error in case those
// params are not in the expected int format
func PaginationFromHTTPRequest(req *http.Request) (*Pagination, error) {
	var limit, offset int
	var err error

	limitStr := req.URL.Query().Get(limitParam)
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return nil, fmt.Errorf("error getting limit from http request: %w", err)
		}
	}

	offsetStr := req.URL.Query().Get(offsetParam)
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			return nil, fmt.Errorf("error getting offset from http request: %w", err)
		}
	}

	return NewPagination(offset, limit), nil
}

// SortingFromHTTPRequest retrieves the sortBy and descending boolean
// from the given req *http.Request by looking into the query parameters
func SortingFromHTTPRequest(req *http.Request) *Sorting {
	descending := false
	if req.URL.Query().Get(descendingParam) == "true" {
		descending = true
	}

	return &Sorting{
		SortBy:     req.URL.Query().Get(sortByParam),
		Descending: descending,
	}
}

// FiltersFromHTTPRequest gets the filters that are coming
// from the given http.Request as a query parameters
func FiltersFromHTTPRequest(req *http.Request) Filters {
	filters := Filters{}
	for paramName, value := range req.URL.Query() {
		if isParameterAllowed(paramName) {
			filters[cleanUpField(paramName)] = filterFromParamName(paramName, value, filters)
		}
	}

	return filters
}

// isParameterAllowed checks whether the provided parameter is one of the reserved parameters names
func isParameterAllowed(name string) bool {
	return name != limitParam && name != offsetParam && name != sortByParam && name != descendingParam
}

// filterFromParamName transforms the provided query filter into the expected system Filter
func filterFromParamName(name string, value []string, filters Filters) Filter {
	if strings.Contains(name, "[]") {
		return Filter{Type: FilterTypeIn, Value: value}
	} else if strings.HasSuffix(name, "_from") {
		// We need to check if the range filter for this field already exists to grab the To value
		if f, ok := filters[cleanUpField(name)]; ok {
			if f.Type == FilterTypeRange {
				return Filter{Type: FilterTypeRange, Value: RangeFilter{From: value[0], To: f.Value.(RangeFilter).To}}
			}
		}
		return Filter{Type: FilterTypeRange, Value: RangeFilter{From: value[0]}}
	} else if strings.HasSuffix(name, "_to") {
		// We need to check if the range filter for this field already exists to grab the From value
		if f, ok := filters[cleanUpField(name)]; ok {
			if f.Type == FilterTypeRange {
				return Filter{Type: FilterTypeRange, Value: RangeFilter{From: f.Value.(RangeFilter).From, To: value[0]}}
			}
		}
		return Filter{Type: FilterTypeRange, Value: RangeFilter{To: value[0]}}
	} else if strings.HasSuffix(name, "_not") {
		return Filter{Type: FilterTypeNotEqual, Value: value[0]}
	} else if strings.HasSuffix(name, "_like") {
		return Filter{Type: FilterTypeLike, Value: value[0]}
	}
	return Filter{Type: FilterTypeEqual, Value: value[0]}
}

// cleanUpField removes unneeded suffixes from the field name: [], _from, _to, _like, _not
// This is helpful to remove the suffixes from fields coming from a http query filter.
func cleanUpField(field string) string {
	field = strings.TrimSuffix(field, "[]")
	field = strings.TrimSuffix(field, "_from")
	field = strings.TrimSuffix(field, "_to")
	field = strings.TrimSuffix(field, "_like")
	field = strings.TrimSuffix(field, "_not")
	return field
}

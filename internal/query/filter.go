// Package query provides a common implementation for filters, pagination and sorting across the company.
//
// It provides the definition of different structs and functions to apply these query options from the grpc
// to the persistance layer.
//
// The query package provides transforming functions to convert from the already common proto query structures
// to the defined query structures.
//
// Lastly, it provides the implementation of GORM functions to apply the query options to the given database query.
// This can be expanded in the future to the needed persistance platforms or packages used.
package query

// FilterType used for determine which kind of filter should be applied to the query.
type FilterType int

const (
	// FilterTypeUnknown used for check the case
	// that the filter is not informed
	FilterTypeUnknown FilterType = iota
	// FilterTypeIn used for check the case that
	// the filter is used for query by an slice of values
	FilterTypeIn
	// FilterTypeNotIn used for filters that apply a NOT IN condition
	FilterTypeNotIn
	// FilterTypeEqual used for check the case that
	// the filter is used for query by one specific value
	FilterTypeEqual
	// FilterTypeGreaterEqual used for manage filters like >=
	FilterTypeGreaterEqual
	// FilterTypeGreater used for manage filters like >
	FilterTypeGreater
	// FilterTypeLowerEqual used for manage filters like <=
	FilterTypeLowerEqual
	// FilterTypeLower used for manage filters like <
	FilterTypeLower
	// FilterTypeRange used for check the case that
	// the filter is used to query with a from/to range
	FilterTypeRange
	// FilterTypeNotEqual used for check the case that
	// the filter is used for query by a not equal condition
	FilterTypeNotEqual
	// FilterTypeLike used for filters that apply a LIKE condition
	FilterTypeLike
)

// Filters type used to hold the multiple filters
// that can be applied to one query.
// The key of this map is the field name and the value the filter
// to apply.
type Filters map[string]Filter

// Filter holds the basic structure needed for determine
// which kind of filter and for which value should be applied.
type Filter struct {
	Type  FilterType
	Value any
}

// RangeFilter holds the from/to information needed for a range filter.
type RangeFilter struct {
	From string
	To   string
}

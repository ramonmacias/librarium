package query

import (
	"fmt"
	"strings"
)

// DatabaseFields is a map type used to translate from the give field name
// to the expected database field.
// This is needed specially when doing joins with different tables, for example:
//   - map[string]string{"company_name": "company.name"}
//
// This allows us to filter and sort by company_name and we let the query know to use the
// company table which could be an alias from a join.
type DatabaseFields map[string]string

// SQLPaginateBy returns the sql query in string format to apply the provided pagination.
// If the pagination provided is nil it will not apply any pagination.
func SQLPaginateBy(pag *Pagination) string {
	if pag == nil {
		return ""
	}
	return fmt.Sprintf("LIMIT %d OFFSET %d", pag.Limit, pag.Offset)
}

// SQLSortBy returns the sql query in string format for the provided sorts.
// If descending is not true an ascending sort will be applied by default.
// If the provided sorting field is not within the database fields it will skip it.
func SQLSortBy(sorts []Sorting, dbFields DatabaseFields) string {
	q := ""
	for _, sort := range sorts {
		if _, ok := dbFields[sort.SortBy]; ok {
			if q == "" {
				// It's the first sort, so we need to add the ORDER BY keyword
				q = fmt.Sprintf("ORDER BY %s %s", dbFields[sort.SortBy], sort.SQLDirection())
			} else {
				q = fmt.Sprintf("%s, %s %s", q, dbFields[sort.SortBy], sort.SQLDirection())
			}
		}
	}
	return q
}

// SQLFilterBy returns the sql query in string format for the provided filters.
// It receives the database fields map.
// If the provided filter field is not within the database fields it will skip it.
func SQLFilterBy(filters Filters, dbFields DatabaseFields) string {
	q := ""
	for field, filter := range filters {
		if _, ok := dbFields[field]; !ok {
			continue
		}

		f := filterByType(filter.Type, dbFields[field], filter.Value)
		if f == "" {
			continue
		}

		if q == "" {
			q = f
		} else {
			q = fmt.Sprintf("%s AND %s", q, f)
		}
	}
	return q
}

func filterByType(t FilterType, field string, value any) string {
	switch t {
	case FilterTypeEqual:
		return fmt.Sprintf("%s = '%s'", field, value.(string))
	case FilterTypeIn:
		return fmt.Sprintf("%s IN (%s)", field, formatSQLSlice(value.([]string)))
	case FilterTypeNotIn:
		return fmt.Sprintf("%s NOT IN (%s)", field, formatSQLSlice(value.([]string)))
	case FilterTypeGreaterEqual:
		return fmt.Sprintf("%s >= '%s'", field, value.(string))
	case FilterTypeGreater:
		return fmt.Sprintf("%s > '%s'", field, value.(string))
	case FilterTypeLowerEqual:
		return fmt.Sprintf("%s <= '%s'", field, value.(string))
	case FilterTypeLower:
		return fmt.Sprintf("%s < '%s'", field, value.(string))
	case FilterTypeRange:
		q := ""
		if value.(RangeFilter).From != "" {
			q = fmt.Sprintf("%s >= '%s'", field, value.(RangeFilter).From)
		}
		if value.(RangeFilter).To != "" {
			if q != "" {
				q = fmt.Sprintf("%s AND %s <= '%s'", q, field, value.(RangeFilter).To)
			} else {
				q = fmt.Sprintf("%s <= '%s'", field, value.(RangeFilter).To)
			}
		}
		return q
	case FilterTypeNotEqual:
		return fmt.Sprintf("%s <> '%s'", field, value.(string))
	case FilterTypeLike:
		return fmt.Sprintf("%s LIKE '%s'", field, "%"+value.(string)+"%")
	case FilterTypeUnknown:
	default:
	}
	return ""
}

func formatSQLSlice(ss []string) string {
	res := make([]string, len(ss))
	for i, s := range ss {
		res[i] = fmt.Sprintf("'%s'", s)
	}
	return strings.Join(res, ",")
}

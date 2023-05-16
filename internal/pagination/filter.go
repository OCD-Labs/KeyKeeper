package pagination

import (
	"math"
	"strings"
)

// Filter contains the parsed query_string
type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

// SortColumn checks that the client-provided Sort field matches one of
// the entries in the safelist and if it does, retrieves it.
func (f Filters) SortColumn() string {
	for _, safeVal := range f.SortSafelist {
		if f.Sort == safeVal {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}

	// failsafe to help stop a SQL injection attack occurring.
	panic("unsafe sort parameter: " + f.Sort)
}

// SortDirection instructs a descending or ascending
// order.
func (f Filters) SortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

// Limit return page size.
func (f Filters) Limit() int {
	return f.PageSize
}

// Offset returns page offset.
func (f Filters) Offset() int {
	return (f.Page - 1) * f.PageSize
}

// Provides extra info about the filtered, sorted and paginated
// info returned on 'GET /v1/movies?<query_string>'
type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

// CalcMetadata calculates and return pagination info
func CalcMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}

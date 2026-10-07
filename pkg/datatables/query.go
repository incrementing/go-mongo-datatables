package datatables

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Query structure, used to query the database
type Query struct {
	Collection    string            `json:"collection"`
	Fields        []string          `json:"fields"`
	LegacyFilters []Filter          `json:"legacy_filters"`
	Filters       bson.M            `json:"filters"`
	Aggregation   []bson.M          `json:"aggregation"`
	OrderBy       []Order           `json:"order_by"`
	Limit         int               `json:"limit"`
	Offset        int               `json:"offset"`
	SearchBy      string            `json:"search_by"`
	Searches      map[string]string `json:"search_fields"`
	Output        string            `json:"output"`
	Download      bool              `json:"download"`
}

type Order struct {
	Field string
	Desc  bool
}

type Filter struct {
	Field string
	Value FilterValue
}

type FilterValue struct {
	Type     string `json:"type"`
	Int      int64
	Float    float64
	Str      string
	Bool     bool
	IntArr   []int64 // used for min and max
	FloatArr []float64
	StrArr   []string
}

type Response struct {
	Data          []bson.D
	Count         int64
	FilteredCount int64
}

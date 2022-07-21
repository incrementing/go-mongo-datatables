package datatables

import (
	"fmt"
	"net/http"
	"strconv"
)

type UrlParam struct {
	Key   string
	Value string
}

// ProcessDataTableInput processes form data from DataTables, and builds query structure
func ProcessDataTableInput(r *http.Request, table string) (*Query, error) {
	query := Query{}

	query.Output = "datatable"

	query.TableName = table
	query.SearchBy = r.FormValue("search[value]")

	if r.FormValue("start") != "" {
		// offset is an int
		start, err := strconv.Atoi(r.FormValue("start"))
		if err != nil {
			return nil, fmt.Errorf("start is not an int")
		}
		query.Offset = start
	} else {
		query.Offset = 0
	}

	if r.FormValue("start") != "length" {
		// offset is an int
		length, err := strconv.Atoi(r.FormValue("length"))
		if err != nil {
			return nil, fmt.Errorf("length is not an int")
		}
		query.Limit = length
	} else {
		query.Offset = 0
	}

	query.Searches = make(map[string]string)

	var i = 0
	for {
		// if key "columns[i][data]" is not in range, break
		if r.FormValue("columns["+fmt.Sprint(i)+"][data]") == "" {
			break
		}

		var fieldName = r.FormValue("columns[" + fmt.Sprint(i) + "][name]")

		if fieldName != "" {
			query.Fields = append(query.Fields, r.FormValue("columns["+fmt.Sprint(i)+"][name]"))
		}

		if r.FormValue("columns["+fmt.Sprint(i)+"][searchable]") == "true" {
			if r.FormValue("columns["+fmt.Sprint(i)+"][search][value]") != "" {
				query.Searches[r.FormValue("columns["+fmt.Sprint(i)+"][name]")] =
					r.FormValue("columns[" + fmt.Sprint(i) + "][search][value]")
			}
		}

		i++
	}

	query.OrderBy = make(map[string]bool)

	i = 0
	for {
		if r.FormValue("order["+fmt.Sprint(i)+"][column]") == "" {
			break
		}

		// offset is an int
		column, err := strconv.Atoi(r.FormValue("order[" + fmt.Sprint(i) + "][column]"))
		if err != nil {
			return nil, fmt.Errorf("order column is not an int")
		}

		if r.FormValue("columns["+fmt.Sprint(column)+"][orderable]") == "false" {
			i++
			continue
		}
		query.OrderBy[query.Fields[column]] = r.FormValue("order["+fmt.Sprint(i)+"][dir]") == "desc"
		i++
	}

	query.Download = r.FormValue("download") == "true"

	// return query
	return &query, nil
}

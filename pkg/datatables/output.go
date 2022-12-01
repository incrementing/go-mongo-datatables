package datatables

import (
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
)

type DataTable struct {
	Data            [][]string `json:"data"`
	RecordsFiltered int64      `json:"recordsFiltered"`
	RecordsTotal    int64      `json:"recordsTotal"`
}

type rodModFunc func([]interface{}, string) []string

// GenerateDataTableOutput generates a json DataTable output from mongo docs, with filtered and total records
func GenerateDataTableOutput(data []primitive.D, totalCount int64, filteredCount int64, query *Query, rowmod rodModFunc) string {
	var dataTable DataTable

	// foreach value in data, add to array
	// return array
	var columnList [][]string
	for _, column := range data {
		// make string list
		var rowList []interface{}

		// make map
		columnMap := map[string]interface{}{}
		for _, v := range column {
			columnMap[v.Key] = v.Value
		}

		for _, field := range query.Fields {
			// split field by .
			var fieldList = strings.Split(field, ".")

			// get value from map
			var value = columnMap[fieldList[0]]

			if value == nil {
				rowList = append(rowList, "")
				continue
			}

			for i := 1; i < len(fieldList); i++ {
				// value is likely to be primitive.D
				var valueMap = map[string]interface{}{}
				for _, v := range value.(primitive.D) {
					valueMap[v.Key] = v.Value
				}

				value = valueMap[fieldList[i]]
			}

			// add value to row
			rowList = append(rowList, value)
		}

		columnList = append(columnList, rowmod(rowList, query.SearchBy))
	}

	// if columnList is nil, return empty array
	if columnList == nil {
		columnList = [][]string{}
	}

	dataTable.Data = columnList
	dataTable.RecordsTotal = totalCount
	dataTable.RecordsFiltered = filteredCount

	// json serialize datatable
	buffer, err := json.Marshal(dataTable)
	if err != nil {
		return "Failed to generate data response"
	}

	return string(buffer)
}

package datatables

import (
	"bytes"
	"context"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"html"
	"html/template"
	"net/http"
	"regexp"
	"strings"
)

type ValueModFunc func(interface{}, map[string]interface{}) interface{}

type DataTableValue struct {
	Name    string
	ModFunc ValueModFunc
}

type DataTableEndpoint struct {
	Database        *mongo.Database
	Context         context.Context
	TableName       string
	MaxRows         int
	SearchValues    []string
	HighlightSearch bool
	Values          []DataTableValue
	Row             []string
	Filters         []Filter
}

func HighlightString(haystack, needle string) string {
	needle = html.EscapeString(needle)

	// extract incasesenitive needle from haystack and get real case
	// of the needle
	re := regexp.MustCompile(`(?i)(` + needle + `)`)
	realNeedle := re.FindString(haystack)

	// if realNeedle is empty, return haystack
	if realNeedle == "" {
		return haystack
	}

	//log.Info().Msgf("realNeedle: %s", realNeedle)

	// print realNeedle
	return strings.Replace(haystack, realNeedle, "<span class=\"textHighlighted;\">"+realNeedle+"</span>", -1)
}

func GenerateDataTable(w http.ResponseWriter, r *http.Request, dt *DataTableEndpoint) error {
	query, err := ProcessDataTableInput(r, dt.TableName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Error().Err(err).Msg("Error processing datatables input")
		return nil
	}

	// extract value strings from dt.Values
	valueStrings := make([]string, len(dt.Values))
	for i, v := range dt.Values {
		valueStrings[i] = v.Name
	}

	// validate scans endpoint
	if query.TableName != dt.TableName || query.Filters != nil || query.Limit > dt.MaxRows {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return nil
	}

	query.Fields = valueStrings

	query.LegacyFilters = append(query.LegacyFilters, dt.Filters...)

	response, err := RetrieveDocuments(query, dt.Context, dt.Database, dt.SearchValues)
	if err != nil {
		log.Error().Err(err).Msg("error retrieving users")
		return err
	}

	output := GenerateDataTableOutput(response.Data, response.Count, response.FilteredCount, query, func(row []interface{}, search string) []string {
		// merge values and row into map
		m := make(map[string]interface{})
		for i, v := range dt.Values {
			// replace . with _ in key
			key := strings.Replace(v.Name, ".", "_", -1)

			m[key] = row[i]
		}

		newMap := make(map[string]interface{})

		for i, v := range dt.Values {
			var value = row[i]
			if v.ModFunc != nil {
				value = v.ModFunc(value, m)
			}

			key := strings.Replace(v.Name, ".", "_", -1)

			newMap[key] = value
		}

		m = newMap

		newRow := make([]string, 0)

		// foreach row
		for _, v := range dt.Row {
			t, err := template.New("row").Parse(v)
			if err != nil {
				log.Warn().Err(err).Msg("error parsing template")
				return nil
			}

			var buf bytes.Buffer
			err = t.Execute(&buf, m)

			if err != nil {
				return nil
			}
			newRow = append(newRow, buf.String())
		}

		if dt.HighlightSearch && search != "" {
			for i, v := range newRow {
				// replace incase-sensitive
				// wrap with "<span style=\"background-color: red\">" and "</span>"
				// highlight search
				newRow[i] = HighlightString(v, search)
			}
		}

		return newRow
	})

	_, err = w.Write([]byte(output))
	if err != nil {
		log.Error().Err(err).Msg("error writing response")
		return err
	}

	return nil
}

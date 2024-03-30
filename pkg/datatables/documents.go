package datatables

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func filterValueToInterface(fv FilterValue) interface{} {
	switch fv.Type {
	case "string":
		return fv.Str
	case "int":
		return fv.Int
	case "float":
		return fv.Float
	case "bool":
		return fv.Bool
	case "int_array":
		return fv.IntArr
	case "float_array":
		return fv.FloatArr
	case "string_array":
		return fv.StrArr
	case "null":
		return nil
	default:
		return nil
	}
}

func addFiltersBson(query *Query, currentBson *bson.M, searchFields []string) (bool, bool) {
	var fieldList []string
	var filtered = false
	var searched = false

	for _, filter := range query.LegacyFilters {
		filtered = true
		fieldList = append(fieldList, filter.Field)
	}

	var andBson []bson.M

	if query.SearchBy != "" && searchFields != nil {
		searchBson := bson.M{}
		filtered = true

		searchBson["$or"] = []bson.M{}

		for _, field := range searchFields {
			searchBson["$or"] = append(searchBson["$or"].([]bson.M),
				bson.M{
					field: primitive.Regex{Pattern: query.SearchBy, Options: "i"},
				})
		}

		andBson = append(andBson, searchBson)
		searched = true
	}

	// foreach field, foreach value
	for _, field := range fieldList {
		// foreach filter with field
		var orBson []bson.M

		for _, filter := range query.LegacyFilters {
			if filter.Field == field {
				filterInterface := filterValueToInterface(filter.Value)

				// if filter is array of int64, make bson max/min
				if filter.Value.Type == "int_array" {
					var min = filterInterface.([]int64)[0]
					var max = filterInterface.([]int64)[1]

					orBson = append(orBson,
						bson.M{
							field: bson.M{
								"$gte": min,
								"$lte": max,
							},
						})
				} else if filter.Value.Type == "float_array" {
					var min = filterInterface.([]float64)[0]
					var max = filterInterface.([]float64)[1]

					orBson = append(orBson,
						bson.M{
							field: bson.M{
								"$gte": min,
								"$lte": max,
							},
						})
				} else {

					orBson = append(orBson,
						bson.M{
							field: filterInterface,
						})
				}
			}
		}

		andBson = append(andBson,
			bson.M{
				"$or": orBson,
			})
	}

	if query.Filters != nil {
		// add it as a filter (its an interface_
		andBson = append(andBson, query.Filters)
		filtered = true
	}

	if filtered {
		(*currentBson)["$and"] = andBson
	}

	return filtered, searched
}

// RetrieveDocuments function, used to retrieve documents from the database
// converts the query to a bson.D object, and then calls the mongo.Collection.Find() function
func RetrieveDocuments(query *Query, ctx context.Context, db *mongo.Database, searchFields []string) (*Response, error) {
	collection := db.Collection(query.TableName)

	findOptions := options.Find()

	if query.Limit > 0 {
		findOptions.SetLimit(int64(query.Limit))
	}

	findOptions.SetSkip(int64(query.Offset))

	// Generate orderBy bson.D object (ordered)
	orderByBson := bson.D{}

	// foreach in findOptions.OrderBy
	for field, desc := range query.OrderBy {
		var orderByInt = 1
		if desc == true {
			orderByInt = -1
		}

		// append to bsonD
		orderByBson = append(orderByBson,
			bson.E{Key: field, Value: orderByInt})
	}

	// generate filter bson.M object (unordered)
	fieldsBson := bson.M{}
	for _, field := range query.Fields {
		fieldsBson[field] = 1
	}

	// remove _id field from fields
	fieldsBson["_id"] = 0

	// set find options
	findOptions.Sort = orderByBson
	findOptions.Projection = fieldsBson

	// generate search bson.M object (unordered)
	findBson := bson.M{}

	// TODO: multiple filters on same field using $and or $or
	var fieldList []string

	for _, filter := range query.LegacyFilters {
		fieldList = append(fieldList, filter.Field)
	}

	filtered, searched := addFiltersBson(query, &findBson, searchFields)

	if query.Aggregation != nil && len(query.Aggregation) > 0 {
		aggr := query.Aggregation
		// add match
		aggr = append(aggr, bson.M{
			"$match": findBson,
		})

		// add sort
		limitedAggr := append(aggr, bson.M{
			"$sort": orderByBson,
		})

		// set skip and limit
		limitedAggr = append(limitedAggr, bson.M{
			"$skip": query.Offset,
		})
		limitedAggr = append(limitedAggr, bson.M{
			"$limit": query.Limit,
		})

		cursor, err := collection.Aggregate(ctx, limitedAggr)
		if err != nil {
			return nil, err
		}

		var data []primitive.D
		err = cursor.All(ctx, &data)
		if err != nil {
			return nil, err
		}

		// get counts
		totalCount, err := collection.EstimatedDocumentCount(ctx, nil)

		filteredCountAggr := append(aggr, bson.M{
			"$count": "count",
		})

		cursor, err = collection.Aggregate(ctx, filteredCountAggr)
		if err != nil {
			return nil, err
		}

		var filteredCount []bson.M
		err = cursor.All(ctx, &filteredCount)
		if err != nil {
			return nil, err
		}

		if data == nil {
			data = []primitive.D{}
		}

		var filterCountInt int32
		if len(filteredCount) > 0 {
			filterCountInt = filteredCount[0]["count"].(int32)
		} else {
			filterCountInt = 0
		}

		var response = &Response{
			Data:          data,
			Count:         totalCount,
			FilteredCount: int64(filterCountInt),
		}

		return response, nil
	}

	// execute query
	cursor, err := collection.Find(ctx, findBson, findOptions)
	if err != nil {
		return nil, err
	}

	var data []primitive.D
	err = cursor.All(ctx, &data)
	if err != nil {
		return nil, err
	}

	// set totalCount and filtered for pagination
	totalCount, err := collection.EstimatedDocumentCount(ctx, nil)
	var filteredCount = totalCount

	if filtered || searched {
		filteredCount, err = collection.CountDocuments(ctx, findBson)
	}

	if data == nil {
		data = []primitive.D{}
	}

	var response = &Response{
		Data:          data,
		Count:         totalCount,
		FilteredCount: filteredCount,
	}

	return response, nil
}

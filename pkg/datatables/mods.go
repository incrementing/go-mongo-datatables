package datatables

import (
	"fmt"
	"github.com/incrementing/go-mongo-datatables/pkg/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func DateModFunc(format string) func(item interface{}, row map[string]interface{}) interface{} {
	return func(item interface{}, row map[string]interface{}) interface{} {
		// assert that item to time.Time
		unixTime, ok := item.(primitive.DateTime)
		if !ok {
			// if it's not a primitive.DateTime, it might be a int64
			unixTimeInt64, ok := item.(int64)
			if !ok {
				fmt.Println("item is not a time? we accept DateTime and int64")
				return item
			}

			unixTime = primitive.DateTime(unixTimeInt64 * 1000)
		}

		tm := unixTime.Time()

		if tm.IsZero() {
			return struct {
				Unix      int64
				Formatted string
			}{
				0, "N/A",
			}
		}

		return struct {
			Unix      int64
			Formatted string
		}{
			tm.Unix(),
			tm.Format(format),
		}
	}
}

func DurationModFunc(item interface{}, row map[string]interface{}) interface{} {
	// assert that item to time.Time
	duration, ok := item.(int64)
	if !ok {
		fmt.Println("item is not a time? we accept DateTime and int64")
		return item
	}

	return struct {
		Seconds   int64
		Formatted string
	}{
		duration,
		util.FormatDuration(duration),
	}
}

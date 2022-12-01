package datatables

import (
	"fmt"
	"github.com/josheyr/go-mongo-datatables/pkg/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
)

func DateModFunc(item interface{}, row map[string]interface{}) interface{} {
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

	return struct {
		Unix      int64
		Formatted string
	}{
		tm.Unix(),
		tm.Format("2006-01-02 15:04:05"),
	}

	// format date with nice date format with time
	return "<span style=\"display: none;\">" + strconv.Itoa(int(tm.Unix())) + "</span>" + tm.Format("2006-01-02 15:04:05")
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

package datatables

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
)

func DateModFunc(item interface{}) interface{} {
	// assert that item to time.Time
	unixTime, ok := item.(primitive.DateTime)
	if !ok {
		panic("item is not a int64")
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

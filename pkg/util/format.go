package util

import (
	"fmt"
	"time"
)

func FormatDuration(i int64) string {
	// make duration look like: 3h 4m 5s
	d := time.Duration(i) * time.Second

	// get hours
	hours := int(d.Hours())

	// get minutes
	minutes := int(d.Minutes()) - (hours * 60)

	// get seconds
	seconds := int(d.Seconds()) - (hours * 60 * 60) - (minutes * 60)

	// make string
	return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
}
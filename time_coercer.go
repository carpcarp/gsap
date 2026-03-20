package sap

import (
	"fmt"
	"math"
	"reflect"
	"time"
)

var timeType = reflect.TypeOf(time.Time{})

// coerceToTime attempts to coerce a value to time.Time.
// Supports RFC3339, date-only (2006-01-02), and Unix timestamps.
func (c *TypeCoercer) coerceToTime(value interface{}, score *Score) (interface{}, error) {
	switch v := value.(type) {
	case string:
		// Try RFC3339 first
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			score.AddFlag(FlagStringToTime, 1)
			return t, nil
		}
		// Try RFC3339Nano
		if t, err := time.Parse(time.RFC3339Nano, v); err == nil {
			score.AddFlag(FlagStringToTime, 1)
			return t, nil
		}
		// Try date-only
		if t, err := time.Parse("2006-01-02", v); err == nil {
			score.AddFlag(FlagStringToTime, 1)
			return t, nil
		}
		// Try datetime without timezone
		if t, err := time.Parse("2006-01-02T15:04:05", v); err == nil {
			score.AddFlag(FlagStringToTime, 2)
			return t, nil
		}
		// Try datetime with space separator
		if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
			score.AddFlag(FlagStringToTime, 2)
			return t, nil
		}
		return nil, fmt.Errorf("cannot parse string as time: %s", v)

	case float64:
		// Interpret as Unix timestamp
		// Distinguish seconds vs milliseconds: if > 1e12, treat as milliseconds
		if math.Abs(v) > 1e12 {
			sec := int64(v) / 1000
			nsec := (int64(v) % 1000) * int64(time.Millisecond)
			score.AddFlag(FlagUnixToTime, 2)
			return time.Unix(sec, nsec).UTC(), nil
		}
		score.AddFlag(FlagUnixToTime, 2)
		return time.Unix(int64(v), 0).UTC(), nil

	default:
		return nil, fmt.Errorf("cannot convert %T to time.Time", value)
	}
}

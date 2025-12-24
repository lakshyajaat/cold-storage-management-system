package timeutil

import (
	"time"
)

// IST is the Indian Standard Time location (UTC+5:30)
var IST *time.Location

func init() {
	var err error
	IST, err = time.LoadLocation("Asia/Kolkata")
	if err != nil {
		// Fallback: create fixed zone if Asia/Kolkata not available
		IST = time.FixedZone("IST", 5*60*60+30*60) // UTC+5:30
	}
}

// Now returns the current time in IST
func Now() time.Time {
	return time.Now().In(IST)
}

// ToIST converts any time to IST
func ToIST(t time.Time) time.Time {
	return t.In(IST)
}

// ParseInIST parses a time string and returns it in IST
func ParseInIST(layout, value string) (time.Time, error) {
	t, err := time.ParseInLocation(layout, value, IST)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

// FormatIST formats a time in IST using the given layout
func FormatIST(t time.Time, layout string) string {
	return t.In(IST).Format(layout)
}

// StartOfDay returns the start of day (00:00:00) in IST for the given time
func StartOfDay(t time.Time) time.Time {
	ist := t.In(IST)
	return time.Date(ist.Year(), ist.Month(), ist.Day(), 0, 0, 0, 0, IST)
}

// EndOfDay returns the end of day (23:59:59) in IST for the given time
func EndOfDay(t time.Time) time.Time {
	ist := t.In(IST)
	return time.Date(ist.Year(), ist.Month(), ist.Day(), 23, 59, 59, 999999999, IST)
}

// Common layouts for IST formatting
const (
	DateLayout     = "2006-01-02"
	TimeLayout     = "15:04:05"
	DateTimeLayout = "2006-01-02 15:04:05"
	DisplayLayout  = "02 Jan 2006, 03:04 PM"
)

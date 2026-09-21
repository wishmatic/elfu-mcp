// Package timing renders instants and durations in the forms an agent reads them in.
package timing

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	justNow = 5 * time.Second

	minute = time.Minute
	hour   = time.Hour
	day    = 24 * time.Hour
	week   = 7 * day

	// Months and years are approximations, deliberately: a caller wants "3 months ago", not a calendar difference.
	month = 30 * day
	year  = 365 * day
)

var zoneLessLayouts = []string{
	"2006-01-02T15:04:05.999999999",
	"2006-01-02T15:04:05",
	"2006-01-02T15:04",
	"2006-01-02 15:04:05.999999999",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"2006-01-02",
}

// Stamp is one instant, rendered every way a caller asks for it.
type Stamp struct {
	At time.Time
}

func (s Stamp) ISO() string {
	return s.At.Format(time.RFC3339)
}

func (s Stamp) Unix() int64 {
	return s.At.Unix()
}

func (s Stamp) Zone() string {
	name, _ := s.At.Zone()

	return name
}

func (s Stamp) Offset() string {
	_, seconds := s.At.Zone()

	sign := "+"
	if seconds < 0 {
		sign, seconds = "-", -seconds
	}

	minutes := seconds / 60

	return fmt.Sprintf("%s%02d:%02d", sign, minutes/60, minutes%60)
}

// Readable spells the instant out for people, e.g. "Monday, 21 September 2026 at 22:49 AEST".
func (s Stamp) Readable() string {
	return s.At.Format("Monday, 2 January 2006 at 15:04 MST")
}

// HumanSince describes the distance from then to now the way a person would say it, e.g. "3 hours ago". A then later
// than now reads "in 3 hours" rather than as a negative number of hours.
func HumanSince(then, now time.Time) string {
	elapsed := now.Sub(then)
	if elapsed < 0 {
		return "in " + humanAmount(-elapsed)
	}

	if elapsed < justNow {
		return "just now"
	}

	return humanAmount(elapsed) + " ago"
}

func humanAmount(d time.Duration) string {
	units := []struct {
		size time.Duration
		name string
	}{
		{year, "year"},
		{month, "month"},
		{week, "week"},
		{day, "day"},
		{hour, "hour"},
		{minute, "minute"},
		{time.Second, "second"},
	}

	for _, unit := range units {
		if n := int64(d / unit.size); n > 0 {
			return count(n, unit.name)
		}
	}

	return "a few seconds"
}

func count(n int64, unit string) string {
	if n == 1 {
		return "1 " + unit
	}

	return fmt.Sprintf("%d %ss", n, unit)
}

// Parse accepts the ISO 8601 forms a model is likely to send. A form carrying a zone is honoured as written; a form
// without one is read in loc, which is the deployment's zone.
func Parse(s string, loc *time.Location) (time.Time, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return time.Time{}, errors.New("empty timestamp")
	}

	if at, err := time.Parse(time.RFC3339Nano, trimmed); err == nil {
		return at, nil
	}

	// RFC 3339 requires a T between the date and the time; writing a space there is common anyway.
	spaced := strings.Replace(trimmed, " ", "T", 1)
	if at, err := time.Parse(time.RFC3339Nano, spaced); err == nil {
		return at, nil
	}

	for _, layout := range zoneLessLayouts {
		if at, err := time.ParseInLocation(layout, trimmed, loc); err == nil {
			return at, nil
		}
	}

	return time.Time{}, fmt.Errorf("%q is not an ISO 8601 timestamp", s)
}

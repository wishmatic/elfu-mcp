package timing

import (
	"strings"
	"testing"
	"time"
)

var sydney = time.FixedZone("AEST", 10*60*60)

func TestStampRendersAnInstant(t *testing.T) {
	at := time.Date(2026, time.September, 21, 22, 49, 3, 0, sydney)
	stamp := Stamp{At: at}

	if got := stamp.ISO(); got != "2026-09-21T22:49:03+10:00" {
		t.Errorf("ISO() = %q, want 2026-09-21T22:49:03+10:00", got)
	}

	if got := stamp.Unix(); got != at.Unix() {
		t.Errorf("Unix() = %d, want %d", got, at.Unix())
	}

	if got := stamp.Zone(); got != "AEST" {
		t.Errorf("Zone() = %q, want AEST", got)
	}

	if got := stamp.Offset(); got != "+10:00" {
		t.Errorf("Offset() = %q, want +10:00", got)
	}

	if got := stamp.Readable(); got != "Monday, 21 September 2026 at 22:49 AEST" {
		t.Errorf("Readable() = %q, want Monday, 21 September 2026 at 22:49 AEST", got)
	}
}

func TestStampISOParsesBack(t *testing.T) {
	at := time.Date(2026, time.September, 21, 22, 49, 3, 0, sydney)

	parsed, err := time.Parse(time.RFC3339, Stamp{At: at}.ISO())
	if err != nil {
		t.Fatalf("time.Parse() error: %v", err)
	}

	if !parsed.Equal(at) {
		t.Errorf("parsed = %s, want %s", parsed, at)
	}
}

func TestStampOffset(t *testing.T) {
	tests := []struct {
		name string
		zone *time.Location
		want string
	}{
		{name: "utc", zone: time.UTC, want: "+00:00"},
		{name: "west of utc", zone: time.FixedZone("PYT", -(3*60+30)*60), want: "-03:30"},
		{name: "half hour east", zone: time.FixedZone("NPT", (5*60+45)*60), want: "+05:45"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stamp := Stamp{At: time.Date(2026, time.September, 21, 12, 0, 0, 0, tt.zone)}

			if got := stamp.Offset(); got != tt.want {
				t.Errorf("Offset() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHumanSince(t *testing.T) {
	now := time.Date(2026, time.September, 21, 22, 49, 3, 0, sydney)

	tests := []struct {
		name string
		then time.Time
		want string
	}{
		{name: "same instant", then: now, want: "just now"},
		{name: "under five seconds", then: now.Add(-4 * time.Second), want: "just now"},
		{name: "seconds", then: now.Add(-5 * time.Second), want: "5 seconds ago"},
		{name: "one second is inside the just now window", then: now.Add(-1 * time.Second), want: "just now"},
		{name: "half a minute", then: now.Add(-30 * time.Second), want: "30 seconds ago"},
		{name: "one minute", then: now.Add(-time.Minute), want: "1 minute ago"},
		{name: "minutes", then: now.Add(-12 * time.Minute), want: "12 minutes ago"},
		{name: "one hour", then: now.Add(-time.Hour), want: "1 hour ago"},
		{name: "hours", then: now.Add(-3 * time.Hour), want: "3 hours ago"},
		{name: "just under a day", then: now.Add(-23 * time.Hour), want: "23 hours ago"},
		{name: "one day", then: now.Add(-25 * time.Hour), want: "1 day ago"},
		{name: "days", then: now.Add(-3 * 24 * time.Hour), want: "3 days ago"},
		{name: "one week", then: now.Add(-8 * 24 * time.Hour), want: "1 week ago"},
		{name: "weeks", then: now.Add(-21 * 24 * time.Hour), want: "3 weeks ago"},
		{name: "one month", then: now.Add(-45 * 24 * time.Hour), want: "1 month ago"},
		{name: "months", then: now.Add(-200 * 24 * time.Hour), want: "6 months ago"},
		{name: "one year", then: now.Add(-400 * 24 * time.Hour), want: "1 year ago"},
		{name: "years", then: now.Add(-800 * 24 * time.Hour), want: "2 years ago"},
		{name: "future hours", then: now.Add(3 * time.Hour), want: "in 3 hours"},
		{name: "future seconds", then: now.Add(2 * time.Second), want: "in 2 seconds"},
		{name: "future one day", then: now.Add(30 * time.Hour), want: "in 1 day"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HumanSince(tt.then, now); got != tt.want {
				t.Errorf("HumanSince() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHumanSinceNeverEmpty(t *testing.T) {
	now := time.Now()

	if got := HumanSince(now, now); strings.TrimSpace(got) == "" {
		t.Error("HumanSince() = an empty string, want a phrase")
	}
}

func TestParse(t *testing.T) {
	want := time.Date(2026, time.September, 21, 19, 49, 3, 0, sydney)

	tests := []struct {
		name string
		in   string
		want time.Time
	}{
		{name: "rfc 3339", in: "2026-09-21T19:49:03+10:00", want: want},
		{name: "utc", in: "2026-09-21T09:49:03Z", want: want},
		{name: "fractional seconds", in: "2026-09-21T09:49:03.500Z", want: want.Add(500 * time.Millisecond)},
		{name: "space separator", in: "2026-09-21 19:49:03+10:00", want: want},
		{name: "zone-less date-time", in: "2026-09-21T19:49:03", want: want},
		{name: "zone-less minute precision", in: "2026-09-21T19:49", want: want.Truncate(time.Minute)},
		{name: "zone-less space", in: "2026-09-21 19:49:03", want: want},
		{name: "date alone", in: "2026-09-21", want: want.Add(-19*time.Hour - 49*time.Minute - 3*time.Second)},
		{name: "surrounding whitespace", in: "  2026-09-21T19:49:03+10:00\n", want: want},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.in, sydney)
			if err != nil {
				t.Fatalf("Parse() error: %v", err)
			}

			if !got.Equal(tt.want) {
				t.Errorf("Parse() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestParseZoneLessUsesLocation(t *testing.T) {
	got, err := Parse("2026-09-21 19:49:03", sydney)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if name, _ := got.Zone(); name != "AEST" {
		t.Errorf("zone = %q, want AEST", name)
	}
}

func TestParseZonedIgnoresLocation(t *testing.T) {
	got, err := Parse("2026-09-21T09:49:03Z", sydney)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if !got.Equal(time.Date(2026, time.September, 21, 9, 49, 3, 0, time.UTC)) {
		t.Errorf("Parse() = %s, want the instant as written", got)
	}
}

func TestParseRejects(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{name: "empty", in: ""},
		{name: "whitespace", in: "   "},
		{name: "free text", in: "yesterday"},
		{name: "impossible date", in: "2026-13-45"},
		{name: "unix seconds", in: "1790004543"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.in, sydney)
			if err == nil {
				t.Fatalf("Parse(%q) error = nil, want an error", tt.in)
			}

			if strings.TrimSpace(tt.in) != "" && !strings.Contains(err.Error(), tt.in) {
				t.Errorf("error = %v, want it to name %q", err, tt.in)
			}
		})
	}
}

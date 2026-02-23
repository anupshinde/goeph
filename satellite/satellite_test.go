package satellite

import (
	"math"
	"testing"
	"time"

	"github.com/anupshinde/goeph/timescale"
)

// ISS TLE (representative, may be outdated — we just need valid propagation)
const (
	issName  = "ISS (ZARYA)"
	issLine1 = "1 25544U 98067A   24001.00000000  .00016717  00000-0  10270-3 0  9005"
	issLine2 = "2 25544  51.6400 208.9163 0006703 247.1970 112.8444 15.49560830999999"
)

func TestNewSat(t *testing.T) {
	sat := NewSat(issName, issLine1, issLine2)
	if sat.Name != issName {
		t.Errorf("name: got %q want %q", sat.Name, issName)
	}
}

func TestSubPoint(t *testing.T) {
	sat := NewSat(issName, issLine1, issLine2)
	// Propagate near the TLE epoch
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	lat, lon := SubPoint(sat.Sat, t0)

	// ISS orbit: inclination ~51.6°, so lat should be within [-52, 52]
	if lat < -52 || lat > 52 {
		t.Errorf("latitude out of ISS range: %f", lat)
	}

	// Longitude should be in [0, 360)
	if lon < 0 || lon >= 360 {
		t.Errorf("longitude out of range: %f", lon)
	}
}

func TestSubPoint_DifferentTimes(t *testing.T) {
	sat := NewSat(issName, issLine1, issLine2)
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := time.Date(2024, 1, 1, 0, 30, 0, 0, time.UTC)

	lat0, lon0 := SubPoint(sat.Sat, t0)
	lat1, lon1 := SubPoint(sat.Sat, t1)

	// 30 minutes later, ISS should be in a different position
	if lat0 == lat1 && lon0 == lon1 {
		t.Error("position unchanged after 30 minutes")
	}

	// Both should still be valid
	if math.IsNaN(lat0) || math.IsNaN(lon0) || math.IsNaN(lat1) || math.IsNaN(lon1) {
		t.Error("got NaN coordinates")
	}
}

// issEpochTT is the ISS TLE epoch (2024-01-01 00:00 UTC) as TT Julian date.
var issEpochTT = timescale.UTCToTT(timescale.TimeToJDUTC(
	time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)))

func TestFindEvents_Basic(t *testing.T) {
	sat := NewSat(issName, issLine1, issLine2)
	// NYC observer, 24-hour search near TLE epoch.
	lat, lon := 40.7128, -74.0060
	startJD := issEpochTT
	endJD := startJD + 1.0 // 1 day

	events, err := FindEvents(sat, lat, lon, startJD, endJD, 0.0)
	if err != nil {
		t.Fatal(err)
	}

	// ISS orbits ~15.5 times/day; not all passes visible from one location.
	// Expect at least a few passes (each with rise + culmination + set).
	if len(events) < 3 {
		t.Errorf("got %d events in 24h, want at least 3 (one pass)", len(events))
	}
	t.Logf("found %d events in 24 hours", len(events))

	// Verify events are in chronological order.
	for i := 1; i < len(events); i++ {
		if events[i].T < events[i-1].T {
			t.Errorf("events not sorted: event %d at %.6f before event %d at %.6f",
				i, events[i].T, i-1, events[i-1].T)
			break
		}
	}
}

func TestFindEvents_PassStructure(t *testing.T) {
	sat := NewSat(issName, issLine1, issLine2)
	lat, lon := 40.7128, -74.0060
	startJD := issEpochTT
	endJD := startJD + 1.0

	events, err := FindEvents(sat, lat, lon, startJD, endJD, 0.0)
	if err != nil {
		t.Fatal(err)
	}

	// Each complete pass should be Rise, Culmination, Set.
	i := 0
	passes := 0
	for i < len(events) {
		if events[i].Kind != Rise {
			t.Errorf("expected Rise at index %d, got kind=%d", i, events[i].Kind)
			break
		}
		if i+2 >= len(events) {
			break // incomplete pass at end
		}
		if events[i+1].Kind != Culmination {
			t.Errorf("expected Culmination at index %d, got kind=%d", i+1, events[i+1].Kind)
			break
		}
		if events[i+2].Kind != Set {
			t.Errorf("expected Set at index %d, got kind=%d", i+2, events[i+2].Kind)
			break
		}

		// Culmination altitude should be >= rise and set altitudes.
		if events[i+1].AltDeg < events[i].AltDeg {
			t.Errorf("pass %d: culmination alt %.2f < rise alt %.2f",
				passes, events[i+1].AltDeg, events[i].AltDeg)
		}

		// Rise time < Culmination time < Set time.
		if events[i].T >= events[i+1].T || events[i+1].T >= events[i+2].T {
			t.Errorf("pass %d: times not ordered: rise=%.6f, culm=%.6f, set=%.6f",
				passes, events[i].T, events[i+1].T, events[i+2].T)
		}

		passes++
		i += 3
	}
	t.Logf("verified %d complete passes", passes)
	if passes == 0 {
		t.Error("no complete passes found")
	}
}

func TestFindEvents_MinAltitude(t *testing.T) {
	sat := NewSat(issName, issLine1, issLine2)
	lat, lon := 40.7128, -74.0060
	startJD := issEpochTT
	endJD := startJD + 1.0

	// Find all passes (min alt = 0°).
	allEvents, err := FindEvents(sat, lat, lon, startJD, endJD, 0.0)
	if err != nil {
		t.Fatal(err)
	}

	// Find only high passes (min alt = 30°).
	highEvents, err := FindEvents(sat, lat, lon, startJD, endJD, 30.0)
	if err != nil {
		t.Fatal(err)
	}

	// Higher threshold should produce fewer or equal events.
	if len(highEvents) > len(allEvents) {
		t.Errorf("30° threshold gave %d events > %d events at 0°",
			len(highEvents), len(allEvents))
	}
	t.Logf("events at 0°: %d, at 30°: %d", len(allEvents), len(highEvents))
}

func TestFindEvents_CulminationAltitude(t *testing.T) {
	sat := NewSat(issName, issLine1, issLine2)
	lat, lon := 40.7128, -74.0060
	startJD := issEpochTT
	endJD := startJD + 2.0 // 2 days for more passes

	events, err := FindEvents(sat, lat, lon, startJD, endJD, 0.0)
	if err != nil {
		t.Fatal(err)
	}

	// Check that culmination altitudes are positive and reasonable.
	for i, e := range events {
		if e.Kind == Culmination {
			if e.AltDeg <= 0 {
				t.Errorf("event %d: culmination alt = %.2f°, should be positive", i, e.AltDeg)
			}
			if e.AltDeg > 90 {
				t.Errorf("event %d: culmination alt = %.2f°, should be <= 90", i, e.AltDeg)
			}
		}
	}
}

func TestFindEvents_ShortRange(t *testing.T) {
	sat := NewSat(issName, issLine1, issLine2)
	lat, lon := 40.7128, -74.0060
	// Very short range (1 hour) — may or may not have events.
	startJD := issEpochTT
	endJD := startJD + 1.0/24.0

	events, err := FindEvents(sat, lat, lon, startJD, endJD, 0.0)
	if err != nil {
		t.Fatal(err)
	}
	// Just verify no errors and events (if any) are ordered.
	for i := 1; i < len(events); i++ {
		if events[i].T < events[i-1].T {
			t.Errorf("events not sorted in short range")
			break
		}
	}
	t.Logf("found %d events in 1 hour", len(events))
}

func TestJdToCalendar(t *testing.T) {
	// J2000.0 = 2451545.0 = 2000-01-01 12:00:00 UTC
	y, mo, d, h, mi, s := jdToCalendar(2451545.0)
	if y != 2000 || mo != 1 || d != 1 || h != 12 || mi != 0 || s != 0 {
		t.Errorf("J2000: got %04d-%02d-%02d %02d:%02d:%02d, want 2000-01-01 12:00:00",
			y, mo, d, h, mi, s)
	}

	// J2000 + 0.5 days = 2000-01-02 00:00:00.
	y, mo, d, h, mi, s = jdToCalendar(2451545.5)
	if y != 2000 || mo != 1 || d != 2 || h != 0 || mi != 0 || s != 0 {
		t.Errorf("J2000+0.5: got %04d-%02d-%02d %02d:%02d:%02d, want 2000-01-02 00:00:00",
			y, mo, d, h, mi, s)
	}

	// 2024-06-15 18:30:00 UTC = JD 2460477.270833...
	y, mo, d, h, mi, s = jdToCalendar(2460477.0 + 6.5/24.0)
	if y != 2024 || mo != 6 || d != 15 || h != 18 || mi != 30 {
		t.Errorf("got %04d-%02d-%02d %02d:%02d:%02d, want 2024-06-15 18:30:00",
			y, mo, d, h, mi, s)
	}
}

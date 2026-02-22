package satellite

import (
	"math"
	"testing"
	"time"
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

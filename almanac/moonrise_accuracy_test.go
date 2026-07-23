package almanac

import (
	"math"
	"testing"
	"time"

	"github.com/anupshinde/goeph/timescale"
)

// Published moonrise/moonset values from the USNO Astronomical Applications
// API (aa.usno.navy.mil/api/rstt/oneday) for 28.6139°N, 77.2090°E (Delhi),
// UTC+5:30, used as external accuracy anchors. Tolerance 3 minutes: target
// accuracy is ~2 min plus the reference's rounding to the minute; observed
// agreement is well under a minute.
func TestMoonriseMoonset_PublishedValues(t *testing.T) {
	lat, lon := 28.6139, 77.2090
	const tolMin = 3.0

	utc := func(y int, mo time.Month, d, h, mi int) float64 {
		return timescale.UTCToTT(timescale.TimeToJDUTC(time.Date(y, mo, d, h, mi, 0, 0, time.UTC)))
	}

	cases := []struct {
		desc    string
		wantJD  float64
		isRise  bool
		winFrom float64
	}{
		// USNO 2026-07-15: moonrise 06:15, moonset 20:23 (UTC+5:30).
		{"2026-07-15 moonrise 06:15+05:30", utc(2026, 7, 15, 0, 45), true, utc(2026, 7, 14, 12, 0)},
		{"2026-07-15 moonset 20:23+05:30", utc(2026, 7, 15, 14, 53), false, utc(2026, 7, 14, 12, 0)},
		// USNO 2026-07-24: moonset 00:51, moonrise 15:20 (UTC+5:30).
		{"2026-07-24 moonset 00:51+05:30", utc(2026, 7, 23, 19, 21), false, utc(2026, 7, 23, 12, 0)},
		{"2026-07-24 moonrise 15:20+05:30", utc(2026, 7, 24, 9, 50), true, utc(2026, 7, 23, 12, 0)},
	}

	for _, tc := range cases {
		events, err := MoonriseMoonset(testEph, lat, lon, tc.winFrom, tc.winFrom+1.2)
		if err != nil {
			t.Fatal(err)
		}
		wantValue := 0
		if tc.isRise {
			wantValue = 1
		}
		best := math.Inf(1)
		for _, e := range events {
			if e.NewValue != wantValue {
				continue
			}
			if d := math.Abs(e.T - tc.wantJD); d < best {
				best = d
			}
		}
		bestMin := best * 24 * 60
		t.Logf("%s: nearest event %.2f min from published", tc.desc, bestMin)
		if bestMin > tolMin {
			t.Errorf("%s: nearest event %.2f min from published value (tol %.0f min)", tc.desc, bestMin, tolMin)
		}
	}
}

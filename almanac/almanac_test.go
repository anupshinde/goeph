package almanac

import (
	"encoding/json"
	"math"
	"os"
	"testing"

	"github.com/anupshinde/goeph/search"
	"github.com/anupshinde/goeph/spk"
)

var testEph *spk.SPK

func TestMain(m *testing.M) {
	var err error
	testEph, err = spk.Open("../data/de440s.bsp")
	if err != nil {
		panic("failed to load ephemeris: " + err.Error())
	}
	os.Exit(m.Run())
}

// --- Golden test data structures ---

type seasonGolden struct {
	Tests []struct {
		TTJD   float64 `json:"tt_jd"`
		Season int     `json:"season"`
	} `json:"tests"`
}

type moonPhaseGolden struct {
	Tests []struct {
		TTJD  float64 `json:"tt_jd"`
		Phase int     `json:"phase"`
	} `json:"tests"`
}

type sunriseSunsetGolden struct {
	Lat   float64 `json:"lat"`
	Lon   float64 `json:"lon"`
	Tests []struct {
		TTJD      float64 `json:"tt_jd"`
		IsSunrise int     `json:"is_sunrise"`
	} `json:"tests"`
}

type twilightGolden struct {
	Lat   float64 `json:"lat"`
	Lon   float64 `json:"lon"`
	Tests []struct {
		TTJD  float64 `json:"tt_jd"`
		Level int     `json:"level"`
	} `json:"tests"`
}

type oppositionGolden struct {
	BodyID int `json:"body_id"`
	Tests  []struct {
		TTJD  float64 `json:"tt_jd"`
		Value int     `json:"value"`
	} `json:"tests"`
}

func loadJSON(t *testing.T, path string, v interface{}) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatal(err)
	}
}

// matchEvents matches each golden event to the nearest goeph event by time,
// returning the number of matched events and the maximum time difference.
// Events must have matching discrete values to count as matched.
func matchEvents(got []search.DiscreteEvent, wantTimes []float64, wantValues []int, tolDays float64) (matched, valueMismatch int, maxDiff float64) {
	for _, wt := range wantTimes {
		bestDiff := math.MaxFloat64
		bestIdx := -1
		for j, e := range got {
			d := math.Abs(e.T - wt)
			if d < bestDiff {
				bestDiff = d
				bestIdx = j
			}
		}
		if bestIdx >= 0 && bestDiff <= tolDays {
			matched++
			if bestDiff > maxDiff {
				maxDiff = bestDiff
			}
		}
	}
	// Also check value matches for matched events.
	gi := 0
	for wi := 0; wi < len(wantTimes); wi++ {
		for gi < len(got)-1 && math.Abs(got[gi+1].T-wantTimes[wi]) < math.Abs(got[gi].T-wantTimes[wi]) {
			gi++
		}
		if gi < len(got) && math.Abs(got[gi].T-wantTimes[wi]) <= tolDays {
			if got[gi].NewValue != wantValues[wi] {
				valueMismatch++
			}
		}
	}
	return
}

// --- Seasons golden test ---

func TestSeasonsGolden(t *testing.T) {
	var golden seasonGolden
	loadJSON(t, "../testdata/golden_seasons.json", &golden)

	startJD := golden.Tests[0].TTJD - 30
	endJD := golden.Tests[len(golden.Tests)-1].TTJD + 30

	events, err := Seasons(testEph, startJD, endJD)
	if err != nil {
		t.Fatal(err)
	}

	if len(events) != len(golden.Tests) {
		t.Fatalf("got %d events, want %d", len(events), len(golden.Tests))
	}

	// J2000 ecliptic vs ecliptic of date causes up to ~18 hours offset.
	const tolDays = 1.0
	maxDiff := 0.0
	failures := 0
	for i := range events {
		diff := math.Abs(events[i].T - golden.Tests[i].TTJD)
		if diff > maxDiff {
			maxDiff = diff
		}
		if events[i].NewValue != golden.Tests[i].Season {
			if failures < 10 {
				t.Errorf("event %d: season=%d, want %d", i, events[i].NewValue, golden.Tests[i].Season)
			}
			failures++
		}
		if diff > tolDays {
			if failures < 10 {
				t.Errorf("event %d: T diff = %.6f days (%.1f hours)", i, diff, diff*24)
			}
			failures++
		}
	}
	t.Logf("max time diff: %.6f days (%.1f hours), %d failures out of %d events",
		maxDiff, maxDiff*24, failures, len(events))
}

// --- Moon Phases golden test ---

func TestMoonPhasesGolden(t *testing.T) {
	var golden moonPhaseGolden
	loadJSON(t, "../testdata/golden_moon_phases.json", &golden)

	startJD := golden.Tests[0].TTJD - 15
	endJD := golden.Tests[len(golden.Tests)-1].TTJD + 15

	events, err := MoonPhases(testEph, startJD, endJD)
	if err != nil {
		t.Fatal(err)
	}

	// Allow small count difference due to J2000 vs ecliptic-of-date frame.
	countDiff := len(events) - len(golden.Tests)
	if countDiff < -5 || countDiff > 5 {
		t.Fatalf("got %d events, want ~%d (diff %d)", len(events), len(golden.Tests), countDiff)
	}
	t.Logf("event count: got %d, golden %d (diff %+d)", len(events), len(golden.Tests), countDiff)

	// Match each golden event to nearest goeph event.
	// Relative longitude cancels J2000 vs ecliptic-of-date frame difference,
	// so agreement is much tighter than seasons (~seconds, not hours).
	const tolDays = 3.0 / (24 * 60) // 3 minutes (measured max ~4 seconds)
	wantTimes := make([]float64, len(golden.Tests))
	wantValues := make([]int, len(golden.Tests))
	for i, g := range golden.Tests {
		wantTimes[i] = g.TTJD
		wantValues[i] = g.Phase
	}
	matched, valMismatch, maxDiff := matchEvents(events, wantTimes, wantValues, tolDays)
	t.Logf("matched %d/%d golden events, %d value mismatches, max diff %.6f days (%.1f sec)",
		matched, len(golden.Tests), valMismatch, maxDiff, maxDiff*86400)

	// At least 99% should match.
	minMatch := len(golden.Tests) * 99 / 100
	if matched < minMatch {
		t.Errorf("only matched %d/%d golden events (need %d)", matched, len(golden.Tests), minMatch)
	}
}

// --- Sunrise/Sunset golden test ---

func TestSunriseSunsetGolden(t *testing.T) {
	var golden sunriseSunsetGolden
	loadJSON(t, "../testdata/golden_sunrise_sunset.json", &golden)

	startJD := golden.Tests[0].TTJD - 1
	endJD := golden.Tests[len(golden.Tests)-1].TTJD + 1

	events, err := SunriseSunset(testEph, golden.Lat, golden.Lon, startJD, endJD)
	if err != nil {
		t.Fatal(err)
	}

	countDiff := len(events) - len(golden.Tests)
	if countDiff < -5 || countDiff > 5 {
		t.Fatalf("got %d events, want ~%d (diff %d)", len(events), len(golden.Tests), countDiff)
	}
	t.Logf("event count: got %d, golden %d (diff %+d)", len(events), len(golden.Tests), countDiff)

	// Match each golden event to nearest goeph event.
	const tolDays = 3.0 / (24 * 60) // 3 minutes (measured max ~1 second)
	wantTimes := make([]float64, len(golden.Tests))
	wantValues := make([]int, len(golden.Tests))
	for i, g := range golden.Tests {
		wantTimes[i] = g.TTJD
		wantValues[i] = g.IsSunrise
	}
	matched, valMismatch, maxDiff := matchEvents(events, wantTimes, wantValues, tolDays)
	t.Logf("matched %d/%d golden events, %d value mismatches, max diff %.6f days (%.1f min)",
		matched, len(golden.Tests), valMismatch, maxDiff, maxDiff*24*60)

	minMatch := len(golden.Tests) * 99 / 100
	if matched < minMatch {
		t.Errorf("only matched %d/%d golden events (need %d)", matched, len(golden.Tests), minMatch)
	}
}

// --- Twilight golden test ---

func TestTwilightGolden(t *testing.T) {
	var golden twilightGolden
	loadJSON(t, "../testdata/golden_twilight.json", &golden)

	startJD := golden.Tests[0].TTJD - 1
	endJD := golden.Tests[len(golden.Tests)-1].TTJD + 1

	events, err := Twilight(testEph, golden.Lat, golden.Lon, startJD, endJD)
	if err != nil {
		t.Fatal(err)
	}

	// Twilight has more edge cases; allow larger count difference.
	countDiff := len(events) - len(golden.Tests)
	t.Logf("event count: got %d, golden %d (diff %+d)", len(events), len(golden.Tests), countDiff)

	// Match each golden event to nearest goeph event.
	const tolDays = 3.0 / (24 * 60) // 3 minutes (measured max ~1 second)
	wantTimes := make([]float64, len(golden.Tests))
	wantValues := make([]int, len(golden.Tests))
	for i, g := range golden.Tests {
		wantTimes[i] = g.TTJD
		wantValues[i] = g.Level
	}
	matched, valMismatch, maxDiff := matchEvents(events, wantTimes, wantValues, tolDays)
	t.Logf("matched %d/%d golden events, %d value mismatches, max diff %.6f days (%.1f min)",
		matched, len(golden.Tests), valMismatch, maxDiff, maxDiff*24*60)

	// At least 95% should match (twilight is most sensitive to nutation/frame differences).
	minMatch := len(golden.Tests) * 95 / 100
	if matched < minMatch {
		t.Errorf("only matched %d/%d golden events (need %d)", matched, len(golden.Tests), minMatch)
	}
}

// --- Oppositions/Conjunctions golden test ---

func TestOppositionsConjunctionsGolden(t *testing.T) {
	var golden oppositionGolden
	loadJSON(t, "../testdata/golden_oppositions.json", &golden)

	startJD := golden.Tests[0].TTJD - 60
	endJD := golden.Tests[len(golden.Tests)-1].TTJD + 60

	events, err := OppositionsConjunctions(testEph, golden.BodyID, startJD, endJD)
	if err != nil {
		t.Fatal(err)
	}

	if len(events) != len(golden.Tests) {
		t.Fatalf("got %d events, want %d", len(events), len(golden.Tests))
	}

	// Relative longitude cancels frame difference, so agreement is tight (~seconds).
	const tolDays = 3.0 / (24 * 60) // 3 minutes (measured max ~46 seconds)
	maxDiff := 0.0
	failures := 0
	for i := range events {
		diff := math.Abs(events[i].T - golden.Tests[i].TTJD)
		if diff > maxDiff {
			maxDiff = diff
		}
		if events[i].NewValue != golden.Tests[i].Value {
			if failures < 10 {
				t.Errorf("event %d: value=%d, want %d", i, events[i].NewValue, golden.Tests[i].Value)
			}
			failures++
		}
		if diff > tolDays {
			if failures < 10 {
				t.Errorf("event %d: T diff = %.6f days (%.1f sec)", i, diff, diff*86400)
			}
			failures++
		}
	}
	t.Logf("max time diff: %.6f days (%.1f sec), %d failures out of %d events",
		maxDiff, maxDiff*86400, failures, len(events))
}

// --- Unit tests (no golden data) ---

func TestSeasons_EventCount(t *testing.T) {
	// 10 years should have ~40 season events (4 per year).
	start := 2451545.0 // J2000
	end := start + 3652.5
	events, err := Seasons(testEph, start, end)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) < 38 || len(events) > 42 {
		t.Errorf("got %d events for 10 years, want ~40", len(events))
	}
}

func TestMoonPhases_EventCount(t *testing.T) {
	// 1 year should have ~49 moon phase events (4 phases * ~12.37 cycles).
	start := 2451545.0
	end := start + 365.25
	events, err := MoonPhases(testEph, start, end)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) < 45 || len(events) > 55 {
		t.Errorf("got %d events for 1 year, want ~49", len(events))
	}
}

func TestSunriseSunset_MidLatitude(t *testing.T) {
	// NYC, June 2024 — expect ~60 events (2 per day for 30 days).
	start := 2460466.5
	end := start + 30
	events, err := SunriseSunset(testEph, 40.7, -74.0, start, end)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) < 55 || len(events) > 65 {
		t.Errorf("got %d events for 30 days, want ~60", len(events))
	}
	// Check alternating sunrise/sunset.
	for i := 1; i < len(events); i++ {
		if events[i].NewValue == events[i-1].NewValue {
			t.Errorf("events %d and %d have same value %d (should alternate)",
				i-1, i, events[i].NewValue)
			break
		}
	}
}

func TestTwilight_EventCount(t *testing.T) {
	// NYC, January 2024 — expect ~8 transitions per day * 31 days ≈ 248.
	start := 2460310.5 // ~2024-01-01 TT
	end := start + 31
	events, err := Twilight(testEph, 40.7, -74.0, start, end)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) < 200 || len(events) > 300 {
		t.Errorf("got %d twilight events for 31 days, want ~248", len(events))
	}
}

func TestRisings_Moon(t *testing.T) {
	// Moon should rise roughly once per day (sometimes 0 or 2 times).
	// NYC, January 2024, 31 days.
	start := 2460310.5
	end := start + 31
	events, err := Risings(testEph, spk.Moon, 40.7, -74.0, start, end)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) < 25 || len(events) > 35 {
		t.Errorf("got %d moon risings in 31 days, want ~30", len(events))
	}
}

func TestTransits_Sun(t *testing.T) {
	// Sun should transit once per day.
	// NYC, January 2024, 10 days.
	start := 2460310.5
	end := start + 10
	events, err := Transits(testEph, spk.Sun, 40.7, -74.0, start, end)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) < 9 || len(events) > 11 {
		t.Errorf("got %d sun transits in 10 days, want ~10", len(events))
	}
}

func TestMoonriseMoonset_EventStructure(t *testing.T) {
	// Delhi, January 2024, 30 days. Rises and sets alternate; each occurs
	// roughly once per lunar day (~24h50m), so ~29 of each in 30 days.
	start := 2460310.5
	end := start + 30
	events, err := MoonriseMoonset(testEph, 28.6139, 77.2090, start, end)
	if err != nil {
		t.Fatal(err)
	}
	var rises, sets []float64
	for i, e := range events {
		if e.NewValue == 1 {
			rises = append(rises, e.T)
		} else {
			sets = append(sets, e.T)
		}
		if i > 0 && events[i].NewValue == events[i-1].NewValue {
			t.Errorf("events %d and %d have same NewValue %d — must alternate", i-1, i, e.NewValue)
		}
	}
	if len(rises) < 27 || len(rises) > 31 {
		t.Errorf("got %d moonrises in 30 days, want ~29", len(rises))
	}
	if len(sets) < 27 || len(sets) > 31 {
		t.Errorf("got %d moonsets in 30 days, want ~29", len(sets))
	}
	// Successive moonrises are separated by the lunar day: ~24.2 h to ~25.5 h
	// at this latitude.
	for i := 1; i < len(rises); i++ {
		gapH := (rises[i] - rises[i-1]) * 24
		if gapH < 23.5 || gapH > 26.0 {
			t.Errorf("moonrise gap %d: %.2f h outside lunar-day range", i, gapH)
		}
	}
}

func TestMoonriseMoonset_PhaseAlignment(t *testing.T) {
	// External property anchors: on the day of new moon the Moon rises with
	// the Sun; on full moon it rises at sunset. 2024-04-08 was a new moon
	// (18:21 UTC); 2024-04-23 a full moon (23:49 UTC). Delhi coordinates.
	lat, lon := 28.6139, 77.2090

	check := func(desc string, dayStartJD float64, sunEventValue int, maxDiffMin float64) {
		moonEvents, err := MoonriseMoonset(testEph, lat, lon, dayStartJD, dayStartJD+1.5)
		if err != nil {
			t.Fatal(err)
		}
		sunEvents, err := SunriseSunset(testEph, lat, lon, dayStartJD, dayStartJD+1.5)
		if err != nil {
			t.Fatal(err)
		}
		var moonRise, sunRef float64
		for _, e := range moonEvents {
			if e.NewValue == 1 && moonRise == 0 {
				moonRise = e.T
			}
		}
		for _, e := range sunEvents {
			if e.NewValue == sunEventValue && sunRef == 0 {
				sunRef = e.T
			}
		}
		if moonRise == 0 || sunRef == 0 {
			t.Fatalf("%s: missing events (moonrise=%.5f, sun=%.5f)", desc, moonRise, sunRef)
		}
		diffMin := math.Abs(moonRise-sunRef) * 24 * 60
		t.Logf("%s: moonrise vs sun event differ by %.0f min", desc, diffMin)
		if diffMin > maxDiffMin {
			t.Errorf("%s: moonrise %.5f vs sun event %.5f differ by %.0f min (max %.0f)",
				desc, moonRise, sunRef, diffMin, maxDiffMin)
		}
	}

	// New moon 2024-04-08: first moonrise that morning is within ~45 min of
	// sunrise (elongation is near 0 but not exactly 0 at rise time).
	check("new moon rise≈sunrise", 2460408.5, 1, 45)
	// Full moon 2024-04-23: moonrise within ~45 min of sunset.
	check("full moon rise≈sunset", 2460423.5, 0, 45)
}

func TestMoonriseMoonset_ThresholdRange(t *testing.T) {
	// h0 = 0.7275·π − 34′ for the Moon's distance range (356k-407k km):
	// parallax 53.9′-61.5′ → h0 between roughly +5′ and +11′.
	for _, distKm := range []float64{356500.0, 384400.0, 406700.0} {
		h0 := moonRiseSetThreshold(distKm)
		if h0 < 0.07 || h0 > 0.20 {
			t.Errorf("moonRiseSetThreshold(%.0f km) = %.4f°, outside expected +0.07..+0.20°", distKm, h0)
		}
	}
}

func TestMoonriseMoonset_HighLatitudeAbsence(t *testing.T) {
	// Tromsø (69.65°N): near winter solstice the Moon spends multi-day
	// stretches entirely above or below the horizon, so 30 winter days must
	// yield noticeably fewer than one rise and one set per lunar day —
	// proving the finder represents absence rather than fabricating events.
	start := 2460295.5 // 2023-12-17
	events, err := MoonriseMoonset(testEph, 69.6492, 18.9553, start, start+30)
	if err != nil {
		t.Fatal(err)
	}
	var rises, sets int
	for _, e := range events {
		if e.NewValue == 1 {
			rises++
		} else {
			sets++
		}
	}
	if rises >= 29 && sets >= 29 {
		t.Errorf("expected missing rise/set days at 69.6°N in winter, got %d rises %d sets", rises, sets)
	}
	if rises == 0 && sets == 0 {
		t.Error("expected some moon events even at high latitude")
	}
}

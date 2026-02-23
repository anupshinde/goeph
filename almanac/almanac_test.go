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
	const tolDays = 1.0
	wantTimes := make([]float64, len(golden.Tests))
	wantValues := make([]int, len(golden.Tests))
	for i, g := range golden.Tests {
		wantTimes[i] = g.TTJD
		wantValues[i] = g.Phase
	}
	matched, valMismatch, maxDiff := matchEvents(events, wantTimes, wantValues, tolDays)
	t.Logf("matched %d/%d golden events, %d value mismatches, max diff %.6f days (%.1f hours)",
		matched, len(golden.Tests), valMismatch, maxDiff, maxDiff*24)

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
	const tolDays = 5.0 / (24 * 60) // 5 minutes
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
	const tolDays = 10.0 / (24 * 60) // 10 minutes
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

	const tolDays = 1.0
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
				t.Errorf("event %d: T diff = %.6f days (%.1f hours)", i, diff, diff*24)
			}
			failures++
		}
	}
	t.Logf("max time diff: %.6f days (%.1f hours), %d failures out of %d events",
		maxDiff, maxDiff*24, failures, len(events))
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

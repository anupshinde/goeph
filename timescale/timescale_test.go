package timescale

import (
	"encoding/json"
	"math"
	"os"
	"testing"
	"time"
)

type goldenTimescale struct {
	Tests []struct {
		UTCJD float64 `json:"utc_jd"`
		TTJD  float64 `json:"tt_jd"`
		UT1JD float64 `json:"ut1_jd"`
	} `json:"tests"`
}

func loadGolden(t *testing.T) goldenTimescale {
	t.Helper()
	data, err := os.ReadFile("../testdata/golden_timescale.json")
	if err != nil {
		t.Fatal(err)
	}
	var g goldenTimescale
	if err := json.Unmarshal(data, &g); err != nil {
		t.Fatal(err)
	}
	return g
}

func TestLeapSecondOffset(t *testing.T) {
	tests := []struct {
		jdUTC float64
		want  float64
	}{
		{2441317.5, 10},  // 1972-01-01 exactly
		{2441318.0, 10},  // just after
		{2441499.5, 11},  // 1972-07-01
		{2457754.5, 37},  // 2017-01-01 (latest)
		{2460000.0, 37},  // future: should return latest
		{2400000.0, 10},  // pre-1972: returns initial 10
	}
	for _, tc := range tests {
		got := LeapSecondOffset(tc.jdUTC)
		if got != tc.want {
			t.Errorf("LeapSecondOffset(%.1f) = %f, want %f", tc.jdUTC, got, tc.want)
		}
	}
}

func TestDeltaT_KnownValues(t *testing.T) {
	dt := DeltaT(2000.0)
	if math.Abs(dt-63.829) > 0.001 {
		t.Errorf("DeltaT(2000) = %f, want ~63.829", dt)
	}

	dt = DeltaT(2000.5)
	dt2000 := DeltaT(2000.0)
	dt2001 := DeltaT(2001.0)
	if dt < math.Min(dt2000, dt2001) || dt > math.Max(dt2000, dt2001) {
		t.Errorf("DeltaT(2000.5) = %f, not between %f and %f", dt, dt2000, dt2001)
	}
}

func TestDeltaT_BoundaryClamp(t *testing.T) {
	dt := DeltaT(1700.0)
	dtFirst := DeltaT(1800.0)
	if dt != dtFirst {
		t.Errorf("DeltaT(1700) = %f, want %f (first entry)", dt, dtFirst)
	}

	dt = DeltaT(2300.0)
	dtLast := DeltaT(2200.0)
	if dt != dtLast {
		t.Errorf("DeltaT(2300) = %f, want %f (last entry)", dt, dtLast)
	}
}

func TestDeltaT_LastInterval(t *testing.T) {
	dt := DeltaT(2199.5)
	dt2199 := DeltaT(2199.0)
	dt2200 := DeltaT(2200.0)
	if dt < math.Min(dt2199, dt2200) || dt > math.Max(dt2199, dt2200) {
		t.Errorf("DeltaT(2199.5) = %f, not between %f and %f", dt, dt2199, dt2200)
	}
}

func TestDeltaT_ExactTableEntry(t *testing.T) {
	dt := DeltaT(1800.0)
	if math.Abs(dt-18.3670) > 0.0001 {
		t.Errorf("DeltaT(1800) = %f, want 18.3670", dt)
	}
}

func TestDeltaT_NearEnd(t *testing.T) {
	// Year 2199.999 — exercises the idx >= n-1 guard near end of table
	dt := DeltaT(2199.999)
	dt2199 := DeltaT(2199.0)
	dt2200 := DeltaT(2200.0)
	if dt < math.Min(dt2199, dt2200) || dt > math.Max(dt2199, dt2200) {
		t.Errorf("DeltaT(2199.999) = %f, not between %f and %f", dt, dt2199, dt2200)
	}
}

func TestTimeToJDUTC(t *testing.T) {
	j2000 := time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)
	jd := TimeToJDUTC(j2000)
	if math.Abs(jd-2451545.0) > 1e-10 {
		t.Errorf("J2000 JD = %.10f, want 2451545.0", jd)
	}

	unix0 := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	jd = TimeToJDUTC(unix0)
	if math.Abs(jd-2440587.5) > 1e-10 {
		t.Errorf("Unix epoch JD = %.10f, want 2440587.5", jd)
	}
}

func TestTimeToJDUTC_Nanoseconds(t *testing.T) {
	t0 := time.Date(2024, 6, 15, 12, 0, 0, 500000000, time.UTC)
	t1 := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	jd0 := TimeToJDUTC(t0)
	jd1 := TimeToJDUTC(t1)
	diffSec := (jd0 - jd1) * SecPerDay
	if math.Abs(diffSec-0.5) > 1e-3 {
		t.Errorf("nanosecond diff: got %.9f s, want 0.5 s", diffSec)
	}
}

func TestUTCToTT(t *testing.T) {
	jdUTC := 2458849.5
	jdTT := UTCToTT(jdUTC)
	expectedOffset := (37.0 + 32.184) / SecPerDay
	diff := jdTT - jdUTC - expectedOffset
	if math.Abs(diff) > 1e-9 {
		t.Errorf("UTCToTT offset error: %.15e days", diff)
	}
}

func TestTTToUT1(t *testing.T) {
	jdTT := 2451545.0
	jdUT1 := TTToUT1(jdTT)
	year := 2000.0 + (jdTT-2451545.0)/365.25
	dt := DeltaT(year)
	expected := jdTT - dt/SecPerDay
	if math.Abs(jdUT1-expected) > 1e-15 {
		t.Errorf("TTToUT1: got %.15f want %.15f", jdUT1, expected)
	}
}

func TestUTCToTT_Golden(t *testing.T) {
	golden := loadGolden(t)

	const tol = 1e-9
	failures := 0
	for i, tc := range golden.Tests {
		got := UTCToTT(tc.UTCJD)
		diff := math.Abs(got - tc.TTJD)
		if diff > tol {
			if failures < 10 {
				t.Errorf("test %d: UTC=%.10f got TT=%.10f want=%.10f diff=%.2e days",
					i, tc.UTCJD, got, tc.TTJD, diff)
			}
			failures++
		}
	}
	if failures > 0 {
		t.Errorf("%d UTCToTT failures out of %d tests", failures, len(golden.Tests))
	}
}

func TestTTToUT1_Golden(t *testing.T) {
	golden := loadGolden(t)

	const tol = 2e-6 // days (~170 ms; delta-T linear interpolation vs Skyfield's spline)
	failures := 0
	for i, tc := range golden.Tests {
		got := TTToUT1(tc.TTJD)
		diff := math.Abs(got - tc.UT1JD)
		if diff > tol {
			if failures < 10 {
				t.Errorf("test %d: TT=%.10f got UT1=%.10f want=%.10f diff=%.2e days",
					i, tc.TTJD, got, tc.UT1JD, diff)
			}
			failures++
		}
	}
	if failures > 0 {
		t.Errorf("%d TTToUT1 failures out of %d tests", failures, len(golden.Tests))
	}
}

// goldenTDBTT matches testdata/golden_tdbtt.json.
type goldenTDBTT struct {
	Tests []struct {
		TTJD          float64 `json:"tt_jd"`
		TDBMinusTTSec float64 `json:"tdb_minus_tt_sec"`
	} `json:"tests"`
}

func TestTDBMinusTT_Amplitude(t *testing.T) {
	// TDB-TT should never exceed ~2ms
	for year := 1850.0; year <= 2150.0; year += 1.0 {
		jd := 2451545.0 + (year-2000.0)*365.25
		dt := TDBMinusTT(jd)
		if math.Abs(dt) > 0.002 {
			t.Errorf("TDB-TT at year %.0f = %f s, exceeds 2ms", year, dt)
		}
	}
}

func TestTDBMinusTT_VariesWithTime(t *testing.T) {
	dt1 := TDBMinusTT(2451545.0)
	dt2 := TDBMinusTT(2451545.0 + 182.625) // half year later
	if dt1 == dt2 {
		t.Error("TDB-TT unchanged after half year")
	}
}

func TestTDBMinusTT_Golden(t *testing.T) {
	data, err := os.ReadFile("../testdata/golden_tdbtt.json")
	if err != nil {
		t.Fatal(err)
	}
	var golden goldenTDBTT
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatal(err)
	}

	const tol = 1e-9 // seconds; should match exactly (identical formula)
	failures := 0
	for i, tc := range golden.Tests {
		got := TDBMinusTT(tc.TTJD)
		diff := math.Abs(got - tc.TDBMinusTTSec)
		if diff > tol {
			if failures < 10 {
				t.Errorf("test %d: tt=%.6f got=%.15f want=%.15f diff=%.2e s",
					i, tc.TTJD, got, tc.TDBMinusTTSec, diff)
			}
			failures++
		}
	}
	if failures > 0 {
		t.Errorf("%d TDB-TT failures out of %d tests (tol=%.0e s)", failures, len(golden.Tests), tol)
	}
}

func BenchmarkTDBMinusTT(b *testing.B) {
	for i := 0; i < b.N; i++ {
		TDBMinusTT(2451545.0 + float64(i))
	}
}

func BenchmarkUTCToTT(b *testing.B) {
	for i := 0; i < b.N; i++ {
		UTCToTT(2451545.0)
	}
}

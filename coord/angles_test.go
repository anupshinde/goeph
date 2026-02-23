package coord

import (
	"math"
	"testing"
)

func TestSeparationAngle_ZeroVectors(t *testing.T) {
	sep := SeparationAngle([3]float64{0, 0, 0}, [3]float64{1, 0, 0})
	if sep != 0 {
		t.Errorf("zero vector: got %f, want 0", sep)
	}
}

func TestSeparationAngle_Parallel(t *testing.T) {
	sep := SeparationAngle([3]float64{1, 0, 0}, [3]float64{2, 0, 0})
	if math.Abs(sep) > 1e-12 {
		t.Errorf("parallel: got %f, want 0", sep)
	}
}

func TestSeparationAngle_Perpendicular(t *testing.T) {
	sep := SeparationAngle([3]float64{1, 0, 0}, [3]float64{0, 1, 0})
	if math.Abs(sep-90.0) > 1e-12 {
		t.Errorf("perpendicular: got %f, want 90", sep)
	}
}

func TestSeparationAngle_Antiparallel(t *testing.T) {
	sep := SeparationAngle([3]float64{1, 0, 0}, [3]float64{-1, 0, 0})
	if math.Abs(sep-180.0) > 1e-12 {
		t.Errorf("antiparallel: got %f, want 180", sep)
	}
}

func TestSeparationAngle_SmallAngle(t *testing.T) {
	// Nearly parallel vectors — tests numerical stability
	a := [3]float64{1, 0, 0}
	b := [3]float64{1, 1e-10, 0}
	sep := SeparationAngle(a, b)
	expected := math.Atan2(1e-10, 1) * rad2deg
	if math.Abs(sep-expected) > 1e-8 {
		t.Errorf("small angle: got %.15e, want %.15e", sep, expected)
	}
}

// goldenSeparation matches testdata/golden_separation.json.
type goldenSeparation struct {
	Tests []struct {
		TDBJD  float64 `json:"tdb_jd"`
		Body1  string  `json:"body1"`
		Body2  string  `json:"body2"`
		SepDeg float64 `json:"separation_deg"`
	} `json:"tests"`
}

func TestSeparationAngle_Golden(t *testing.T) {
	// This test validates separation angle via the Sun-Moon separation
	// computed from their astrometric position vectors.
	// We load SPK positions and compute separation ourselves.
	var golden goldenSeparation
	loadJSON(t, "../testdata/golden_separation.json", &golden)

	// Load SPK position data to get the actual vectors
	type spkEntry struct {
		TDBJD  float64    `json:"tdb_jd"`
		BodyID int        `json:"body_id"`
		PosKm  [3]float64 `json:"pos_km"`
	}
	var spkData struct {
		Tests []spkEntry `json:"tests"`
	}
	loadJSON(t, "../testdata/golden_spk.json", &spkData)

	// Index SPK data by (tdb_jd, body_id) for lookup
	type key struct {
		tdb    float64
		bodyID int
	}
	posMap := make(map[key][3]float64)
	for _, e := range spkData.Tests {
		posMap[key{e.TDBJD, e.BodyID}] = e.PosKm
	}

	const tol = 1e-8 // degrees; should match nearly exactly since same positions
	failures := 0
	for i, tc := range golden.Tests {
		sunPos, ok1 := posMap[key{tc.TDBJD, 10}]
		moonPos, ok2 := posMap[key{tc.TDBJD, 301}]
		if !ok1 || !ok2 {
			continue
		}
		got := SeparationAngle(sunPos, moonPos)
		diff := math.Abs(got - tc.SepDeg)
		if diff > tol {
			if failures < 10 {
				t.Errorf("test %d: tdb=%.6f got=%.10f want=%.10f diff=%.2e",
					i, tc.TDBJD, got, tc.SepDeg, diff)
			}
			failures++
		}
	}
	if failures > 0 {
		t.Errorf("%d separation failures out of %d tests (tol=%.0e°)", failures, len(golden.Tests), tol)
	}
}

func TestPhaseAngle_FullyLit(t *testing.T) {
	// Observer and sun on same side of target
	obsToTarget := [3]float64{1, 0, 0}
	sunToTarget := [3]float64{1, 0, 0}
	pa := PhaseAngle(obsToTarget, sunToTarget)
	if math.Abs(pa) > 1e-12 {
		t.Errorf("fully lit: got %f, want 0", pa)
	}
}

func TestPhaseAngle_HalfLit(t *testing.T) {
	obsToTarget := [3]float64{1, 0, 0}
	sunToTarget := [3]float64{0, 1, 0}
	pa := PhaseAngle(obsToTarget, sunToTarget)
	if math.Abs(pa-90) > 1e-12 {
		t.Errorf("half lit: got %f, want 90", pa)
	}
}

func TestFractionIlluminated_Values(t *testing.T) {
	tests := []struct {
		phase float64
		want  float64
	}{
		{0, 1.0},
		{90, 0.5},
		{180, 0.0},
		{60, 0.75},
	}
	for _, tc := range tests {
		got := FractionIlluminated(tc.phase)
		if math.Abs(got-tc.want) > 1e-12 {
			t.Errorf("FractionIlluminated(%f) = %f, want %f", tc.phase, got, tc.want)
		}
	}
}

// goldenPhase matches testdata/golden_phase.json.
type goldenPhase struct {
	Tests []struct {
		TDBJD              float64    `json:"tdb_jd"`
		BodyName           string     `json:"body_name"`
		PhaseAngleDeg      float64    `json:"phase_angle_deg"`
		FractionIlluminate float64    `json:"fraction_illuminated"`
		ObsToTargetKm      [3]float64 `json:"obs_to_target_km"`
		SunToTargetKm      [3]float64 `json:"sun_to_target_km"`
	} `json:"tests"`
}

func TestPhaseAngle_Golden(t *testing.T) {
	var golden goldenPhase
	loadJSON(t, "../testdata/golden_phase.json", &golden)

	// With exact input vectors from Skyfield, tolerance is limited only by
	// floating-point arithmetic in the angle formula itself (measured max ~6e-14°).
	const tol = 1e-10 // degrees
	failures := 0
	maxDiff := 0.0

	for _, tc := range golden.Tests {
		got := PhaseAngle(tc.ObsToTargetKm, tc.SunToTargetKm)
		diff := math.Abs(got - tc.PhaseAngleDeg)
		if diff > maxDiff {
			maxDiff = diff
		}
		if diff > tol {
			if failures < 10 {
				t.Errorf("%s tdb=%.6f: got=%.6f want=%.6f diff=%.10f°",
					tc.BodyName, tc.TDBJD, got, tc.PhaseAngleDeg, diff)
			}
			failures++
		}
	}
	if failures > 0 {
		t.Errorf("%d phase angle failures out of %d tests (tol=%.0e°)", failures, len(golden.Tests), tol)
	}
	t.Logf("Phase angle: %d tests, maxDiff=%.2e°", len(golden.Tests), maxDiff)
}

func TestPositionAngle_NorthSouth(t *testing.T) {
	// Two points on the same RA, different Dec: PA should be 0 (north) or 180 (south)
	pa := PositionAngle(6, 0, 6, 10)
	if math.Abs(pa) > 1e-10 {
		t.Errorf("due north: got %f, want 0", pa)
	}

	pa = PositionAngle(6, 10, 6, 0)
	if math.Abs(pa-180) > 1e-10 {
		t.Errorf("due south: got %f, want 180", pa)
	}
}

func TestPositionAngle_East(t *testing.T) {
	// At equator, increasing RA = east = PA 90°
	pa := PositionAngle(6, 0, 6.01, 0)
	if math.Abs(pa-90) > 0.1 {
		t.Errorf("east: got %f, want ~90", pa)
	}
}

func TestElongation_KnownValues(t *testing.T) {
	tests := []struct {
		target, ref, want float64
	}{
		{90, 0, 90},
		{0, 90, 270},
		{180, 0, 180},
		{10, 350, 20},
		{350, 10, 340},
	}
	for _, tc := range tests {
		got := Elongation(tc.target, tc.ref)
		if math.Abs(got-tc.want) > 1e-12 {
			t.Errorf("Elongation(%f, %f) = %f, want %f", tc.target, tc.ref, got, tc.want)
		}
	}
}

// goldenElongation matches testdata/golden_elongation.json.
type goldenElongation struct {
	Tests []struct {
		TDBJD         float64 `json:"tdb_jd"`
		MoonEclLon    float64 `json:"moon_ecl_lon_deg"`
		SunEclLon     float64 `json:"sun_ecl_lon_deg"`
		ElongationDeg float64 `json:"elongation_deg"`
	} `json:"tests"`
}

func TestElongation_Golden(t *testing.T) {
	var golden goldenElongation
	loadJSON(t, "../testdata/golden_elongation.json", &golden)

	const tol = 1e-10 // should match exactly since it's just mod arithmetic
	failures := 0
	for i, tc := range golden.Tests {
		got := Elongation(tc.MoonEclLon, tc.SunEclLon)
		diff := math.Abs(got - tc.ElongationDeg)
		if diff > 180 {
			diff = 360 - diff
		}
		if diff > tol {
			if failures < 10 {
				t.Errorf("test %d: got=%.10f want=%.10f diff=%.2e",
					i, got, tc.ElongationDeg, diff)
			}
			failures++
		}
	}
	if failures > 0 {
		t.Errorf("%d elongation failures out of %d tests", failures, len(golden.Tests))
	}
}

func BenchmarkSeparationAngle(b *testing.B) {
	a := [3]float64{1e8, -5e7, 2e7}
	v := [3]float64{-3e7, 4e7, 1e7}
	for i := 0; i < b.N; i++ {
		SeparationAngle(a, v)
	}
}

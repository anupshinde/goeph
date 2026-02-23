package eclipse

import (
	"math"
	"os"
	"testing"

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

func TestFindLunarEclipses_2024(t *testing.T) {
	// 2024 has two lunar eclipses:
	//   1. March 25, 2024 — Penumbral
	//   2. September 18, 2024 — Partial (umbral magnitude ~0.08)
	// Reference: NASA eclipse catalog.
	startJD := 2460310.5 // ~2024-01-01 TT
	endJD := startJD + 365.25

	eclipses, err := FindLunarEclipses(testEph, startJD, endJD)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("found %d lunar eclipses in 2024", len(eclipses))
	for i, e := range eclipses {
		kindName := "unknown"
		switch e.Kind {
		case Penumbral:
			kindName = "Penumbral"
		case Partial:
			kindName = "Partial"
		case Total:
			kindName = "Total"
		}
		t.Logf("  eclipse %d: JD=%.4f, type=%s, umbral_mag=%.4f, penumbral_mag=%.4f, sep=%.0f km",
			i, e.T, kindName, e.UmbralMag, e.PenumbralMag, e.ClosestApproachKm)
	}

	if len(eclipses) < 2 {
		t.Errorf("expected at least 2 lunar eclipses in 2024, got %d", len(eclipses))
	}
}

func TestFindLunarEclipses_TotalEclipse(t *testing.T) {
	// November 8, 2022 was a total lunar eclipse.
	// JD ~2459892 (2022-11-08).
	startJD := 2459850.0 // ~2022-10-01
	endJD := 2459950.0   // ~2023-01-09

	eclipses, err := FindLunarEclipses(testEph, startJD, endJD)
	if err != nil {
		t.Fatal(err)
	}

	foundTotal := false
	for _, e := range eclipses {
		t.Logf("eclipse: JD=%.4f, kind=%d, umbral_mag=%.4f", e.T, e.Kind, e.UmbralMag)
		if e.Kind == Total {
			foundTotal = true
			// Total eclipse should have umbral magnitude > 1.
			if e.UmbralMag < 1.0 {
				t.Errorf("total eclipse has umbral mag %.4f, want >= 1.0", e.UmbralMag)
			}
		}
	}

	if !foundTotal {
		t.Error("expected a total lunar eclipse near 2022-11-08")
	}
}

func TestFindLunarEclipses_Decade(t *testing.T) {
	// Over 10 years, there should be roughly 15-25 lunar eclipses.
	// (Average ~2.4 per year.)
	startJD := 2451545.0          // J2000
	endJD := startJD + 10*365.25  // 10 years

	eclipses, err := FindLunarEclipses(testEph, startJD, endJD)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("found %d lunar eclipses in 10 years (2000-2010)", len(eclipses))
	if len(eclipses) < 10 || len(eclipses) > 35 {
		t.Errorf("got %d eclipses, want 10-35 for a decade", len(eclipses))
	}

	// Verify all eclipses have valid fields.
	for i, e := range eclipses {
		if e.Kind < Penumbral || e.Kind > Total {
			t.Errorf("eclipse %d: invalid kind %d", i, e.Kind)
		}
		if e.PenumbralMag <= 0 {
			t.Errorf("eclipse %d: penumbral mag %.4f, want > 0", i, e.PenumbralMag)
		}
		if e.ClosestApproachKm < 0 {
			t.Errorf("eclipse %d: negative separation %.0f km", i, e.ClosestApproachKm)
		}
		if e.UmbralRadiusKm < 0 || e.UmbralRadiusKm > 10000 {
			t.Errorf("eclipse %d: unreasonable umbral radius %.0f km", i, e.UmbralRadiusKm)
		}
		if e.PenumbralRadiusKm < e.UmbralRadiusKm {
			t.Errorf("eclipse %d: penumbral radius %.0f < umbral %.0f",
				i, e.PenumbralRadiusKm, e.UmbralRadiusKm)
		}
	}

	// Count types.
	counts := map[int]int{}
	for _, e := range eclipses {
		counts[e.Kind]++
	}
	t.Logf("types: penumbral=%d, partial=%d, total=%d",
		counts[Penumbral], counts[Partial], counts[Total])
}

func TestFindLunarEclipses_Ordering(t *testing.T) {
	startJD := 2451545.0
	endJD := startJD + 5*365.25

	eclipses, err := FindLunarEclipses(testEph, startJD, endJD)
	if err != nil {
		t.Fatal(err)
	}

	for i := 1; i < len(eclipses); i++ {
		if eclipses[i].T <= eclipses[i-1].T {
			t.Errorf("eclipses not sorted: eclipse %d at %.4f <= eclipse %d at %.4f",
				i, eclipses[i].T, i-1, eclipses[i-1].T)
		}
	}
}

func TestMoonShadowSeparation(t *testing.T) {
	// At a known eclipse time, separation should be small.
	// At a non-eclipse time, separation should be large.

	// Non-eclipse: Moon at first quarter (elongation ~90°).
	// Use a date roughly at first quarter.
	sepQuarter := moonShadowSeparation(testEph, 2451552.0) // ~Jan 8 2000

	// Near a full moon (elongation ~180°).
	sepFull := moonShadowSeparation(testEph, 2451565.0) // ~Jan 21 2000

	// At first quarter, Moon should be far from shadow axis.
	// At full moon, much closer.
	if sepFull >= sepQuarter {
		t.Errorf("full moon separation %.0f km >= quarter moon %.0f km", sepFull, sepQuarter)
	}

	t.Logf("quarter moon separation: %.0f km, full moon: %.0f km", sepQuarter, sepFull)

	// Shadow axis separation at quarter moon should be > 200,000 km (roughly Moon orbital distance).
	if sepQuarter < 100000 {
		t.Errorf("quarter moon separation %.0f km, want > 100000", sepQuarter)
	}
}

func TestEclipticElongation(t *testing.T) {
	// Test with simple vectors.
	// Moon at ecliptic lon=0, Sun at ecliptic lon=0 → elongation = 0.
	moon := [3]float64{1, 0, 0}
	sun := [3]float64{1, 0, 0}
	elong := eclipticElongation(moon, sun)
	if math.Abs(elong) > 1e-10 && math.Abs(elong-360) > 1e-10 {
		t.Errorf("same direction: elongation = %.4f, want 0 or 360", elong)
	}

	// Moon at ecliptic lon=180° → elongation = 180.
	moon2 := [3]float64{-1, 0, 0}
	elong2 := eclipticElongation(moon2, sun)
	if math.Abs(elong2-180) > 1e-10 {
		t.Errorf("opposite direction: elongation = %.4f, want 180", elong2)
	}
}

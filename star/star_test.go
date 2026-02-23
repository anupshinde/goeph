package star

import (
	"math"
	"testing"
)

func TestGalacticCenterICRF_UnitVector(t *testing.T) {
	x, y, z := GalacticCenterICRF()
	r := math.Sqrt(x*x + y*y + z*z)
	if math.Abs(r-1.0) > 1e-15 {
		t.Errorf("not a unit vector: |r| = %.15f", r)
	}
}

func TestGalacticCenterICRF_Direction(t *testing.T) {
	x, y, z := GalacticCenterICRF()

	// GC is at RA ~17.76h, Dec ~-29° → should have:
	// - negative x (RA > 12h means x < 0 for large RA)
	// - negative y (RA ~17.76h → cos(266.4°) < 0, sin(266.4°) < 0)
	// - negative z (Dec < 0)
	if z >= 0 {
		t.Errorf("expected negative z for Dec<0, got %f", z)
	}

	// RA = 17.76h = 266.4° → in third quadrant, both cos and sin negative
	raRad := gcRAHours * 15.0 * math.Pi / 180.0
	expectedSignX := math.Cos(raRad) // should be slightly negative (~-0.063)
	if (x > 0) != (expectedSignX > 0) {
		t.Errorf("x sign mismatch: got %f, expected sign of cos(RA)=%f", x, expectedSignX)
	}

	// Verify the coordinates are consistent with the declared RA/Dec
	dec := math.Asin(z) * 180.0 / math.Pi
	ra := math.Atan2(y, x) * 180.0 / math.Pi
	if ra < 0 {
		ra += 360
	}
	raHours := ra / 15.0

	if math.Abs(raHours-gcRAHours) > 1e-10 {
		t.Errorf("RA mismatch: got %.10f h, want %.10f h", raHours, gcRAHours)
	}
	if math.Abs(dec-gcDecDeg) > 1e-10 {
		t.Errorf("Dec mismatch: got %.10f°, want %.10f°", dec, gcDecDeg)
	}
}

// --- Star type tests ---

func TestStar_NoMotion(t *testing.T) {
	// A star with no proper motion should return the same RA/Dec at any epoch.
	s := &Star{
		RAHours: 6.0,
		DecDeg:  45.0,
	}
	ra, dec := s.RADec(j2000)
	if math.Abs(ra-6.0) > 1e-12 {
		t.Errorf("RA = %.12f, want 6.0", ra)
	}
	if math.Abs(dec-45.0) > 1e-12 {
		t.Errorf("Dec = %.12f, want 45.0", dec)
	}

	// 100 years later should be essentially the same.
	ra2, dec2 := s.RADec(j2000 + 36525)
	if math.Abs(ra2-6.0) > 1e-6 {
		t.Errorf("RA after 100y = %.12f, want ~6.0", ra2)
	}
	if math.Abs(dec2-45.0) > 1e-6 {
		t.Errorf("Dec after 100y = %.12f, want ~45.0", dec2)
	}
}

func TestStar_ProperMotion_BarnardsStar(t *testing.T) {
	// Barnard's Star — largest known proper motion.
	// Hipparcos catalog (epoch J1991.25 = JD 2448349.0625):
	//   RA  = 17h 57m 48.49803s
	//   Dec = +04° 41' 36.2072"
	//   pmRA  = -798.71 mas/yr
	//   pmDec = 10337.77 mas/yr
	//   parallax = 548.31 mas
	//   RV = -110.6 km/s
	s := &Star{
		RAHours:       17.0 + 57.0/60.0 + 48.49803/3600.0,
		DecDeg:        4.0 + 41.0/60.0 + 36.2072/3600.0,
		RAMasPerYear:  -798.71,
		DecMasPerYear: 10337.77,
		ParallaxMas:   548.31,
		RadialKmPerS:  -110.6,
		Epoch:         2448349.0625, // J1991.25
	}

	// At the catalog epoch, RA/Dec should match input.
	ra0, dec0 := s.RADec(s.Epoch)
	if math.Abs(ra0-s.RAHours) > 1e-8 {
		t.Errorf("RA at epoch = %.10f, want %.10f", ra0, s.RAHours)
	}
	if math.Abs(dec0-s.DecDeg) > 1e-8 {
		t.Errorf("Dec at epoch = %.10f, want %.10f", dec0, s.DecDeg)
	}

	// After 10 years (~3652.5 days), Dec should change by ~103.4 arcsec.
	// 10337.77 mas/yr * 10 yr = 103377.7 mas = 103.378 arcsec ≈ 0.0287°.
	_, dec10 := s.RADec(s.Epoch + 3652.5)
	decShift := dec10 - dec0
	expectedDecShift := 10337.77 * 10.0 / 3600000.0 // degrees
	if math.Abs(decShift-expectedDecShift)/expectedDecShift > 0.01 {
		t.Errorf("Dec shift after 10y = %.6f°, want ~%.6f° (%.1f%%)",
			decShift, expectedDecShift,
			100*math.Abs(decShift-expectedDecShift)/expectedDecShift)
	}
}

func TestStar_PositionAU(t *testing.T) {
	// Star at RA=0h, Dec=0°, parallax=1000 mas (1 arcsec = 1 parsec = 206265 AU).
	// 1000 mas = 1 arcsec → distance = 1/sin(1") ≈ 206265 AU.
	s := &Star{
		RAHours:     0,
		DecDeg:      0,
		ParallaxMas: 1000.0,
	}
	pos := s.PositionAU(j2000)

	// Should be along x-axis at ~206265 AU.
	dist := math.Sqrt(pos[0]*pos[0] + pos[1]*pos[1] + pos[2]*pos[2])
	expected := 1.0 / math.Sin(1000.0*1e-3*math.Pi/(180.0*3600.0))
	if math.Abs(dist-expected)/expected > 1e-10 {
		t.Errorf("distance = %.2f AU, want %.2f AU", dist, expected)
	}
	if math.Abs(pos[0]-dist) > 1 {
		t.Errorf("pos[0] = %.2f, want ~%.2f (along x-axis)", pos[0], dist)
	}
}

func TestStar_ZeroParallax(t *testing.T) {
	// Zero parallax should not panic; effectively infinite distance.
	s := &Star{
		RAHours: 12.0,
		DecDeg:  30.0,
	}
	pos := s.PositionAU(j2000)
	dist := math.Sqrt(pos[0]*pos[0] + pos[1]*pos[1] + pos[2]*pos[2])
	if dist < 1e6 {
		t.Errorf("distance with zero parallax = %.0f AU, want very large", dist)
	}
}

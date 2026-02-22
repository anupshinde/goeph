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

package elements

import (
	"math"
	"testing"
)

func TestCircularOrbit(t *testing.T) {
	// Circular orbit: r=7000 km, v=sqrt(mu/r) tangential
	muSun := 132712440041.94 // km³/s² for Sun
	r := 1.496e8             // ~1 AU in km
	v := math.Sqrt(muSun / r)

	pos := [3]float64{r, 0, 0}
	vel := [3]float64{0, v, 0}

	el := FromStateVector(pos, vel, muSun)

	if math.Abs(el.Eccentricity) > 1e-10 {
		t.Errorf("circular orbit: eccentricity = %e, want ~0", el.Eccentricity)
	}
	if math.Abs(el.SemiMajorAxisKm-r)/r > 1e-10 {
		t.Errorf("circular orbit: a = %f, want %f", el.SemiMajorAxisKm, r)
	}
	if math.Abs(el.InclinationDeg) > 1e-10 {
		t.Errorf("circular orbit: inc = %f, want 0", el.InclinationDeg)
	}
}

func TestEllipticalOrbit(t *testing.T) {
	// Earth-like orbit: a=1 AU, e=0.0167
	muSun := 132712440041.94
	a := 1.496e8
	e := 0.0167
	// At periapsis: r = a(1-e), v = sqrt(mu*(2/r - 1/a))
	rPeri := a * (1 - e)
	vPeri := math.Sqrt(muSun * (2.0/rPeri - 1.0/a))

	pos := [3]float64{rPeri, 0, 0}
	vel := [3]float64{0, vPeri, 0}

	el := FromStateVector(pos, vel, muSun)

	if math.Abs(el.Eccentricity-e)/e > 1e-6 {
		t.Errorf("eccentricity = %f, want %f", el.Eccentricity, e)
	}
	if math.Abs(el.SemiMajorAxisKm-a)/a > 1e-6 {
		t.Errorf("a = %f km, want %f km", el.SemiMajorAxisKm, a)
	}
	// At periapsis, true anomaly should be 0
	if math.Abs(el.TrueAnomalyDeg) > 1e-6 {
		t.Errorf("true anomaly = %f°, want ~0°", el.TrueAnomalyDeg)
	}
	// Period should be ~365.25 days
	if math.Abs(el.PeriodDays-365.25)/365.25 > 0.01 {
		t.Errorf("period = %f days, want ~365.25", el.PeriodDays)
	}
}

func TestInclinedOrbit(t *testing.T) {
	// 45° inclination orbit
	muSun := 132712440041.94
	r := 1.496e8
	v := math.Sqrt(muSun / r)

	// Velocity tilted 45° out of XY plane
	pos := [3]float64{r, 0, 0}
	vel := [3]float64{0, v * math.Cos(math.Pi/4), v * math.Sin(math.Pi/4)}

	el := FromStateVector(pos, vel, muSun)

	if math.Abs(el.InclinationDeg-45.0) > 0.01 {
		t.Errorf("inclination = %f°, want 45°", el.InclinationDeg)
	}
}

func TestPeriapsisApoapsis(t *testing.T) {
	muSun := 132712440041.94
	a := 1.5e8
	e := 0.5
	rPeri := a * (1 - e)
	vPeri := math.Sqrt(muSun * (2.0/rPeri - 1.0/a))

	pos := [3]float64{rPeri, 0, 0}
	vel := [3]float64{0, vPeri, 0}

	el := FromStateVector(pos, vel, muSun)

	wantQ := a * (1 - e)
	wantApo := a * (1 + e)
	if math.Abs(el.PeriapsisDistanceKm-wantQ)/wantQ > 1e-6 {
		t.Errorf("periapsis = %f, want %f", el.PeriapsisDistanceKm, wantQ)
	}
	if math.Abs(el.ApoapsisDistanceKm-wantApo)/wantApo > 1e-6 {
		t.Errorf("apoapsis = %f, want %f", el.ApoapsisDistanceKm, wantApo)
	}
}

package kepler

import (
	"math"
	"testing"
)

const j2000 = 2451545.0

// Ceres orbital elements (MPC, J2000 ecliptic)
var ceresOrbit = Orbit{
	SemiMajorAxisAU: 2.7670463,
	Eccentricity:    0.0785115,
	InclinationDeg:  10.5868,
	LongAscNodeDeg:  80.3055,
	ArgPeriapsisDeg: 73.5977,
	MeanAnomalyDeg:  77.372,
	EpochJD:         2451545.0, // J2000.0
}

// Halley's Comet orbital elements (ecliptic J2000)
var halleyOrbit = Orbit{
	PerihelionAU:    0.586,
	Eccentricity:    0.9671,
	InclinationDeg:  162.26,
	LongAscNodeDeg:  58.42,
	ArgPeriapsisDeg: 111.33,
	PeriapsisTimeJD: 2446467.395, // 1986-02-09
}

func TestOrbit_CircularAtEpoch(t *testing.T) {
	// Circular orbit at 1 AU: position should be at distance 1 AU.
	o := &Orbit{
		SemiMajorAxisAU: 1.0,
		Eccentricity:    0.0,
		InclinationDeg:  0.0,
		LongAscNodeDeg:  0.0,
		ArgPeriapsisDeg: 0.0,
		MeanAnomalyDeg:  0.0,
		EpochJD:         j2000,
	}

	pos := o.PositionAU(j2000)
	dist := math.Sqrt(pos[0]*pos[0] + pos[1]*pos[1] + pos[2]*pos[2])
	if math.Abs(dist-1.0) > 1e-10 {
		t.Errorf("circular orbit distance = %.10f AU, want 1.0", dist)
	}
}

func TestOrbit_CircularHalfPeriod(t *testing.T) {
	// Circular orbit at 1 AU. After half the period, should be at opposite side.
	o := &Orbit{
		SemiMajorAxisAU: 1.0,
		Eccentricity:    0.0,
		InclinationDeg:  0.0,
		LongAscNodeDeg:  0.0,
		ArgPeriapsisDeg: 0.0,
		MeanAnomalyDeg:  0.0,
		EpochJD:         j2000,
	}

	// Period = 2π/n, half period = π/n.
	o.init()
	halfPeriod := math.Pi / o.n

	pos0 := o.PositionAU(j2000)
	pos1 := o.PositionAU(j2000 + halfPeriod)

	// Positions should be opposite: pos0 ≈ -pos1.
	for i := 0; i < 3; i++ {
		if math.Abs(pos0[i]+pos1[i]) > 1e-8 {
			t.Errorf("axis %d: pos0=%.8f, pos1=%.8f, sum=%.8f (want ~0)",
				i, pos0[i], pos1[i], pos0[i]+pos1[i])
		}
	}
}

func TestOrbit_EllipticPerihelion(t *testing.T) {
	// At periapsis (M=0), distance should equal a*(1-e).
	o := &Orbit{
		SemiMajorAxisAU: 2.0,
		Eccentricity:    0.5,
		InclinationDeg:  0.0,
		LongAscNodeDeg:  0.0,
		ArgPeriapsisDeg: 0.0,
		MeanAnomalyDeg:  0.0, // at periapsis
		EpochJD:         j2000,
	}

	pos := o.PositionAU(j2000)
	dist := math.Sqrt(pos[0]*pos[0] + pos[1]*pos[1] + pos[2]*pos[2])
	expected := 2.0 * (1.0 - 0.5) // a*(1-e) = 1.0 AU
	if math.Abs(dist-expected) > 1e-10 {
		t.Errorf("perihelion distance = %.10f AU, want %.10f", dist, expected)
	}
}

func TestOrbit_EllipticAphelion(t *testing.T) {
	// At apoapsis (M=180°), distance should equal a*(1+e).
	o := &Orbit{
		SemiMajorAxisAU: 2.0,
		Eccentricity:    0.5,
		InclinationDeg:  0.0,
		LongAscNodeDeg:  0.0,
		ArgPeriapsisDeg: 0.0,
		MeanAnomalyDeg:  180.0, // at apoapsis
		EpochJD:         j2000,
	}

	pos := o.PositionAU(j2000)
	dist := math.Sqrt(pos[0]*pos[0] + pos[1]*pos[1] + pos[2]*pos[2])
	expected := 2.0 * (1.0 + 0.5) // a*(1+e) = 3.0 AU
	if math.Abs(dist-expected) > 1e-10 {
		t.Errorf("aphelion distance = %.10f AU, want %.10f", dist, expected)
	}
}

func TestOrbit_Ceres_Periodicity(t *testing.T) {
	// Ceres should return to the same position after one orbital period.
	o := &ceresOrbit
	o.init()

	period := 2 * math.Pi / o.n // days
	pos0 := o.PositionAU(j2000)
	pos1 := o.PositionAU(j2000 + period)

	for i := 0; i < 3; i++ {
		if math.Abs(pos0[i]-pos1[i]) > 1e-8 {
			t.Errorf("axis %d: pos0=%.10f, pos1=%.10f, diff=%.2e",
				i, pos0[i], pos1[i], pos0[i]-pos1[i])
		}
	}
}

func TestOrbit_Ceres_Distance(t *testing.T) {
	// Ceres distance should be between perihelion and aphelion.
	o := &ceresOrbit
	qExpected := o.SemiMajorAxisAU * (1.0 - o.Eccentricity) // ~2.55 AU
	QExpected := o.SemiMajorAxisAU * (1.0 + o.Eccentricity) // ~2.98 AU

	for dt := 0.0; dt < 1600; dt += 100 {
		pos := o.PositionAU(j2000 + dt)
		dist := math.Sqrt(pos[0]*pos[0] + pos[1]*pos[1] + pos[2]*pos[2])
		if dist < qExpected-0.01 || dist > QExpected+0.01 {
			t.Errorf("dt=%.0f: distance=%.4f AU, want [%.4f, %.4f]",
				dt, dist, qExpected, QExpected)
		}
	}
}

func TestOrbit_Halley_Distance(t *testing.T) {
	// At periapsis (1986-02-09), distance should be ~0.586 AU.
	o := &halleyOrbit
	pos := o.PositionAU(o.PeriapsisTimeJD)
	dist := math.Sqrt(pos[0]*pos[0] + pos[1]*pos[1] + pos[2]*pos[2])
	if math.Abs(dist-o.PerihelionAU) > 0.001 {
		t.Errorf("Halley perihelion distance = %.6f AU, want %.6f", dist, o.PerihelionAU)
	}
}

func TestOrbit_Halley_HighEccentricity(t *testing.T) {
	// Halley (e=0.9671) — verify distance varies over a wide range.
	o := &halleyOrbit
	minDist := math.MaxFloat64
	maxDist := 0.0

	// Sample 200 points over ~5 years near periapsis.
	for i := 0; i < 200; i++ {
		dt := float64(i-100) * 20.0 // ±2000 days from periapsis
		pos := o.PositionAU(o.PeriapsisTimeJD + dt)
		dist := math.Sqrt(pos[0]*pos[0] + pos[1]*pos[1] + pos[2]*pos[2])
		if dist < minDist {
			minDist = dist
		}
		if dist > maxDist {
			maxDist = dist
		}
	}

	t.Logf("Halley distance range: %.3f - %.3f AU", minDist, maxDist)
	if minDist > 0.6 {
		t.Errorf("min distance %.3f AU, want < 0.6 (near perihelion)", minDist)
	}
	if maxDist < 5.0 {
		t.Errorf("max distance %.3f AU, want > 5.0 (far from perihelion)", maxDist)
	}
}

func TestOrbit_Inclination(t *testing.T) {
	// Orbit with 90° inclination should have z component.
	o := &Orbit{
		SemiMajorAxisAU: 1.0,
		Eccentricity:    0.0,
		InclinationDeg:  90.0,
		LongAscNodeDeg:  0.0,
		ArgPeriapsisDeg: 0.0,
		MeanAnomalyDeg:  90.0, // quarter orbit
		EpochJD:         j2000,
	}

	pos := o.PositionAU(j2000)
	// For i=90°, Ω=0, ω=0, M=90°: position should be along ecliptic z-axis
	// which in ICRF is rotated by obliquity. The ecliptic z-component is 1 AU.
	dist := math.Sqrt(pos[0]*pos[0] + pos[1]*pos[1] + pos[2]*pos[2])
	if math.Abs(dist-1.0) > 1e-8 {
		t.Errorf("distance = %.10f AU, want 1.0", dist)
	}
}

func TestOrbit_PositionKm(t *testing.T) {
	// Verify km conversion.
	o := &Orbit{
		SemiMajorAxisAU: 1.0,
		Eccentricity:    0.0,
		InclinationDeg:  0.0,
		LongAscNodeDeg:  0.0,
		ArgPeriapsisDeg: 0.0,
		MeanAnomalyDeg:  0.0,
		EpochJD:         j2000,
	}

	posAU := o.PositionAU(j2000)
	posKm := o.PositionKm(j2000)

	for i := 0; i < 3; i++ {
		expected := posAU[i] * auKm
		if math.Abs(posKm[i]-expected) > 1e-3 {
			t.Errorf("axis %d: km=%.3f, want %.3f", i, posKm[i], expected)
		}
	}
}

func TestOrbit_ParabolicBarker(t *testing.T) {
	// Parabolic orbit (e=1): at periapsis, distance should equal q.
	o := &Orbit{
		PerihelionAU:    1.0,
		Eccentricity:    1.0,
		InclinationDeg:  0.0,
		LongAscNodeDeg:  0.0,
		ArgPeriapsisDeg: 0.0,
		PeriapsisTimeJD: j2000,
	}

	pos := o.PositionAU(j2000)
	dist := math.Sqrt(pos[0]*pos[0] + pos[1]*pos[1] + pos[2]*pos[2])
	if math.Abs(dist-1.0) > 1e-8 {
		t.Errorf("parabolic periapsis distance = %.10f AU, want 1.0", dist)
	}

	// After some time, distance should increase.
	pos2 := o.PositionAU(j2000 + 100)
	dist2 := math.Sqrt(pos2[0]*pos2[0] + pos2[1]*pos2[1] + pos2[2]*pos2[2])
	if dist2 <= dist {
		t.Errorf("parabolic distance did not increase: at t0=%.4f, at t0+100d=%.4f", dist, dist2)
	}
}

func TestOrbit_HyperbolicPeriapsis(t *testing.T) {
	// Hyperbolic orbit (e=1.5): at periapsis, distance should equal q.
	o := &Orbit{
		PerihelionAU:    1.0,
		Eccentricity:    1.5,
		InclinationDeg:  0.0,
		LongAscNodeDeg:  0.0,
		ArgPeriapsisDeg: 0.0,
		PeriapsisTimeJD: j2000,
	}

	pos := o.PositionAU(j2000)
	dist := math.Sqrt(pos[0]*pos[0] + pos[1]*pos[1] + pos[2]*pos[2])
	if math.Abs(dist-1.0) > 1e-6 {
		t.Errorf("hyperbolic periapsis distance = %.10f AU, want 1.0", dist)
	}
}

func TestSolveElliptic_KnownValues(t *testing.T) {
	// For circular orbit (e=0), E = M, ν = M.
	o := &Orbit{
		SemiMajorAxisAU: 1.0,
		Eccentricity:    0.0,
		EpochJD:         j2000,
	}
	o.init()

	tests := []float64{0.0, math.Pi / 4, math.Pi / 2, math.Pi, -math.Pi / 3}
	for _, M := range tests {
		nu, r := o.solveElliptic(M)
		if math.Abs(nu-M) > 1e-14 {
			t.Errorf("e=0, M=%.4f: nu=%.10f, want %.10f", M, nu, M)
		}
		if math.Abs(r-1.0) > 1e-14 {
			t.Errorf("e=0, M=%.4f: r=%.10f, want 1.0", M, r)
		}
	}
}

func TestOrbit_EarthLikeOrbit(t *testing.T) {
	// Earth-like orbit: a=1AU, e=0.0167, i=0.
	// Mean motion n ≈ 0.01720/day (Gaussian constant).
	o := &Orbit{
		SemiMajorAxisAU: 1.0,
		Eccentricity:    0.0167,
		InclinationDeg:  0.0,
		LongAscNodeDeg:  0.0,
		ArgPeriapsisDeg: 0.0,
		MeanAnomalyDeg:  0.0,
		EpochJD:         j2000,
	}
	o.init()

	// Period should be ~365.25 days.
	period := 2 * math.Pi / o.n
	if math.Abs(period-365.25) > 0.5 {
		t.Errorf("Earth-like period = %.2f days, want ~365.25", period)
	}

	// Distance at perihelion.
	pos := o.PositionAU(j2000)
	dist := math.Sqrt(pos[0]*pos[0] + pos[1]*pos[1] + pos[2]*pos[2])
	expectedQ := 1.0 * (1.0 - 0.0167)
	if math.Abs(dist-expectedQ) > 1e-8 {
		t.Errorf("perihelion distance = %.10f AU, want %.10f", dist, expectedQ)
	}
}

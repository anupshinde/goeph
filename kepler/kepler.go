// Package kepler provides Keplerian orbit propagation for minor planets and
// comets. Given orbital elements at an epoch, it computes heliocentric
// position at any time using Kepler's equation.
//
// Orbital elements are in the J2000 ecliptic frame, matching the convention
// used by the Minor Planet Center and JPL. Returned positions are in the
// ICRF (equatorial) frame for compatibility with the rest of goeph.
package kepler

import "math"

const (
	// GMSunAU3D2 is the gravitational parameter of the Sun in AU³/day².
	// Equal to the square of the Gaussian gravitational constant k.
	GMSunAU3D2 = 2.9591220828559115e-4

	// auKm is the IAU astronomical unit in km.
	auKm = 149597870.7

	deg2rad = math.Pi / 180.0
	rad2deg = 180.0 / math.Pi

	// J2000 mean obliquity: 84381.448 arcseconds (Lieske 1979).
	obliquitySin = 0.3977771559319137062
	obliquityCos = 0.9174820620691818140
)

// Orbit represents a Keplerian orbit defined by classical orbital elements.
// Elements are in the J2000 ecliptic frame.
type Orbit struct {
	// SemiMajorAxisAU is the semi-major axis in AU.
	// Required for elliptic orbits (e < 1). For parabolic (e = 1),
	// use PerihelionAU instead.
	SemiMajorAxisAU float64

	// PerihelionAU is the perihelion distance in AU.
	// If zero, computed from SemiMajorAxisAU * (1 - Eccentricity).
	PerihelionAU float64

	// Eccentricity of the orbit. 0 ≤ e < 1 = elliptic, e = 1 = parabolic, e > 1 = hyperbolic.
	Eccentricity float64

	// InclinationDeg is the orbital inclination in degrees.
	InclinationDeg float64

	// LongAscNodeDeg is the longitude of the ascending node (Ω) in degrees.
	LongAscNodeDeg float64

	// ArgPeriapsisDeg is the argument of periapsis (ω) in degrees.
	ArgPeriapsisDeg float64

	// MeanAnomalyDeg is the mean anomaly at EpochJD, in degrees.
	// For comets, set PeriapsisTimeJD instead.
	MeanAnomalyDeg float64

	// EpochJD is the TDB Julian date at which the elements are valid.
	EpochJD float64

	// PeriapsisTimeJD is the TDB Julian date of periapsis passage.
	// If set (non-zero), overrides MeanAnomalyDeg.
	PeriapsisTimeJD float64

	// GM is the gravitational parameter of the central body in AU³/day².
	// If zero, GMSunAU3D2 (Sun) is used.
	GM float64

	// precomputed
	ready bool
	mu    float64 // GM in AU³/day²
	a     float64 // semi-major axis in AU
	q     float64 // perihelion distance in AU
	e     float64 // eccentricity
	n     float64 // mean motion in rad/day
	rot   [3][3]float64
}

// init precomputes derived quantities. Called lazily on first use.
func (o *Orbit) init() {
	if o.ready {
		return
	}
	o.ready = true

	o.mu = o.GM
	if o.mu == 0 {
		o.mu = GMSunAU3D2
	}

	o.e = o.Eccentricity

	// Compute semi-major axis and perihelion distance.
	if o.SemiMajorAxisAU != 0 {
		o.a = o.SemiMajorAxisAU
		o.q = o.a * (1.0 - o.e)
	} else if o.PerihelionAU != 0 {
		o.q = o.PerihelionAU
		if o.e < 1.0 {
			o.a = o.q / (1.0 - o.e)
		}
	}

	// Mean motion (rad/day) for elliptic orbits.
	if o.e < 1.0 && o.a > 0 {
		o.n = math.Sqrt(o.mu / (o.a * o.a * o.a))
	}

	// Rotation matrix from perifocal (PQW) frame to ecliptic J2000.
	i := o.InclinationDeg * deg2rad
	omega := o.LongAscNodeDeg * deg2rad
	w := o.ArgPeriapsisDeg * deg2rad

	sinI, cosI := math.Sincos(i)
	sinO, cosO := math.Sincos(omega)
	sinW, cosW := math.Sincos(w)

	// R = Rz(-Ω) · Rx(-i) · Rz(-ω)
	// Columns of R are the P, Q, W unit vectors in the ecliptic frame.
	o.rot = [3][3]float64{
		{cosO*cosW - sinO*sinW*cosI, -cosO*sinW - sinO*cosW*cosI, sinO * sinI},
		{sinO*cosW + cosO*sinW*cosI, -sinO*sinW + cosO*cosW*cosI, -cosO * sinI},
		{sinW * sinI, cosW * sinI, cosI},
	}
}

// PositionAU returns the heliocentric ICRF position in AU at the given
// TDB Julian date.
func (o *Orbit) PositionAU(tdbJD float64) [3]float64 {
	o.init()

	// Compute mean anomaly at time t.
	M := o.meanAnomalyAt(tdbJD)

	// Solve Kepler's equation for true anomaly and radius.
	var nu, r float64
	switch {
	case o.e < 1.0:
		nu, r = o.solveElliptic(M)
	case o.e == 1.0:
		nu, r = o.solveParabolic(M)
	default:
		nu, r = o.solveHyperbolic(M)
	}

	// Position in the perifocal (PQW) frame.
	cosNu, sinNu := math.Cos(nu), math.Sin(nu)
	xPQW := r * cosNu
	yPQW := r * sinNu

	// Rotate perifocal → ecliptic J2000.
	xEcl := o.rot[0][0]*xPQW + o.rot[0][1]*yPQW
	yEcl := o.rot[1][0]*xPQW + o.rot[1][1]*yPQW
	zEcl := o.rot[2][0]*xPQW + o.rot[2][1]*yPQW

	// Rotate ecliptic → equatorial (ICRF).
	// Rx(-ε): x' = x, y' = cos(ε)*y - sin(ε)*z, z' = sin(ε)*y + cos(ε)*z
	return [3]float64{
		xEcl,
		obliquityCos*yEcl - obliquitySin*zEcl,
		obliquitySin*yEcl + obliquityCos*zEcl,
	}
}

// PositionKm returns the heliocentric ICRF position in km at the given
// TDB Julian date.
func (o *Orbit) PositionKm(tdbJD float64) [3]float64 {
	pos := o.PositionAU(tdbJD)
	return [3]float64{
		pos[0] * auKm,
		pos[1] * auKm,
		pos[2] * auKm,
	}
}

// meanAnomalyAt computes the mean anomaly in radians at time tdbJD.
func (o *Orbit) meanAnomalyAt(tdbJD float64) float64 {
	if o.PeriapsisTimeJD != 0 {
		dt := tdbJD - o.PeriapsisTimeJD // days since periapsis
		if o.e < 1.0 {
			return o.n * dt
		}
		// For parabolic/hyperbolic, return dt directly (handled by solver).
		return dt
	}
	// Use mean anomaly at epoch + mean motion.
	M0 := o.MeanAnomalyDeg * deg2rad
	dt := tdbJD - o.EpochJD
	return M0 + o.n*dt
}

// solveElliptic solves Kepler's equation M = E - e*sin(E) for an elliptic orbit.
// Returns true anomaly (radians) and radius (AU).
func (o *Orbit) solveElliptic(M float64) (nu, r float64) {
	e := o.e

	// Normalize M to [-π, π].
	M = math.Mod(M, 2*math.Pi)
	if M > math.Pi {
		M -= 2 * math.Pi
	} else if M < -math.Pi {
		M += 2 * math.Pi
	}

	// Newton-Raphson iteration for eccentric anomaly.
	E := M // initial guess
	if e > 0.8 {
		E = math.Pi // better starting point for high eccentricity
		if M > 0 {
			E = math.Pi
		} else {
			E = -math.Pi
		}
	}

	for iter := 0; iter < 50; iter++ {
		sinE, cosE := math.Sincos(E)
		f := E - e*sinE - M
		fp := 1.0 - e*cosE
		dE := -f / fp
		E += dE
		if math.Abs(dE) < 1e-15 {
			break
		}
	}

	// True anomaly from eccentric anomaly.
	sinE, cosE := math.Sincos(E)
	nu = math.Atan2(math.Sqrt(1-e*e)*sinE, cosE-e)

	// Radius.
	r = o.a * (1.0 - e*cosE)
	return
}

// solveParabolic solves Barker's equation for a parabolic orbit (e = 1).
// M here is dt (days since periapsis). Returns true anomaly and radius.
func (o *Orbit) solveParabolic(dt float64) (nu, r float64) {
	// For parabolic orbit: r = q * (1 + tan²(ν/2)), and
	// Barker's equation: D + D³/3 = sqrt(2μ/q³) * dt
	// where D = tan(ν/2).
	q := o.q
	W := 3.0 * math.Sqrt(o.mu/(2.0*q*q*q)) * dt

	// Solve D³ + 3D - 3W = 0 using the real cubic root.
	// Substitution: D = 2*sqrt(1)*sinh(1/3 * arcsinh(3W/2))
	// Simplified: D = 2*sinh(arcsinh(W)/3) ... but let's use a standard approach.
	Y := math.Cbrt(W + math.Sqrt(W*W+1))
	D := Y - 1.0/Y

	nu = 2.0 * math.Atan(D)
	r = q * (1.0 + D*D)
	return
}

// solveHyperbolic solves the hyperbolic Kepler equation M = e*sinh(H) - H.
// M here is dt (days since periapsis). Returns true anomaly and radius.
func (o *Orbit) solveHyperbolic(dt float64) (nu, r float64) {
	e := o.e
	a := -o.q / (e - 1.0) // semi-major axis (negative for hyperbolic)

	// Mean anomaly for hyperbolic orbit.
	M := math.Sqrt(o.mu/(a*a*a)) * dt // note: a < 0 so a³ < 0, but -a³ > 0
	// Actually for hyperbolic: n = sqrt(mu / (-a)³), and M = n * dt
	absA := math.Abs(a)
	M = math.Sqrt(o.mu/(absA*absA*absA)) * dt

	// Newton-Raphson for hyperbolic anomaly H.
	H := M // initial guess
	for iter := 0; iter < 50; iter++ {
		sinhH := math.Sinh(H)
		coshH := math.Cosh(H)
		f := e*sinhH - H - M
		fp := e*coshH - 1.0
		dH := -f / fp
		H += dH
		if math.Abs(dH) < 1e-15 {
			break
		}
	}

	// True anomaly from hyperbolic anomaly.
	nu = 2.0 * math.Atan(math.Sqrt((e+1.0)/(e-1.0))*math.Tanh(H/2.0))

	// Radius.
	r = absA * (e*math.Cosh(H) - 1.0)
	return
}

// Package elements computes osculating Keplerian orbital elements from
// position and velocity state vectors.
//
// Based on the algorithms in Bate, Mueller & White, "Fundamentals of
// Astrodynamics" (1971), Section 2.4. Matches Skyfield's elementslib.py.
package elements

import "math"

const (
	twoPi     = 2 * math.Pi
	rad2deg   = 180.0 / math.Pi
	secPerDay = 86400.0
)

// OsculatingElements holds a complete set of Keplerian orbital elements.
type OsculatingElements struct {
	SemiMajorAxisKm      float64 // a — semi-major axis in km (Inf for parabolic)
	SemiMinorAxisKm      float64 // b — semi-minor axis in km
	SemiLatusRectumKm    float64 // p — semi-latus rectum in km
	Eccentricity         float64 // e — eccentricity (0=circular, <1=elliptic, 1=parabolic, >1=hyperbolic)
	InclinationDeg       float64 // i — inclination in degrees
	LongAscNodeDeg       float64 // Ω — longitude of ascending node in degrees
	ArgPeriapsisDeg      float64 // ω — argument of periapsis in degrees
	TrueAnomalyDeg       float64 // ν — true anomaly in degrees
	EccentricAnomalyDeg  float64 // E — eccentric anomaly in degrees (hyperbolic anomaly for e>1)
	MeanAnomalyDeg       float64 // M — mean anomaly in degrees
	MeanMotionDegPerDay  float64 // n — mean motion in degrees/day
	PeriapsisDistanceKm  float64 // q — periapsis distance in km
	ApoapsisDistanceKm   float64 // Q — apoapsis distance in km (Inf for e≥1)
	PeriodDays           float64 // P — orbital period in days (Inf for e≥1)
	TrueLongitudeDeg     float64 // l — true longitude (Ω + ω + ν) in degrees
	MeanLongitudeDeg     float64 // L — mean longitude (Ω + ω + M) in degrees
	LongPeriapsisDeg     float64 // ϖ — longitude of periapsis (Ω + ω) in degrees
	ArgLatitudeDeg       float64 // u — argument of latitude (ω + ν) in degrees
	PeriapsisTimeDays    float64 // time of periapsis relative to epoch (days)
}

// FromStateVector computes osculating Keplerian orbital elements from a
// position and velocity state vector.
//
// posKm is position in km, velKmPerSec is velocity in km/s.
// muKm3s2 is the gravitational parameter GM in km³/s² (e.g., 132712440041.94 for the Sun).
func FromStateVector(posKm, velKmPerSec [3]float64, muKm3s2 float64) OsculatingElements {
	r := length(posKm)
	v := length(velKmPerSec)

	// Specific angular momentum h = r × v
	hVec := cross(posKm, velKmPerSec)
	h := length(hVec)

	// Eccentricity vector e = ((v²-μ/r)r - (r·v)v) / μ
	rdv := dot(posKm, velKmPerSec)
	v2 := v * v
	factor := v2 - muKm3s2/r
	eVec := [3]float64{
		(factor*posKm[0] - rdv*velKmPerSec[0]) / muKm3s2,
		(factor*posKm[1] - rdv*velKmPerSec[1]) / muKm3s2,
		(factor*posKm[2] - rdv*velKmPerSec[2]) / muKm3s2,
	}
	e := length(eVec)

	// Node vector n = [-hy, hx, 0]
	nVec := [3]float64{-hVec[1], hVec[0], 0}
	n := length(nVec)

	// Semi-latus rectum
	p := h * h / muKm3s2

	// Inclination
	inc := math.Acos(clamp(hVec[2]/h, -1, 1))

	// Longitude of ascending node
	var omega float64
	if n > 1e-15 {
		omega = math.Atan2(hVec[0], -hVec[1])
		if omega < 0 {
			omega += twoPi
		}
	}

	// True anomaly
	nu := trueAnomaly(eVec, e, nVec, n, posKm, velKmPerSec, r, rdv)

	// Argument of periapsis
	w := argPeriapsis(eVec, e, nVec, n, posKm, velKmPerSec, hVec)

	// Semi-major axis
	var a float64
	e2 := e * e
	if math.Abs(e-1.0) < 1e-15 {
		a = math.Inf(1)
	} else {
		a = p / (1.0 - e2)
	}

	// Semi-minor axis
	var b float64
	if e < 1.0 {
		b = p / math.Sqrt(1.0-e2)
	} else if e > 1.0 {
		b = p * math.Sqrt(e2-1.0) / (1.0 - e2) // negative for hyperbolic, use abs
		if b < 0 {
			b = -b
		}
	}

	// Eccentric anomaly
	E := eccentricAnomaly(nu, e)

	// Mean anomaly
	M := meanAnomaly(E, e)

	// Mean motion (rad/s → deg/day)
	var nMot float64
	absA := math.Abs(a)
	if absA > 0 && !math.IsInf(absA, 0) {
		nMot = math.Sqrt(muKm3s2 / (absA * absA * absA)) // rad/s
	}

	// Periapsis/apoapsis distance
	var q, Q float64
	if math.Abs(e-1.0) < 1e-15 {
		q = p / 2.0
	} else {
		q = p * (1.0 - e) / (1.0 - e2)
	}
	if e < 1.0 {
		Q = p * (1.0 + e) / (1.0 - e2)
	} else {
		Q = math.Inf(1)
	}

	// Period
	var period float64
	if a > 0 && !math.IsInf(a, 0) {
		period = twoPi * math.Sqrt(a*a*a/muKm3s2) / secPerDay
	} else {
		period = math.Inf(1)
	}

	// Periapsis time
	var tPeri float64
	if nMot > 1e-20 {
		tPeri = M / nMot / secPerDay // days
	}

	// Composite angles
	trueLon := math.Mod(omega+w+nu+4*twoPi, twoPi)
	meanLon := math.Mod(omega+w+M+4*twoPi, twoPi)
	longPeri := math.Mod(omega+w+4*twoPi, twoPi)
	argLat := math.Mod(w+nu+4*twoPi, twoPi)

	return OsculatingElements{
		SemiMajorAxisKm:      a,
		SemiMinorAxisKm:      b,
		SemiLatusRectumKm:    p,
		Eccentricity:         e,
		InclinationDeg:       inc * rad2deg,
		LongAscNodeDeg:       omega * rad2deg,
		ArgPeriapsisDeg:      w * rad2deg,
		TrueAnomalyDeg:       nu * rad2deg,
		EccentricAnomalyDeg:  E * rad2deg,
		MeanAnomalyDeg:       M * rad2deg,
		MeanMotionDegPerDay:  nMot * rad2deg * secPerDay,
		PeriapsisDistanceKm:  q,
		ApoapsisDistanceKm:   Q,
		PeriodDays:           period,
		TrueLongitudeDeg:     trueLon * rad2deg,
		MeanLongitudeDeg:     meanLon * rad2deg,
		LongPeriapsisDeg:     longPeri * rad2deg,
		ArgLatitudeDeg:       argLat * rad2deg,
		PeriapsisTimeDays:    tPeri,
	}
}

func trueAnomaly(eVec [3]float64, e float64, nVec [3]float64, n float64, pos, vel [3]float64, r, rdv float64) float64 {
	if e > 1e-15 {
		// Non-circular: angle between eccentricity vector and position
		nu := angleBetween(eVec, pos)
		if rdv < 0 {
			nu = twoPi - nu
		}
		if e > 1.0-1e-15 {
			// Hyperbolic: normalize to [-π, π]
			nu = normPi(nu)
		}
		return nu
	}
	if n < 1e-15 {
		// Circular equatorial
		nu := math.Acos(clamp(pos[0]/r, -1, 1))
		if vel[0] > 0 {
			nu = twoPi - nu
		}
		return nu
	}
	// Circular non-equatorial
	nu := angleBetween(nVec, pos)
	if pos[2] < 0 {
		nu = twoPi - nu
	}
	return nu
}

func argPeriapsis(eVec [3]float64, e float64, nVec [3]float64, n float64, pos, vel, hVec [3]float64) float64 {
	if e < 1e-15 {
		return 0 // circular orbit: ω undefined, set to 0
	}
	if n > 1e-15 {
		// Non-equatorial
		w := angleBetween(nVec, eVec)
		if eVec[2] < 0 {
			w = twoPi - w
		}
		return w
	}
	// Equatorial
	w := math.Atan2(eVec[1], eVec[0])
	if w < 0 {
		w += twoPi
	}
	// Check prograde/retrograde
	crossRV := cross(pos, vel)
	if crossRV[2] < 0 {
		w = twoPi - w
	}
	return w
}

func eccentricAnomaly(nu, e float64) float64 {
	if e < 1.0 {
		E := 2.0 * math.Atan(math.Sqrt((1.0-e)/(1.0+e))*math.Tan(nu/2.0))
		if E < 0 {
			E += twoPi
		}
		return E
	}
	if e > 1.0 {
		// Hyperbolic anomaly
		tanNu2 := math.Tan(nu / 2.0)
		ratio := tanNu2 / math.Sqrt((e+1.0)/(e-1.0))
		E := 2.0 * math.Atanh(ratio)
		return normPi(E)
	}
	return 0 // parabolic
}

func meanAnomaly(E, e float64) float64 {
	if e < 1.0 {
		M := E - e*math.Sin(E)
		M = math.Mod(M+twoPi, twoPi)
		return M
	}
	if e > 1.0 {
		M := e*math.Sinh(E) - E
		return normPi(M)
	}
	return 0
}

func angleBetween(u, v [3]float64) float64 {
	uMag := length(u)
	vMag := length(v)
	if uMag == 0 || vMag == 0 {
		return 0
	}
	// Kahan's numerically stable formula
	a := [3]float64{u[0] * vMag, u[1] * vMag, u[2] * vMag}
	b := [3]float64{v[0] * uMag, v[1] * uMag, v[2] * uMag}
	diff := [3]float64{a[0] - b[0], a[1] - b[1], a[2] - b[2]}
	sum := [3]float64{a[0] + b[0], a[1] + b[1], a[2] + b[2]}
	return 2.0 * math.Atan2(length(diff), length(sum))
}

func normPi(angle float64) float64 {
	a := math.Mod(angle+math.Pi, twoPi)
	if a < 0 {
		a += twoPi
	}
	return a - math.Pi
}

func cross(a, b [3]float64) [3]float64 {
	return [3]float64{
		a[1]*b[2] - a[2]*b[1],
		a[2]*b[0] - a[0]*b[2],
		a[0]*b[1] - a[1]*b[0],
	}
}

func dot(a, b [3]float64) float64 {
	return a[0]*b[0] + a[1]*b[1] + a[2]*b[2]
}

func length(a [3]float64) float64 {
	return math.Sqrt(a[0]*a[0] + a[1]*a[1] + a[2]*a[2])
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

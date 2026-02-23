// Package magnitude computes visual apparent magnitudes for planets using the
// Mallama & Hilton (2018) phase curves. Matches Skyfield's magnitudelib.py.
package magnitude

import "math"

const (
	rad2deg = 180.0 / math.Pi
	deg2rad = math.Pi / 180.0
)

// Saturn's ICRF pole direction (RA=40.589°, Dec=83.537°).
var saturnPole = [3]float64{0.08547883, 0.07323576, 0.99364475}

// Uranus's ICRF pole direction (RA=257.311°, Dec=-15.175°).
var uranusPole = [3]float64{-0.21199958, -0.94155916, -0.26176809}

// PlanetaryMagnitude computes the visual apparent magnitude of a planet.
//
// bodyID is the NAIF body ID (1-8 for barycenters, or 199/299/.../899 for bodies).
// phaseAngleDeg is the Sun-planet-observer angle in degrees.
// rAU is the planet's distance from the Sun in AU.
// deltaAU is the planet's distance from the observer in AU.
//
// For Saturn (6/699) and Uranus (7/799), the sub-solar and sub-observer latitudes
// are needed. Use PlanetaryMagnitudeWithGeometry for those planets.
//
// Returns NaN for unsupported body IDs.
func PlanetaryMagnitude(bodyID int, phaseAngleDeg, rAU, deltaAU float64) float64 {
	switch normalizeBodyID(bodyID) {
	case 1:
		return mercury(rAU, deltaAU, phaseAngleDeg)
	case 2:
		return venus(rAU, deltaAU, phaseAngleDeg)
	case 3:
		return earth(rAU, deltaAU, phaseAngleDeg)
	case 4:
		return mars(rAU, deltaAU, phaseAngleDeg)
	case 5:
		return jupiter(rAU, deltaAU, phaseAngleDeg)
	case 6:
		return saturn(rAU, deltaAU, phaseAngleDeg, 0, 0, true)
	case 7:
		return uranus(rAU, deltaAU, phaseAngleDeg, 0, 0)
	case 8:
		return neptune(rAU, deltaAU, phaseAngleDeg, 2020.0)
	}
	return math.NaN()
}

// PlanetaryMagnitudeWithGeometry computes visual magnitude with full geometry.
//
// sunToPlanetAU is the Sun-to-planet vector in AU (ICRF).
// obsToPlanetAU is the observer-to-planet vector in AU (ICRF).
// year is the decimal year (needed for Neptune's temporal variation).
//
// This function computes the phase angle, distances, and sub-latitudes
// (for Saturn and Uranus) from the input vectors.
func PlanetaryMagnitudeWithGeometry(bodyID int, sunToPlanetAU, obsToPlanetAU [3]float64, year float64) float64 {
	r := vecLength(sunToPlanetAU)
	delta := vecLength(obsToPlanetAU)
	phaseAngle := angleBetween(sunToPlanetAU, obsToPlanetAU) * rad2deg

	bid := normalizeBodyID(bodyID)
	switch bid {
	case 1:
		return mercury(r, delta, phaseAngle)
	case 2:
		return venus(r, delta, phaseAngle)
	case 3:
		return earth(r, delta, phaseAngle)
	case 4:
		return mars(r, delta, phaseAngle)
	case 5:
		return jupiter(r, delta, phaseAngle)
	case 6:
		sunSubLat := subLatitude(saturnPole, sunToPlanetAU)
		earthSubLat := subLatitude(saturnPole, obsToPlanetAU)
		return saturn(r, delta, phaseAngle, sunSubLat, earthSubLat, true)
	case 7:
		sunSubLat := subLatitude(uranusPole, sunToPlanetAU)
		earthSubLat := subLatitude(uranusPole, obsToPlanetAU)
		return uranus(r, delta, phaseAngle, sunSubLat, earthSubLat)
	case 8:
		return neptune(r, delta, phaseAngle, year)
	}
	return math.NaN()
}

// SaturnPole returns Saturn's ICRF pole unit vector for sub-latitude computation.
func SaturnPole() [3]float64 { return saturnPole }

// UranusPole returns Uranus's ICRF pole unit vector for sub-latitude computation.
func UranusPole() [3]float64 { return uranusPole }

func normalizeBodyID(id int) int {
	switch id {
	case 199, 1:
		return 1
	case 299, 2:
		return 2
	case 399, 3:
		return 3
	case 499, 4:
		return 4
	case 599, 5:
		return 5
	case 699, 6:
		return 6
	case 799, 7:
		return 7
	case 899, 8:
		return 8
	}
	return 0
}

// mercury — Mallama & Hilton Equation #2.
func mercury(r, delta, phi float64) float64 {
	dm := 5 * math.Log10(r*delta)
	pf := phi * (6.3280e-02 + phi*(-1.6336e-03+phi*(3.3644e-05+
		phi*(-3.4265e-07+phi*(1.6893e-09+phi*(-3.0334e-12))))))
	return -0.613 + dm + pf
}

// venus — Mallama & Hilton Equations #3 and #4.
func venus(r, delta, phi float64) float64 {
	dm := 5 * math.Log10(r*delta)
	var pf float64
	if phi < 163.7 {
		pf = phi * (-1.044e-03 + phi*(3.687e-04+phi*(-2.814e-06+phi*8.938e-09)))
	} else {
		pf = (236.05828+4.384) + phi*(-2.81914e+00+phi*8.39034e-03)
	}
	return -4.384 + dm + pf
}

// earth — Mallama & Hilton Equation #5.
func earth(r, delta, phi float64) float64 {
	dm := 5 * math.Log10(r*delta)
	return -3.99 + dm + phi*(-1.060e-03+phi*2.054e-04)
}

// mars — Mallama & Hilton Equations #6 and #7.
func mars(r, delta, phi float64) float64 {
	dm := 5 * math.Log10(r*delta)
	var base, pf float64
	if phi <= 50.0 {
		base = -1.601
		pf = phi * (2.267e-02 + phi*(-1.302e-04))
	} else {
		base = -0.367
		pf = phi * (-0.02573 + phi*3.445e-04)
	}
	return base + dm + pf
}

// jupiter — Mallama & Hilton Equations #8 and #9.
func jupiter(r, delta, phi float64) float64 {
	dm := 5 * math.Log10(r*delta)
	if phi <= 12.0 {
		return -9.395 + dm + phi*(6.16e-04*phi-3.7e-04)
	}
	pp := phi / 180.0
	poly := ((((-1.876*pp+2.809)*pp-0.062)*pp-0.363)*pp-1.507)*pp + 1.0
	return -9.428 + dm - 2.5*math.Log10(poly)
}

// saturn — Mallama & Hilton Equations #10, #11, #12.
func saturn(r, delta, phi, sunSubLat, earthSubLat float64, rings bool) float64 {
	dm := 5 * math.Log10(r*delta)

	product := sunSubLat * earthSubLat
	var subLatGeoc float64
	if product >= 0 {
		subLatGeoc = math.Sqrt(product)
	}

	if phi <= 6.5 && subLatGeoc <= 27.0 {
		if rings {
			sinSL := math.Sin(subLatGeoc * deg2rad)
			return -8.914 + dm - 1.825*sinSL + 0.026*phi -
				0.378*sinSL*math.Exp(-2.25*phi)
		}
		return -8.95 + dm + phi*(-3.7e-04+phi*6.16e-04)
	}

	if phi > 6.5 && !rings {
		return -8.94 + dm + phi*(2.446e-04+phi*(2.672e-04+phi*(-1.506e-06+phi*4.767e-09)))
	}

	return math.NaN()
}

// uranus — Mallama & Hilton Equations #14 and #15.
func uranus(r, delta, phi, sunSubLat, earthSubLat float64) float64 {
	dm := 5 * math.Log10(r*delta)
	subLat := (math.Abs(sunSubLat) + math.Abs(earthSubLat)) / 2.0
	mag := -7.110 + dm + (-0.00084 * subLat)
	if phi > 3.1 {
		mag += phi * (1.045e-4*phi + 6.587e-3)
	}
	return mag
}

// neptune — Mallama & Hilton Equations #16 and #17.
func neptune(r, delta, phi, year float64) float64 {
	dm := 5 * math.Log10(r*delta)
	base := -6.89 - 0.0054*(year-1980.0)
	if base < -7.00 {
		base = -7.00
	}
	if base > -6.89 {
		base = -6.89
	}
	mag := base + dm
	if phi > 1.9 {
		if year >= 2000.0 {
			mag += phi * (7.944e-3 + phi*9.617e-5)
		} else {
			return math.NaN()
		}
	}
	return mag
}

// subLatitude computes the sub-observer latitude on a planet given the planet's
// pole unit vector and the observer-to-planet direction vector.
func subLatitude(pole, direction [3]float64) float64 {
	a := angleBetween(pole, direction)
	return a*rad2deg - 90.0
}

func angleBetween(u, v [3]float64) float64 {
	uMag := vecLength(u)
	vMag := vecLength(v)
	if uMag == 0 || vMag == 0 {
		return 0
	}
	a := [3]float64{u[0] * vMag, u[1] * vMag, u[2] * vMag}
	b := [3]float64{v[0] * uMag, v[1] * uMag, v[2] * uMag}
	diff := [3]float64{a[0] - b[0], a[1] - b[1], a[2] - b[2]}
	sum := [3]float64{a[0] + b[0], a[1] + b[1], a[2] + b[2]}
	return 2.0 * math.Atan2(vecLength(diff), vecLength(sum))
}

func vecLength(v [3]float64) float64 {
	return math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
}

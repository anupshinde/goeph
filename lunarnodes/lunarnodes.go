package lunarnodes

import (
	"math"

	"github.com/anupshinde/goeph/coord"
	"github.com/anupshinde/goeph/spk"
)

const (
	j2000JD = 2451545.0
	deg2rad = math.Pi / 180.0
)

// J2000 mean obliquity ecliptic pole in ICRF: (0, -sin(ε), cos(ε))
const (
	eclipticPoleSinE = 0.3977771559319137062
	eclipticPoleCosE = 0.9174820620691818140
)

// MeanLunarNodes returns the mean North and South node ecliptic longitudes
// (degrees) for the given TDB Julian date. Uses Meeus formula.
// Note: This is not derived from Skyfield — it was added independently.
func MeanLunarNodes(tdbJD float64) (northLon, southLon float64) {
	T := (tdbJD - j2000JD) / 36525.0

	omega := 125.04452 - 1934.136261*T + 0.0020708*T*T + T*T*T/450000.0

	northLon = math.Mod(omega, 360.0)
	if northLon < 0 {
		northLon += 360.0
	}
	southLon = math.Mod(northLon+180.0, 360.0)
	return
}

// TrueNode returns the ecliptic J2000 longitude (degrees) of the Moon's
// instantaneous ascending and descending nodes, computed from the Moon's
// geocentric state vectors (position and velocity) using the loaded SPK ephemeris.
func TrueNode(eph *spk.SPK, tdbJD float64) (northLon, southLon float64) {
	r := eph.GeocentricPosition(spk.Moon, tdbJD)
	v := eph.GeocentricVelocity(spk.Moon, tdbJD)

	// Orbit normal: h = r × v (perpendicular to orbital plane)
	h := cross3(r, v)
	hu := unit3(h)

	// Ecliptic pole in ICRF: k = (0, -sin(ε), cos(ε))
	k := [3]float64{0, -eclipticPoleSinE, eclipticPoleCosE}

	// Node line direction: n = k × h (intersection of orbital and ecliptic planes)
	n := cross3(k, hu)

	// Convert node direction to ecliptic longitude
	_, northLon = coord.ICRFToEcliptic(n[0], n[1], n[2])
	southLon = math.Mod(northLon+180.0, 360.0)
	return
}

// MeanLunarNodeICRF returns unit ICRF direction vectors for the mean North
// and South lunar nodes at the given TDB Julian date. Uses the same Meeus
// formula as MeanLunarNodes, converted from J2000 ecliptic longitude to ICRF.
func MeanLunarNodeICRF(tdbJD float64) (north, south [3]float64) {
	northLon, _ := MeanLunarNodes(tdbJD)
	lonRad := northLon * deg2rad

	// Unit vector in J2000 ecliptic: (cos λ, sin λ, 0)
	ex, ey := math.Cos(lonRad), math.Sin(lonRad)

	// Ecliptic → ICRF: rotate around X by -ε (inverse of ICRF→ecliptic Rx(+ε))
	north = [3]float64{ex, eclipticPoleCosE * ey, eclipticPoleSinE * ey}
	south = [3]float64{-north[0], -north[1], -north[2]}
	return
}

// TrueNodeICRF returns unit ICRF direction vectors for the true (instantaneous)
// North and South lunar nodes at the given TDB Julian date, computed from the
// Moon's geocentric state vectors via the loaded SPK ephemeris.
func TrueNodeICRF(eph *spk.SPK, tdbJD float64) (north, south [3]float64) {
	r := eph.GeocentricPosition(spk.Moon, tdbJD)
	v := eph.GeocentricVelocity(spk.Moon, tdbJD)

	h := cross3(r, v)
	hu := unit3(h)

	k := [3]float64{0, -eclipticPoleSinE, eclipticPoleCosE}
	n := cross3(k, hu)
	north = unit3(n)
	south = [3]float64{-north[0], -north[1], -north[2]}
	return
}

func cross3(a, b [3]float64) [3]float64 {
	return [3]float64{
		a[1]*b[2] - a[2]*b[1],
		a[2]*b[0] - a[0]*b[2],
		a[0]*b[1] - a[1]*b[0],
	}
}

func unit3(v [3]float64) [3]float64 {
	n := math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
	return [3]float64{v[0] / n, v[1] / n, v[2] / n}
}

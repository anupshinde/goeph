package lunarnodes

import "math"

const j2000JD = 2451545.0

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

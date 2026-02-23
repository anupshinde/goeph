package coord

import "math"

const cKmPerDay = 299792.458 * 86400.0 // speed of light in km/day

// Aberration applies special-relativistic stellar aberration to an astrometric
// position vector. Uses the full Lorentz transformation (not the classical v/c
// approximation). Matches Skyfield's add_aberration() in relativity.py.
//
// position is the observer-to-target vector in km (astrometric position).
// velocity is the observer's barycentric velocity in km/day.
// lightTime is the light travel time to the target in days.
//
// Returns the apparent position in km.
func Aberration(position, velocity [3]float64, lightTime float64) [3]float64 {
	p1mag := lightTime * cKmPerDay // distance in km
	vemag := length3(velocity)     // observer speed in km/day

	if p1mag == 0 || vemag == 0 {
		return position
	}

	beta := vemag / cKmPerDay            // v/c ratio
	dot := dot3(position, velocity)       // km * km/day
	cosd := dot / (p1mag * vemag)         // cosine of angle between position and velocity
	gammai := math.Sqrt(1.0 - beta*beta)  // inverse Lorentz factor (sqrt(1 - v²/c²))
	p := beta * cosd                      // dimensionless
	q := (1.0 + p/(1.0+gammai)) * lightTime // days
	r := 1.0 + p                          // dimensionless

	var result [3]float64
	for i := 0; i < 3; i++ {
		result[i] = (gammai*position[i] + q*velocity[i]) / r
	}
	return result
}

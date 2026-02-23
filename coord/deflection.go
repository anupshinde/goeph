package coord

import "math"

const (
	// Heliocentric gravitational constant GM_sun in m^3/s^2 (IAU 2012)
	gs = 1.32712440017987e20
	// Speed of light in m/s
	cMPerSec = 299792458.0
)

// Deflection computes the gravitational deflection of light by a single body.
// Returns the deflection correction vector in km (to be added to the position).
//
// position is the observer-to-target vector in km (astrometric position).
// pe is the observer-to-deflector vector in km (same sign convention as Skyfield).
// rmass is the reciprocal mass: GM_sun / GM_deflector (1.0 for the Sun).
//
// Matches Skyfield's _compute_deflection() in relativity.py.
func Deflection(position, pe [3]float64, rmass float64) [3]float64 {
	// Vector from deflector to target
	pq := add3(position, pe)

	pmag := length3(position)
	qmag := length3(pq)
	emag := length3(pe)

	if pmag == 0 || qmag == 0 || emag == 0 {
		return [3]float64{}
	}

	// Unit vectors
	phat := scale3(1.0/pmag, position)
	qhat := scale3(1.0/qmag, pq)
	ehat := scale3(1.0/emag, pe)

	// Dot products
	pdotq := dot3(phat, qhat)
	qdote := dot3(qhat, ehat)
	edotp := dot3(ehat, phat)

	// If deflector is on the line toward or away from the target (within ~1 arcsec),
	// skip deflection to avoid numerical issues.
	if math.Abs(edotp) > 0.99999999999 {
		return [3]float64{}
	}

	// Scale factor: 2*GM_sun / (c^2 * distance_to_deflector_in_meters * reciprocal_mass)
	fac1 := 2.0 * gs / (cMPerSec * cMPerSec * emag * 1000.0 * rmass)
	fac2 := 1.0 + qdote

	var d [3]float64
	for i := 0; i < 3; i++ {
		d[i] = fac1 * (pdotq*ehat[i] - edotp*qhat[i]) / fac2 * pmag
	}
	return d
}

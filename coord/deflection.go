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
// deflectorToObserver runs from the deflecting body to the observer, in km,
// taken at the time the light ray passed closest to the deflector. Note the
// direction: the expression below needs both of the deflector's own vectors,
// to the observer and to the target, and recovers the second as
// position + deflectorToObserver.
// rmass is the reciprocal mass: GM_sun / GM_deflector (1.0 for the Sun).
//
// For a target far behind the deflector the result is the classical
// deflection 2·GM/(c²·E) · (1 + cos χ)/sin χ, where E is the observer's
// distance from the deflector and χ the angle between them as seen by the
// observer — 1.75″ for a ray grazing the Sun's limb. The deflection always
// moves the apparent position away from the deflector.
func Deflection(position, deflectorToObserver [3]float64, rmass float64) [3]float64 {
	pe := deflectorToObserver
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
	// skip deflection to avoid numerical issues. A ray through the deflector's
	// centre sends 1 + qdote to zero, which is the singularity this guards.
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

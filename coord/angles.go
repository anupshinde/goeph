package coord

import "math"

// SeparationAngle returns the angular separation in degrees between two
// Cartesian vectors. Uses Kahan's numerically stable formula.
// See: https://people.eecs.berkeley.edu/~wkahan/Mindless.pdf Section 12.
func SeparationAngle(a, b [3]float64) float64 {
	lenA := math.Sqrt(a[0]*a[0] + a[1]*a[1] + a[2]*a[2])
	lenB := math.Sqrt(b[0]*b[0] + b[1]*b[1] + b[2]*b[2])
	if lenA == 0 || lenB == 0 {
		return 0
	}

	// u = a * |b|, v = b * |a|
	var diffSq, sumSq float64
	for i := 0; i < 3; i++ {
		u := a[i] * lenB
		v := b[i] * lenA
		d := u - v
		s := u + v
		diffSq += d * d
		sumSq += s * s
	}

	return 2.0 * math.Atan2(math.Sqrt(diffSq), math.Sqrt(sumSq)) * rad2deg
}

// PhaseAngle returns the phase angle in degrees given two direction vectors:
// obsToTarget (observer to target) and sunToTarget (Sun to target).
// The phase angle is 0° when fully illuminated and 180° when fully in shadow.
//
// To compute sunToTarget from goeph's SPK package:
//
//	obsToTarget := spk.Observe(targetID, tdb)
//	obsToSun := spk.Observe(spk.Sun, tdb)
//	sunToTarget := [3]float64{obsToTarget[0]-obsToSun[0], obsToTarget[1]-obsToSun[1], obsToTarget[2]-obsToSun[2]}
func PhaseAngle(obsToTarget, sunToTarget [3]float64) float64 {
	return SeparationAngle(obsToTarget, sunToTarget)
}

// FractionIlluminated returns the fraction of a spherical body's disc that
// is illuminated, given the phase angle in degrees. Returns a value in [0, 1].
func FractionIlluminated(phaseAngleDeg float64) float64 {
	return 0.5 * (1.0 + math.Cos(phaseAngleDeg*deg2rad))
}

// PositionAngle returns the position angle from one sky position to another,
// measured North through East (counterclockwise on the sky), in degrees [0, 360).
// Both positions are given as RA (hours) and Dec (degrees).
func PositionAngle(ra1Hours, dec1Deg, ra2Hours, dec2Deg float64) float64 {
	ra1 := ra1Hours * 15.0 * deg2rad
	dec1 := dec1Deg * deg2rad
	ra2 := ra2Hours * 15.0 * deg2rad
	dec2 := dec2Deg * deg2rad

	dRA := ra2 - ra1
	pa := math.Atan2(math.Sin(dRA),
		math.Cos(dec1)*math.Tan(dec2)-math.Sin(dec1)*math.Cos(dRA))
	pa = math.Mod(pa, 2*math.Pi)
	if pa < 0 {
		pa += 2 * math.Pi
	}
	result := pa * rad2deg
	if result >= 360.0 {
		result -= 360.0
	}
	return result
}

// Elongation returns the elongation of a target from a reference body,
// given their ecliptic longitudes in degrees. Returns degrees in [0, 360).
// For moon phase, pass the Moon's ecliptic longitude as target and the
// Sun's as reference: 0°=new moon, 90°=first quarter, 180°=full, 270°=last quarter.
func Elongation(targetLonDeg, referenceLonDeg float64) float64 {
	e := math.Mod(targetLonDeg-referenceLonDeg, 360.0)
	if e < 0 {
		e += 360.0
	}
	return e
}

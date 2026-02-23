package coord

import "math"

// Refraction returns the atmospheric refraction correction in degrees for a
// given apparent (observed) altitude. Uses Bennett's formula (1982).
// Returns 0 for altitudes below -1° or above 89.9°.
//
// Parameters:
//   - altDeg: apparent altitude in degrees
//   - tempC: temperature in degrees Celsius
//   - pressureMbar: atmospheric pressure in millibars
func Refraction(altDeg, tempC, pressureMbar float64) float64 {
	if altDeg < -1.0 || altDeg > 89.9 {
		return 0
	}
	r := 0.016667 / math.Tan((altDeg+7.31/(altDeg+4.4))*deg2rad)
	return r * (0.28 * pressureMbar / (tempC + 273.0))
}

// Refract returns the apparent altitude in degrees after atmospheric
// refraction, given a true (geometric) altitude. Uses iterative convergence
// of Bennett's formula (within 3e-5 degrees, ~0.1 arcsecond).
//
// Parameters:
//   - altDeg: true (geometric) altitude in degrees
//   - tempC: temperature in degrees Celsius
//   - pressureMbar: atmospheric pressure in millibars
func Refract(altDeg, tempC, pressureMbar float64) float64 {
	alt := altDeg
	for i := 0; i < 20; i++ {
		prev := alt
		alt = altDeg + Refraction(alt, tempC, pressureMbar)
		if math.Abs(alt-prev) < 3e-5 {
			break
		}
	}
	return alt
}

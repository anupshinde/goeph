package star

import "github.com/anupshinde/goeph/coord"

// Galactic Center (Sgr A*) J2000 coordinates:
//
//	RA  = 17h 45m 40.0409s
//	Dec = -29° 00' 28.118"
var gcRAHours = 17.0 + 45.0/60.0 + 40.0409/3600.0
var gcDecDeg = -(29.0 + 0.0/60.0 + 28.118/3600.0)

// GalacticCenterICRF returns the ICRF unit vector for the Galactic Center.
func GalacticCenterICRF() (x, y, z float64) {
	return coord.RADecToICRF(gcRAHours, gcDecDeg)
}

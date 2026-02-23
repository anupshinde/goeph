package coord

import "math"

// GalacticMatrix is the rotation matrix from ICRF (J2000) to Galactic
// System II (IAU 1958). Apply as v_gal = GalacticMatrix * v_icrf.
// Source: SPICE Toolkit / Skyfield.
var GalacticMatrix = [3][3]float64{
	{-0.054875539395742523, -0.87343710472759606, -0.48383499177002515},
	{0.49410945362774389, -0.44482959429757496, 0.74698224869989183},
	{-0.86766613568337381, -0.19807638961301985, 0.45598379452141991},
}

// B1950Matrix is the rotation matrix from ICRF (J2000) to the mean equator
// and equinox of B1950 (FK4). Apply as v_B1950 = B1950Matrix * v_icrf.
// Source: SPICE Toolkit / Skyfield.
var B1950Matrix = [3][3]float64{
	{0.99992570795236291, 0.011178938126427691, 0.0048590038414544293},
	{-0.011178938137770135, 0.9999375133499887, -2.715792625851078e-05},
	{-0.0048590038153592712, -2.7162594714247048e-05, 0.9999881946023742},
}

// ICRSToJ2000Matrix is the frame bias matrix from ICRS to the dynamical
// mean equator and equinox of J2000. The bias is a few milliarcseconds.
// Source: IERS Conventions 2003, Chapter 5.
var ICRSToJ2000Matrix [3][3]float64

func init() {
	const asec2rad = deg2rad / 3600.0

	// ICRS frame biases in arcseconds
	xi0 := -0.0166170 * asec2rad
	eta0 := -0.0068192 * asec2rad
	da0 := -0.01460 * asec2rad

	yx := -da0
	zx := xi0
	xy := da0
	zy := eta0
	xz := -xi0
	yz := -eta0

	// Second-order diagonal corrections
	xx := 1.0 - 0.5*(yx*yx+zx*zx)
	yy := 1.0 - 0.5*(yx*yx+zy*zy)
	zz := 1.0 - 0.5*(zy*zy+zx*zx)

	ICRSToJ2000Matrix = [3][3]float64{
		{xx, xy, xz},
		{yx, yy, yz},
		{zx, zy, zz},
	}
}

// ICRFToGalactic converts an ICRF Cartesian vector to Galactic latitude and
// longitude in degrees. Longitude is in [0, 360).
func ICRFToGalactic(x, y, z float64) (latDeg, lonDeg float64) {
	gx := GalacticMatrix[0][0]*x + GalacticMatrix[0][1]*y + GalacticMatrix[0][2]*z
	gy := GalacticMatrix[1][0]*x + GalacticMatrix[1][1]*y + GalacticMatrix[1][2]*z
	gz := GalacticMatrix[2][0]*x + GalacticMatrix[2][1]*y + GalacticMatrix[2][2]*z

	r := math.Sqrt(gx*gx + gy*gy + gz*gz)
	if r == 0 {
		return 0, 0
	}

	latDeg = math.Asin(gz/r) * rad2deg
	lonDeg = math.Atan2(gy, gx) * rad2deg
	lonDeg = math.Mod(lonDeg+360.0, 360.0)
	return latDeg, lonDeg
}

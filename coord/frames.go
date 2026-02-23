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

// InertialFrame is a static (time-independent) reference frame defined by a
// rotation matrix from ICRF. Examples include the Galactic frame, B1950, and
// the ecliptic. Apply the rotation matrix to an ICRF vector to get coordinates
// in this frame.
type InertialFrame struct {
	Name   string
	Matrix [3][3]float64 // ICRF → frame rotation matrix
}

// XYZ applies the frame rotation to an ICRF position vector, returning
// Cartesian coordinates in this frame.
func (f InertialFrame) XYZ(posICRF [3]float64) [3]float64 {
	m := f.Matrix
	return [3]float64{
		m[0][0]*posICRF[0] + m[0][1]*posICRF[1] + m[0][2]*posICRF[2],
		m[1][0]*posICRF[0] + m[1][1]*posICRF[1] + m[1][2]*posICRF[2],
		m[2][0]*posICRF[0] + m[2][1]*posICRF[1] + m[2][2]*posICRF[2],
	}
}

// LatLon applies the frame rotation to an ICRF position vector, returning
// latitude and longitude in degrees. Longitude is in [0, 360).
func (f InertialFrame) LatLon(posICRF [3]float64) (latDeg, lonDeg float64) {
	v := f.XYZ(posICRF)
	r := math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
	if r == 0 {
		return 0, 0
	}
	latDeg = math.Asin(v[2]/r) * rad2deg
	lonDeg = math.Mod(math.Atan2(v[1], v[0])*rad2deg+360.0, 360.0)
	return
}

// TimeBasedFrame is a time-dependent reference frame defined by a function
// that returns a rotation matrix from ICRF at a given Julian date.
// Examples include the true equator of date and the ITRF (Earth-fixed frame).
type TimeBasedFrame struct {
	Name     string
	MatrixAt func(jd float64) [3][3]float64 // ICRF → frame rotation matrix at time jd
}

// XYZ applies the frame rotation at the given Julian date to an ICRF position
// vector, returning Cartesian coordinates in this frame.
func (f TimeBasedFrame) XYZ(posICRF [3]float64, jd float64) [3]float64 {
	m := f.MatrixAt(jd)
	return [3]float64{
		m[0][0]*posICRF[0] + m[0][1]*posICRF[1] + m[0][2]*posICRF[2],
		m[1][0]*posICRF[0] + m[1][1]*posICRF[1] + m[1][2]*posICRF[2],
		m[2][0]*posICRF[0] + m[2][1]*posICRF[1] + m[2][2]*posICRF[2],
	}
}

// LatLon applies the frame rotation at the given Julian date to an ICRF
// position vector, returning latitude and longitude in degrees. Longitude is
// in [0, 360).
func (f TimeBasedFrame) LatLon(posICRF [3]float64, jd float64) (latDeg, lonDeg float64) {
	v := f.XYZ(posICRF, jd)
	r := math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
	if r == 0 {
		return 0, 0
	}
	latDeg = math.Asin(v[2]/r) * rad2deg
	lonDeg = math.Mod(math.Atan2(v[1], v[0])*rad2deg+360.0, 360.0)
	return
}

// Predefined InertialFrame instances for common reference frames.
var (
	Galactic = InertialFrame{Name: "Galactic", Matrix: GalacticMatrix}
	B1950    = InertialFrame{Name: "B1950", Matrix: B1950Matrix}

	// Ecliptic is the J2000 mean ecliptic frame. This rotates ICRF around the
	// X-axis by the J2000 mean obliquity (23.4393°).
	Ecliptic = InertialFrame{
		Name: "Ecliptic",
		Matrix: [3][3]float64{
			{1, 0, 0},
			{0, obliquityCos, obliquitySin},
			{0, -obliquitySin, obliquityCos},
		},
	}
)

// ITRFFrame returns a TimeBasedFrame for the International Terrestrial
// Reference Frame (Earth-fixed). The jd argument is UT1 Julian date.
func ITRFFrame() TimeBasedFrame {
	return TimeBasedFrame{
		Name: "ITRF",
		MatrixAt: func(jdUT1 float64) [3][3]float64 {
			T := (jdUT1 - j2000JD) / 36525.0

			// Precession: J2000 → mean equator of date.
			// precessionMatrixInverse returns P^T (date→J2000). To go
			// J2000→date, we use its columns as rows.
			PT := precessionMatrixInverse(T)
			var P [3][3]float64
			for i := 0; i < 3; i++ {
				for j := 0; j < 3; j++ {
					P[i][j] = PT[j][i]
				}
			}

			// Nutation: mean → true equator of date.
			dpsiRad, depsRad := nutationAngles(T)
			epsM := meanObliquity(T)
			NT := nutationMatrixTranspose(dpsiRad, depsRad, epsM)
			var N [3][3]float64
			for i := 0; i < 3; i++ {
				for j := 0; j < 3; j++ {
					N[i][j] = NT[j][i]
				}
			}

			// NPB = N * P * B (ICRS → true equator of date via frame bias + precession + nutation)
			B := ICRSToJ2000Matrix
			var PB [3][3]float64
			for i := 0; i < 3; i++ {
				for j := 0; j < 3; j++ {
					for k := 0; k < 3; k++ {
						PB[i][j] += P[i][k] * B[k][j]
					}
				}
			}
			var NPB [3][3]float64
			for i := 0; i < 3; i++ {
				for j := 0; j < 3; j++ {
					for k := 0; k < 3; k++ {
						NPB[i][j] += N[i][k] * PB[k][j]
					}
				}
			}

			// Earth rotation: Rz(-GAST)
			gastRad := GAST(jdUT1) * deg2rad
			sinG, cosG := math.Sincos(gastRad)

			// R = Rz(-GAST) * NPB
			var R [3][3]float64
			for j := 0; j < 3; j++ {
				R[0][j] = cosG*NPB[0][j] + sinG*NPB[1][j]
				R[1][j] = -sinG*NPB[0][j] + cosG*NPB[1][j]
				R[2][j] = NPB[2][j]
			}
			return R
		},
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

package coord

import "math"

// Altaz converts a geocentric ICRF position vector to altitude and azimuth
// for a ground observer at the given geodetic latitude and longitude.
// jdUT1 is the UT1 Julian date (needed for Earth rotation).
//
// The position should be a geocentric or topocentric vector in km (typically
// from SPK.Apparent). For distant bodies (Sun, planets), geocentric and
// topocentric directions agree to <0.01°. For the Moon, topocentric positions
// are needed for arcsecond-level accuracy (parallax ~1°).
//
// Returns altitude (degrees, positive above horizon, geometric — no refraction),
// azimuth (degrees, 0=North, 90=East), and distance (km).
//
// The rotation chain is: ICRF → mean equator of date (precession) → true equator
// of date (nutation) → ITRF (Earth rotation via GAST) → local horizon (lat/lon).
// Matches Skyfield's rotation_at() + altaz() pipeline.
func Altaz(posICRF [3]float64, latDeg, lonDeg, jdUT1 float64) (altDeg, azDeg, distKm float64) {
	T := (jdUT1 - j2000JD) / 36525.0

	// Step 0: Frame bias — ICRS → J2000 dynamical (IERS Conventions 2003).
	B := &ICRSToJ2000Matrix
	var posJ2000 [3]float64
	for i := 0; i < 3; i++ {
		posJ2000[i] = B[i][0]*posICRF[0] + B[i][1]*posICRF[1] + B[i][2]*posICRF[2]
	}

	// Step 1: Precession — J2000 → mean equator of date.
	// precessionMatrixInverse returns P^T (date→J2000). We need P (J2000→date),
	// so we apply the transpose: (P·v)[i] = Σ_j P^T[j][i] · v[j].
	PT := precessionMatrixInverse(T)
	var pos [3]float64
	for i := 0; i < 3; i++ {
		pos[i] = PT[0][i]*posJ2000[0] + PT[1][i]*posJ2000[1] + PT[2][i]*posJ2000[2]
	}

	// Step 2: Nutation — mean equator → true equator of date.
	// nutationMatrixTranspose returns N^T (true→mean). We need N (mean→true),
	// so we apply the transpose: (N·v)[i] = Σ_j N^T[j][i] · v[j].
	dpsiRad, depsRad := nutationAngles(T)
	epsM := meanObliquity(T)
	NT := nutationMatrixTranspose(dpsiRad, depsRad, epsM)
	var posTr [3]float64
	for i := 0; i < 3; i++ {
		posTr[i] = NT[0][i]*pos[0] + NT[1][i]*pos[1] + NT[2][i]*pos[2]
	}

	// Step 3: Earth rotation — true equator of date → ITRF via Rz(-GAST).
	gastRad := GAST(jdUT1) * deg2rad
	sinG, cosG := math.Sincos(gastRad)
	xITRF := cosG*posTr[0] + sinG*posTr[1]
	yITRF := -sinG*posTr[0] + cosG*posTr[1]
	zITRF := posTr[2]

	// Step 4: Local horizon — ITRF → topocentric North-East-Up.
	// Rz(-lon) rotates the observer to the prime meridian,
	// then rot_y(lat) with reversed rows maps to N-E-U frame.
	// Matches Skyfield's mxm(rot_y(lat)[::-1], rot_z(-lon)).
	lat := latDeg * deg2rad
	lon := lonDeg * deg2rad
	sinLat, cosLat := math.Sincos(lat)
	sinLon, cosLon := math.Sincos(lon)

	// Rz(-lon)
	x1 := cosLon*xITRF + sinLon*yITRF
	y1 := -sinLon*xITRF + cosLon*yITRF
	z1 := zITRF

	// rot_y(lat) with reversed rows: [[-sinLat,0,cosLat],[0,1,0],[cosLat,0,sinLat]]
	xLocal := -sinLat*x1 + cosLat*z1
	yLocal := y1
	zLocal := cosLat*x1 + sinLat*z1

	// Step 5: Spherical coordinates.
	distKm = math.Sqrt(xLocal*xLocal + yLocal*yLocal + zLocal*zLocal)
	rXY := math.Sqrt(xLocal*xLocal + yLocal*yLocal)
	altDeg = math.Atan2(zLocal, rXY) * rad2deg
	azDeg = math.Mod(math.Atan2(yLocal, xLocal)*rad2deg+360.0, 360.0)

	return
}

// HourAngleDec computes the hour angle and declination of a geocentric ICRF
// position vector for an observer at the given longitude.
// jdUT1 is the UT1 Julian date.
//
// Hour angle is measured westward from the local meridian (0° = on meridian,
// positive = west of meridian). Declination is measured from the true equator
// of date.
//
// Returns hour angle (degrees, 0–360) and declination (degrees, -90 to +90).
func HourAngleDec(posICRF [3]float64, lonDeg, jdUT1 float64) (haDeg, decDeg float64) {
	T := (jdUT1 - j2000JD) / 36525.0

	// Frame bias: ICRS → J2000 dynamical
	B := &ICRSToJ2000Matrix
	var posJ2000 [3]float64
	for i := 0; i < 3; i++ {
		posJ2000[i] = B[i][0]*posICRF[0] + B[i][1]*posICRF[1] + B[i][2]*posICRF[2]
	}

	// Precession: J2000 → mean equator of date
	PT := precessionMatrixInverse(T)
	var pos [3]float64
	for i := 0; i < 3; i++ {
		pos[i] = PT[0][i]*posJ2000[0] + PT[1][i]*posJ2000[1] + PT[2][i]*posJ2000[2]
	}

	// Nutation: mean → true equator of date
	dpsiRad, depsRad := nutationAngles(T)
	epsM := meanObliquity(T)
	NT := nutationMatrixTranspose(dpsiRad, depsRad, epsM)
	var posTr [3]float64
	for i := 0; i < 3; i++ {
		posTr[i] = NT[0][i]*pos[0] + NT[1][i]*pos[1] + NT[2][i]*pos[2]
	}

	// RA and Dec in true equator of date
	r := math.Sqrt(posTr[0]*posTr[0] + posTr[1]*posTr[1] + posTr[2]*posTr[2])
	if r == 0 {
		return 0, 0
	}
	rXY := math.Sqrt(posTr[0]*posTr[0] + posTr[1]*posTr[1])
	decDeg = math.Atan2(posTr[2], rXY) * rad2deg
	raDeg := math.Mod(math.Atan2(posTr[1], posTr[0])*rad2deg+360.0, 360.0)

	// Hour angle = GAST + longitude - RA
	gastDeg := GAST(jdUT1)
	haDeg = math.Mod(gastDeg+lonDeg-raDeg+720.0, 360.0)

	return
}

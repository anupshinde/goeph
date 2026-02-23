package coord

import "math"

const (
	deg2rad    = math.Pi / 180.0
	rad2deg    = 180.0 / math.Pi
	arcsec2rad = deg2rad / 3600.0

	// J2000 mean obliquity: 84381.448 arcseconds (Lieske 1979, same as Skyfield)
	obliquitySin = 0.3977771559319137062
	obliquityCos = 0.9174820620691818140

	// WGS84 ellipsoid
	wgs84A  = 6378.137 // equatorial radius in km
	wgs84F  = 1.0 / 298.257223563
	wgs84E2 = wgs84F * (2.0 - wgs84F) // eccentricity squared

	j2000JD   = 2451545.0
	secPerDay = 86400.0

	// Conversion factor: 0.1 microarcseconds to radians
	tenthUas2Rad = arcsec2rad / 1e7
)

// Location represents a ground location with WGS84 coordinates.
type Location struct {
	Name string
	Lat  float64 // degrees, positive north
	Lon  float64 // degrees, positive east
}

// Full IAU 2000A nutation coefficient tables (678 luni-solar + 687 planetary terms)
// are in nutation_data.go, generated from Skyfield's nutation.npz.

// ICRFToEcliptic converts an ICRF Cartesian vector to ecliptic latitude and
// longitude (degrees). Uses the J2000 mean ecliptic (matching Skyfield's
// default ecliptic_latlon()).
func ICRFToEcliptic(x, y, z float64) (latDeg, lonDeg float64) {
	// Rotate around X-axis by obliquity: equatorial → ecliptic
	ex := x
	ey := obliquityCos*y + obliquitySin*z
	ez := -obliquitySin*y + obliquityCos*z

	r := math.Sqrt(ex*ex + ey*ey + ez*ez)
	if r == 0 {
		return 0, 0
	}

	latDeg = math.Asin(ez/r) * rad2deg
	lonDeg = math.Atan2(ey, ex) * rad2deg
	lonDeg = math.Mod(lonDeg+360.0, 360.0)
	return latDeg, lonDeg
}

// RADecToICRF converts J2000 RA (hours) and Dec (degrees) to an ICRF unit vector.
func RADecToICRF(raHours, decDeg float64) (x, y, z float64) {
	ra := raHours * 15.0 * deg2rad // hours → degrees → radians
	dec := decDeg * deg2rad
	cosDec := math.Cos(dec)
	x = cosDec * math.Cos(ra)
	y = cosDec * math.Sin(ra)
	z = math.Sin(dec)
	return
}

// EarthRotationAngle returns the Earth Rotation Angle in degrees for a given
// UT1 Julian date. Uses the formula from IAU Resolution B1.8 of 2000.
// This is the modern replacement for GMST.
func EarthRotationAngle(jdUT1 float64) float64 {
	th := 0.7790572732640 + 0.00273781191135448*(jdUT1-j2000JD)
	era := math.Mod(th, 1.0) + math.Mod(jdUT1, 1.0)
	era = math.Mod(era, 1.0)
	if era < 0 {
		era += 1.0
	}
	return era * 360.0
}

// GMST returns Greenwich Mean Sidereal Time in degrees for a given UT1 Julian date.
// Uses the IAU 1982 formula (Meeus).
func GMST(jdUT1 float64) float64 {
	du := jdUT1 - j2000JD
	T := du / 36525.0

	gmst := 280.46061837 + 360.98564736629*du +
		0.000387933*T*T - T*T*T/38710000.0

	return math.Mod(gmst, 360.0)
}

// fundamentalArgs computes the Delaunay arguments for the IAU 2000A nutation model.
// T is Julian centuries from J2000 TDB. Returns l, l', F, D, Ω in radians.
// From IERS Conventions 2003 Eq. 5.43 (Simon et al. 1994).
func fundamentalArgs(T float64) (l, lp, F, D, om float64) {
	l = (485868.249036 + T*(1717915923.2178+T*(31.8792+T*(0.051635-T*0.00024470)))) * arcsec2rad
	lp = (1287104.79305 + T*(129596581.0481+T*(-0.5532+T*(0.000136+T*0.00001149)))) * arcsec2rad
	F = (335779.526232 + T*(1739527262.8478+T*(-12.7512+T*(-0.001037+T*0.00000417)))) * arcsec2rad
	D = (1072260.70369 + T*(1602961601.2090+T*(-6.3706+T*(0.006593-T*0.00003169)))) * arcsec2rad
	om = (450160.398036 + T*(-6962890.5431+T*(7.4722+T*(0.007702-T*0.00005939)))) * arcsec2rad
	return
}

// meanObliquity returns the mean obliquity of the ecliptic at date, in radians.
// Uses the IAU 1980 formula (Lieske 1979).
func meanObliquity(T float64) float64 {
	return (84381.448 + T*(-46.8150+T*(-0.00059+T*0.001813))) * arcsec2rad
}

// nutationAngles computes nutation in longitude (dpsi) and obliquity (deps).
// T is Julian centuries from J2000 TDB.
// Returns dpsi and deps in radians.
// Dispatches to nutationAnglesStandard (30 terms) or nutationAnglesFull (1365 terms)
// based on the package-level NutationPrecision setting.
func nutationAngles(T float64) (dpsiRad, depsRad float64) {
	if nutationPrecision == NutationFull {
		return nutationAnglesFull(T)
	}
	return nutationAnglesStandard(T)
}

// nutationTerm holds one row of the IAU 2000A luni-solar nutation series.
// Units for s, sdot, cp, c, cdot, sp: 0.1 microarcseconds (0.1 uas).
type nutationTerm struct {
	nl, nlp, nf, nd, nom int     // integer multipliers for l, l', F, D, Ω
	s, sdot, cp          float64 // dpsi: (s + sdot*T)*sin(arg) + cp*cos(arg)
	c, cdot, sp          float64 // deps: (c + cdot*T)*cos(arg) + sp*sin(arg)
}

// Top 30 IAU 2000A luni-solar nutation terms by |s| amplitude.
// Source: Skyfield nutation.npz / IERS Conventions 2003 Table 5.3a.
var nutationTerms = []nutationTerm{
	// nl nlp  nf  nd nom          s       sdot        cp             c      cdot        sp
	{0, 0, 0, 0, 1, -172064161, -174666, 33386, 92052331, 9086, 15377},
	{0, 0, 2, -2, 2, -13170906, -1675, -13696, 5730336, -3015, -4587},
	{0, 0, 2, 0, 2, -2276413, -234, 2796, 978459, -485, 1374},
	{0, 0, 0, 0, 2, 2074554, 207, -698, -897492, 470, -291},
	{0, 1, 0, 0, 0, 1475877, -3633, 11817, 73871, -184, -1924},
	{1, 0, 0, 0, 0, 711159, 73, -872, -6750, 0, 358},
	{0, 1, 2, -2, 2, -516821, 1226, -524, 224386, -677, -174},
	{0, 0, 2, 0, 1, -387298, -367, 380, 200728, 18, 318},
	{1, 0, 2, 0, 2, -301461, -36, 816, 129025, -63, 367},
	{0, -1, 2, -2, 2, 215829, -494, 111, -95929, 299, 132},
	{-1, 0, 0, 2, 0, 156994, 10, -168, -1235, 0, 82},
	{0, 0, 2, -2, 1, 128227, 137, 181, -68982, -9, 39},
	{-1, 0, 2, 0, 2, 123457, 11, 19, -53311, 32, -4},
	{0, 0, 0, 2, 0, 63384, 11, -150, -1220, 0, 29},
	{1, 0, 0, 0, 1, 63110, 63, 27, -33228, 0, -9},
	{-1, 0, 2, 2, 2, -59641, -11, 149, 25543, -11, 66},
	{-1, 0, 0, 0, 1, -57976, -63, -189, 31429, 0, -75},
	{1, 0, 2, 0, 1, -51613, -42, 129, 26366, 0, 78},
	{-2, 0, 0, 2, 0, -47722, 0, -18, 477, 0, -25},
	{-2, 0, 2, 0, 1, 45893, 50, 31, -24236, -10, 20},
	{0, 0, 2, 2, 2, -38571, -1, 158, 16452, -11, 68},
	{0, -2, 2, -2, 2, 32481, 0, 0, -13870, 0, 0},
	{2, 0, 2, 0, 2, -31046, -1, 131, 13238, -11, 59},
	{2, 0, 0, 0, 0, 29243, 0, -74, -609, 0, 13},
	{1, 0, 2, -2, 2, 28593, 0, -1, -12338, 10, -3},
	{0, 0, 2, 0, 0, 25887, 0, -66, -550, 0, 11},
	{0, 0, -2, 2, 0, 21783, 0, 13, -167, 0, 13},
	{-1, 0, 2, 0, 1, 20441, 21, 10, -10758, 0, -3},
	{0, 2, 0, 0, 0, 16707, -85, -10, 168, -1, 10},
	{0, 2, 2, -2, 2, -15794, 72, -16, 6850, -42, -5},
}

// nutationAnglesStandard computes nutation using the 30 largest luni-solar terms.
// ~1 arcsec precision, ~45x faster than the full series.
func nutationAnglesStandard(T float64) (dpsiRad, depsRad float64) {
	l, lp, F, D, om := fundamentalArgs(T)

	var dpsi, deps float64
	for i := range nutationTerms {
		t := &nutationTerms[i]
		arg := float64(t.nl)*l + float64(t.nlp)*lp + float64(t.nf)*F +
			float64(t.nd)*D + float64(t.nom)*om
		sinArg, cosArg := math.Sincos(arg)
		dpsi += (t.s + t.sdot*T) * sinArg
		dpsi += t.cp * cosArg
		deps += (t.c + t.cdot*T) * cosArg
		deps += t.sp * sinArg
	}

	// Convert from 0.1 microarcseconds to radians
	dpsiRad = dpsi * tenthUas2Rad
	depsRad = deps * tenthUas2Rad
	return
}

// nutationAnglesFull computes nutation using the full IAU 2000A series:
// 678 luni-solar + 687 planetary terms. ~0.001 arcsec precision.
func nutationAnglesFull(T float64) (dpsiRad, depsRad float64) {
	l, lp, F, D, om := fundamentalArgs(T)
	fa := [5]float64{l, lp, F, D, om}

	// --- Luni-solar nutation (678 terms) ---
	var dpsi, deps float64
	for i := 0; i < 678; i++ {
		n := &luniSolarNals[i]
		arg := float64(n[0])*fa[0] + float64(n[1])*fa[1] + float64(n[2])*fa[2] +
			float64(n[3])*fa[3] + float64(n[4])*fa[4]
		sinArg, cosArg := math.Sincos(arg)
		lc := &luniSolarLonCoeffs[i]
		dpsi += (lc[0] + lc[1]*T)*sinArg + lc[2]*cosArg
		oc := &luniSolarOblCoeffs[i]
		deps += (oc[0] + oc[1]*T)*cosArg + oc[2]*sinArg
	}

	// --- Planetary nutation (687 terms) ---
	// Planetary fundamental arguments (mean longitudes, radians)
	pa := [14]float64{
		fa[0], fa[1], fa[2], fa[3], fa[4], // Delaunay args
		4.402608842 + 2608.7903141574*T,    // Mercury
		3.176146697 + 1021.3285546211*T,    // Venus
		1.753470314 + 628.3075849991*T,     // Earth
		6.203480913 + 334.0612426700*T,     // Mars
		0.599546497 + 52.9690962641*T,      // Jupiter
		0.874016757 + 21.3299104960*T,      // Saturn
		5.481293872 + 7.4781598567*T,       // Uranus
		5.311886287 + 3.8133035638*T,       // Neptune
		(0.024381750 + 0.00000538691*T) * T, // general precession
	}

	for i := 0; i < 687; i++ {
		n := &planetaryNapl[i]
		var arg float64
		for j := 0; j < 14; j++ {
			arg += float64(n[j]) * pa[j]
		}
		sinArg, cosArg := math.Sincos(arg)
		lc := &planetaryLonCoeffs[i]
		dpsi += lc[0]*sinArg + lc[1]*cosArg
		oc := &planetaryOblCoeffs[i]
		deps += oc[0]*sinArg + oc[1]*cosArg
	}

	// Convert from 0.1 microarcseconds to radians
	dpsiRad = dpsi * tenthUas2Rad
	depsRad = deps * tenthUas2Rad
	return
}

// nutationMatrixTranspose returns N^T, the transpose of the nutation matrix.
// N = R1(-trueOb) * R3(dpsi) * R1(meanOb) transforms mean equinox → true equinox.
// N^T transforms true equinox → mean equinox.
func nutationMatrixTranspose(dpsiRad, depsRad, epsMRad float64) [3][3]float64 {
	epsTRad := epsMRad + depsRad // true obliquity

	sinDpsi, cosDpsi := math.Sincos(dpsiRad)
	sinEpsM, cosEpsM := math.Sincos(epsMRad)
	sinEpsT, cosEpsT := math.Sincos(epsTRad)

	// N matrix (mean → true) using standard R3 convention (R3(α) has -sinα at [0][1]):
	//   N[0] = { cosDpsi, -sinDpsi*cosEpsM, -sinDpsi*sinEpsM }
	//   N[1] = { sinDpsi*cosEpsT, cosDpsi*cosEpsM*cosEpsT + sinEpsM*sinEpsT, cosDpsi*sinEpsM*cosEpsT - cosEpsM*sinEpsT }
	//   N[2] = { sinDpsi*sinEpsT, cosDpsi*cosEpsM*sinEpsT - sinEpsM*cosEpsT, cosDpsi*sinEpsM*sinEpsT + cosEpsM*cosEpsT }
	//
	// Return N^T (transpose):
	return [3][3]float64{
		{cosDpsi, sinDpsi * cosEpsT, sinDpsi * sinEpsT},
		{-sinDpsi * cosEpsM, cosDpsi*cosEpsM*cosEpsT + sinEpsM*sinEpsT, cosDpsi*cosEpsM*sinEpsT - sinEpsM*cosEpsT},
		{-sinDpsi * sinEpsM, cosDpsi*sinEpsM*cosEpsT - cosEpsM*sinEpsT, cosDpsi*sinEpsM*sinEpsT + cosEpsM*cosEpsT},
	}
}

// GAST returns Greenwich Apparent Sidereal Time in degrees, which includes
// the nutation correction (equation of equinoxes).
func GAST(jdUT1 float64) float64 {
	gmst := GMST(jdUT1)
	T := (jdUT1 - j2000JD) / 36525.0

	dpsiRad, _ := nutationAngles(T)
	epsM := meanObliquity(T)

	// Equation of equinoxes = dpsi * cos(meanOb), in degrees
	eqeqDeg := (dpsiRad * math.Cos(epsM)) * rad2deg

	return math.Mod(gmst+eqeqDeg, 360.0)
}

// precessionMatrixInverse computes the IAU 2006 precession matrix P that transforms
// vectors from J2000 to the mean equator and equinox of date.
// T is Julian centuries from J2000 TDB.
// Returns the transpose (inverse) P^T which transforms FROM date TO J2000.
func precessionMatrixInverse(T float64) [3][3]float64 {
	// IAU 2006 precession angles (arcseconds)
	zetaA := (2.650545 + 2306.083227*T + 0.2988499*T*T +
		0.01801828*T*T*T - 0.000005971*T*T*T*T) * arcsec2rad
	zA := (-2.650545 + 2306.077181*T + 1.0927348*T*T +
		0.01826837*T*T*T - 0.000028596*T*T*T*T) * arcsec2rad
	thetaA := (2004.191903*T - 0.4294934*T*T -
		0.04182264*T*T*T - 0.000007089*T*T*T*T) * arcsec2rad

	cosZetaA := math.Cos(zetaA)
	sinZetaA := math.Sin(zetaA)
	cosZA := math.Cos(zA)
	sinZA := math.Sin(zA)
	cosThetaA := math.Cos(thetaA)
	sinThetaA := math.Sin(thetaA)

	// Precession matrix P = Rz(-zA) · Ry(thetaA) · Rz(-zetaA)
	// P transforms J2000 → date
	// We want P^T (date → J2000)
	p11 := cosZA*cosThetaA*cosZetaA - sinZA*sinZetaA
	p12 := -cosZA*cosThetaA*sinZetaA - sinZA*cosZetaA
	p13 := -cosZA * sinThetaA
	p21 := sinZA*cosThetaA*cosZetaA + cosZA*sinZetaA
	p22 := -sinZA*cosThetaA*sinZetaA + cosZA*cosZetaA
	p23 := -sinZA * sinThetaA
	p31 := sinThetaA * cosZetaA
	p32 := -sinThetaA * sinZetaA
	p33 := cosThetaA

	// Return P^T (transpose = inverse for rotation matrix)
	return [3][3]float64{
		{p11, p21, p31},
		{p12, p22, p32},
		{p13, p23, p33},
	}
}

// TEMEToICRF converts a TEME (True Equator, Mean Equinox) position vector
// from SGP4 propagation to ICRF/GCRS coordinates.
//
// posKmTEME is the satellite position in km from SGP4 (TEME frame).
// jdUT1 is the UT1 Julian date (used for Earth rotation via nutation/precession).
//
// The TEME frame is the output frame of SGP4. It uses the true equator of
// date but a "mean" equinox that differs from the classical mean equinox
// by the equation of the equinoxes. The conversion chain is:
//
//	TEME → true equator of date (via equation of equinoxes rotation)
//	     → mean equator of date (via nutation inverse)
//	     → ICRF/J2000 (via precession inverse)
func TEMEToICRF(posKmTEME [3]float64, jdUT1 float64) [3]float64 {
	T := (jdUT1 - j2000JD) / 36525.0

	// Step 1: Compute equation of equinoxes = dpsi * cos(meanObliquity)
	dpsiRad, depsRad := nutationAngles(T)
	epsM := meanObliquity(T)
	eqEqRad := dpsiRad * math.Cos(epsM)

	// Step 2: Rotate TEME by Rz(eq_eq) → true equator/equinox of date
	sinE, cosE := math.Sincos(eqEqRad)
	xTrue := cosE*posKmTEME[0] - sinE*posKmTEME[1]
	yTrue := sinE*posKmTEME[0] + cosE*posKmTEME[1]
	zTrue := posKmTEME[2]

	// Step 3: Apply N^T (true equinox → mean equinox of date)
	NT := nutationMatrixTranspose(dpsiRad, depsRad, epsM)
	xMean := NT[0][0]*xTrue + NT[0][1]*yTrue + NT[0][2]*zTrue
	yMean := NT[1][0]*xTrue + NT[1][1]*yTrue + NT[1][2]*zTrue
	zMean := NT[2][0]*xTrue + NT[2][1]*yTrue + NT[2][2]*zTrue

	// Step 4: Apply P^T (mean equinox of date → J2000)
	PT := precessionMatrixInverse(T)
	xJ2000 := PT[0][0]*xMean + PT[0][1]*yMean + PT[0][2]*zMean
	yJ2000 := PT[1][0]*xMean + PT[1][1]*yMean + PT[1][2]*zMean
	zJ2000 := PT[2][0]*xMean + PT[2][1]*yMean + PT[2][2]*zMean

	// Step 5: Apply B^T (frame bias inverse: J2000 → ICRS)
	B := &ICRSToJ2000Matrix
	return [3]float64{
		B[0][0]*xJ2000 + B[1][0]*yJ2000 + B[2][0]*zJ2000,
		B[0][1]*xJ2000 + B[1][1]*yJ2000 + B[2][1]*zJ2000,
		B[0][2]*xJ2000 + B[1][2]*yJ2000 + B[2][2]*zJ2000,
	}
}

// GeodeticToICRF converts geodetic coordinates (lat/lon in degrees) to an ICRF
// direction vector at the given UT1 Julian date.
// Uses the full transformation: ICRF = P^T * N^T * Rz(GAST) * ITRF
func GeodeticToICRF(latDeg, lonDeg, jdUT1 float64) (x, y, z float64) {
	lat := latDeg * deg2rad
	lon := lonDeg * deg2rad

	sinLat := math.Sin(lat)
	cosLat := math.Cos(lat)
	sinLon := math.Sin(lon)
	cosLon := math.Cos(lon)

	// WGS84 normal radius of curvature
	N := wgs84A / math.Sqrt(1.0-wgs84E2*sinLat*sinLat)

	// Step 1: ITRF Cartesian position (km)
	xITRF := N * cosLat * cosLon
	yITRF := N * cosLat * sinLon
	zITRF := N * (1.0 - wgs84E2) * sinLat

	// Step 2: Compute nutation quantities
	T := (jdUT1 - j2000JD) / 36525.0
	dpsiRad, depsRad := nutationAngles(T)
	epsM := meanObliquity(T)

	// Step 3: GAST = GMST + equation of equinoxes
	gmstDeg := GMST(jdUT1)
	eqeqDeg := (dpsiRad * math.Cos(epsM)) * rad2deg
	gastRad := (gmstDeg + eqeqDeg) * deg2rad

	// Step 4: Rotate ITRF → true equinox of date by GAST
	sinG, cosG := math.Sincos(gastRad)
	xTrue := cosG*xITRF - sinG*yITRF
	yTrue := sinG*xITRF + cosG*yITRF
	zTrue := zITRF

	// Step 5: Apply N^T to go from true equinox of date → mean equinox of date
	NT := nutationMatrixTranspose(dpsiRad, depsRad, epsM)
	xMean := NT[0][0]*xTrue + NT[0][1]*yTrue + NT[0][2]*zTrue
	yMean := NT[1][0]*xTrue + NT[1][1]*yTrue + NT[1][2]*zTrue
	zMean := NT[2][0]*xTrue + NT[2][1]*yTrue + NT[2][2]*zTrue

	// Step 6: Apply P^T to go from mean equinox of date → J2000
	P := precessionMatrixInverse(T)
	xJ2000 := P[0][0]*xMean + P[0][1]*yMean + P[0][2]*zMean
	yJ2000 := P[1][0]*xMean + P[1][1]*yMean + P[1][2]*zMean
	zJ2000 := P[2][0]*xMean + P[2][1]*yMean + P[2][2]*zMean

	// Step 7: Apply B^T (frame bias inverse) to go from J2000 → ICRS
	B := &ICRSToJ2000Matrix
	xICRF := B[0][0]*xJ2000 + B[1][0]*yJ2000 + B[2][0]*zJ2000
	yICRF := B[0][1]*xJ2000 + B[1][1]*yJ2000 + B[2][1]*zJ2000
	zICRF := B[0][2]*xJ2000 + B[1][2]*yJ2000 + B[2][2]*zJ2000

	// Step 8: Normalize to unit vector
	r := math.Sqrt(xICRF*xICRF + yICRF*yICRF + zICRF*zICRF)
	return xICRF / r, yICRF / r, zICRF / r
}

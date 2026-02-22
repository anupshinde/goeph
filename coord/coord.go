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

// nutationAngles computes nutation in longitude (dpsi) and obliquity (deps)
// using the top 30 terms of the IAU 2000A luni-solar series.
// T is Julian centuries from J2000 TDB.
// Returns dpsi and deps in radians.
func nutationAngles(T float64) (dpsiRad, depsRad float64) {
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

// nutationMatrixTranspose returns N^T, the transpose of the nutation matrix.
// N = R1(-trueOb) * R3(dpsi) * R1(meanOb) transforms mean equinox → true equinox.
// N^T transforms true equinox → mean equinox.
func nutationMatrixTranspose(dpsiRad, depsRad, epsMRad float64) [3][3]float64 {
	epsTRad := epsMRad + depsRad // true obliquity

	sinDpsi, cosDpsi := math.Sincos(dpsiRad)
	sinEpsM, cosEpsM := math.Sincos(epsMRad)
	sinEpsT, cosEpsT := math.Sincos(epsTRad)

	// N matrix (mean → true) using IAU R3 convention (R3(α) has +sinα at [0][1]):
	//   N[0] = { cosDpsi, sinDpsi*cosEpsM, sinDpsi*sinEpsM }
	//   N[1] = { -sinDpsi*cosEpsT, cosDpsi*cosEpsM*cosEpsT + sinEpsM*sinEpsT, cosDpsi*sinEpsM*cosEpsT - cosEpsM*sinEpsT }
	//   N[2] = { -sinDpsi*sinEpsT, cosDpsi*cosEpsM*sinEpsT - sinEpsM*cosEpsT, cosDpsi*sinEpsM*sinEpsT + cosEpsM*cosEpsT }
	//
	// Return N^T (transpose):
	return [3][3]float64{
		{cosDpsi, -sinDpsi * cosEpsT, -sinDpsi * sinEpsT},
		{sinDpsi * cosEpsM, cosDpsi*cosEpsM*cosEpsT + sinEpsM*sinEpsT, cosDpsi*cosEpsM*sinEpsT - sinEpsM*cosEpsT},
		{sinDpsi * sinEpsM, cosDpsi*sinEpsM*cosEpsT - cosEpsM*sinEpsT, cosDpsi*sinEpsM*sinEpsT + cosEpsM*cosEpsT},
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

	// Step 6: Apply P^T to go from mean equinox of date → J2000/ICRF
	P := precessionMatrixInverse(T)
	xICRF := P[0][0]*xMean + P[0][1]*yMean + P[0][2]*zMean
	yICRF := P[1][0]*xMean + P[1][1]*yMean + P[1][2]*zMean
	zICRF := P[2][0]*xMean + P[2][1]*yMean + P[2][2]*zMean

	// Step 7: Normalize to unit vector
	r := math.Sqrt(xICRF*xICRF + yICRF*yICRF + zICRF*zICRF)
	return xICRF / r, yICRF / r, zICRF / r
}

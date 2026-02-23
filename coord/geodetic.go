package coord

import "math"

// ITRFToGeodetic converts ITRF Cartesian coordinates (km) to geodetic
// latitude, longitude (degrees), and height above the WGS84 ellipsoid (km).
//
// Uses Bowring's iterative method (converges in 2-3 iterations for terrestrial
// positions, and handles all cases including poles and equator).
func ITRFToGeodetic(x, y, z float64) (latDeg, lonDeg, heightKm float64) {
	lonDeg = math.Atan2(y, x) * rad2deg

	// Distance from Z axis
	p := math.Sqrt(x*x + y*y)

	if p == 0 {
		// On the polar axis
		if z >= 0 {
			latDeg = 90.0
		} else {
			latDeg = -90.0
		}
		heightKm = math.Abs(z) - wgs84A*(1.0-wgs84F)
		return
	}

	// Initial estimate using Bowring's formula
	b := wgs84A * (1.0 - wgs84F) // semi-minor axis
	theta := math.Atan2(z*wgs84A, p*b)
	sinTheta, cosTheta := math.Sincos(theta)

	lat := math.Atan2(
		z+wgs84E2/(1.0-wgs84F)*b*sinTheta*sinTheta*sinTheta,
		p-wgs84E2*wgs84A*cosTheta*cosTheta*cosTheta,
	)

	// Iterate (3 iterations for sub-mm accuracy)
	for range 3 {
		sinLat := math.Sin(lat)
		N := wgs84A / math.Sqrt(1.0-wgs84E2*sinLat*sinLat)
		lat = math.Atan2(z+wgs84E2*N*sinLat, p)
	}

	sinLat := math.Sin(lat)
	cosLat := math.Cos(lat)
	N := wgs84A / math.Sqrt(1.0-wgs84E2*sinLat*sinLat)

	if math.Abs(cosLat) > 1e-10 {
		heightKm = p/cosLat - N
	} else {
		heightKm = math.Abs(z)/math.Abs(sinLat) - N*(1.0-wgs84E2)
	}

	latDeg = lat * rad2deg
	return
}

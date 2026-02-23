package satellite

import (
	"math"
	"time"

	gosatellite "github.com/joshuaferrara/go-satellite"

	"github.com/anupshinde/goeph/coord"
)

// Sat holds a named satellite for propagation.
type Sat struct {
	Name string
	Sat  gosatellite.Satellite
}

// NewSat creates a Sat from TLE lines using WGS84 gravity model.
func NewSat(name, line1, line2 string) Sat {
	return Sat{
		Name: name,
		Sat:  gosatellite.TLEToSat(line1, line2, gosatellite.GravityWGS84),
	}
}

// SubPoint returns the sub-satellite point (geographic lat/lon in degrees).
func SubPoint(s gosatellite.Satellite, t time.Time) (latDeg, lonDeg float64) {
	year := t.Year()
	month := int(t.Month())
	day := t.Day()
	hour := t.Hour()
	min := t.Minute()
	sec := t.Second()

	pos, _ := gosatellite.Propagate(s, year, month, day, hour, min, sec)
	jd := gosatellite.JDay(year, month, day, hour, min, sec)
	gmst := gosatellite.ThetaG_JD(jd)

	_, _, latLong := gosatellite.ECIToLLA(pos, gmst)
	ll := gosatellite.LatLongDeg(latLong)

	lonDeg = math.Mod(ll.Longitude+360.0, 360.0)
	return ll.Latitude, lonDeg
}

// TEMEToICRF converts a TEME (True Equator, Mean Equinox) position vector
// from SGP4 propagation to ICRF/GCRS coordinates.
//
// posKmTEME is the satellite position in km from SGP4 (TEME frame).
// jdUT1 is the UT1 Julian date (used for Earth rotation via GAST).
//
// The TEME frame is the output frame of SGP4. It uses the true equator of
// date but a "mean" equinox that differs from the classical mean equinox
// by the equation of the equinoxes. The conversion chain is:
//
//	TEME → equator of date (via equation of equinoxes rotation)
//	     → mean equator of date (via nutation^-1)
//	     → ICRF/J2000 (via precession^-1)
//
// This matches Skyfield's TEME→GCRS conversion for SGP4 satellite positions.
func TEMEToICRF(posKmTEME [3]float64, jdUT1 float64) [3]float64 {
	return coord.TEMEToICRF(posKmTEME, jdUT1)
}

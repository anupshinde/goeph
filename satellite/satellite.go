package satellite

import (
	"math"
	"time"

	gosatellite "github.com/joshuaferrara/go-satellite"

	"github.com/anupshinde/goeph/coord"
	"github.com/anupshinde/goeph/search"
	"github.com/anupshinde/goeph/timescale"
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

// Event kinds returned by FindEvents.
const (
	Rise        = 0 // Satellite rises above the altitude threshold
	Culmination = 1 // Satellite reaches maximum altitude during a pass
	Set         = 2 // Satellite sets below the altitude threshold
)

// SatEvent represents a satellite pass event (rise, culmination, or set).
type SatEvent struct {
	T      float64 // TT Julian date of the event
	Kind   int     // Rise=0, Culmination=1, Set=2
	AltDeg float64 // Altitude in degrees at the event time
}

// FindEvents finds satellite rise, culmination, and set events as seen from a
// ground observer in the given TT Julian date range.
//
// latDeg, lonDeg: observer geodetic latitude and longitude in degrees.
// minAltDeg: minimum altitude threshold in degrees (typically 0).
//
// Returns events sorted by time. Each visible pass produces up to three events:
// Rise (satellite crosses above threshold), Culmination (maximum altitude),
// and Set (satellite crosses below threshold).
func FindEvents(sat Sat, latDeg, lonDeg, startJD, endJD, minAltDeg float64) ([]SatEvent, error) {
	// Step size ~1 minute. LEO orbital period ~90 min, shortest visible pass ~2 min.
	const stepDays = 1.0 / 1440.0 // 1 minute

	altFunc := satAltitudeFunc(sat, latDeg, lonDeg)

	// Find rise/set transitions using discrete search.
	discreteFunc := func(ttJD float64) int {
		if altFunc(ttJD) >= minAltDeg {
			return 1
		}
		return 0
	}
	transitions, err := search.FindDiscrete(startJD, endJD, stepDays, discreteFunc, 0)
	if err != nil {
		return nil, err
	}

	// Group transitions into passes and find culminations.
	var events []SatEvent
	for i := 0; i < len(transitions); i++ {
		e := transitions[i]
		if e.NewValue == 1 {
			// Rise event.
			riseT := e.T
			events = append(events, SatEvent{T: riseT, Kind: Rise, AltDeg: altFunc(riseT)})

			// Look for the matching set event.
			setT := endJD
			if i+1 < len(transitions) && transitions[i+1].NewValue == 0 {
				setT = transitions[i+1].T
				i++ // consume the set event

				// Find culmination between rise and set.
				maxima, err := search.FindMaxima(riseT, setT, stepDays, altFunc, 0)
				if err == nil && len(maxima) > 0 {
					// Use the highest maximum.
					best := maxima[0]
					for _, m := range maxima[1:] {
						if m.Value > best.Value {
							best = m
						}
					}
					events = append(events, SatEvent{T: best.T, Kind: Culmination, AltDeg: best.Value})
				}

				events = append(events, SatEvent{T: setT, Kind: Set, AltDeg: altFunc(setT)})
			}
		}
	}

	return events, nil
}

// satAltitudeFunc returns a function that computes the satellite's altitude
// in degrees as seen from the given ground observer at a TT Julian date.
func satAltitudeFunc(sat Sat, latDeg, lonDeg float64) func(float64) float64 {
	return func(ttJD float64) float64 {
		jdUT1 := timescale.TTToUT1(ttJD)

		// Convert JD to calendar for SGP4 propagation.
		y, mo, d, h, mi, s := jdToCalendar(jdUT1)
		pos, _ := gosatellite.Propagate(sat.Sat, y, mo, d, h, mi, s)

		// SGP4 position is in km, TEME frame. Convert to ICRF.
		posKmTEME := [3]float64{pos.X, pos.Y, pos.Z}
		satICRF := coord.TEMEToICRF(posKmTEME, jdUT1)

		// Observer position in ICRF (km).
		ox, oy, oz := coord.GeodeticToICRF(latDeg, lonDeg, jdUT1)

		// Topocentric vector in ICRF.
		topoICRF := [3]float64{
			satICRF[0] - ox,
			satICRF[1] - oy,
			satICRF[2] - oz,
		}

		alt, _, _ := coord.Altaz(topoICRF, latDeg, lonDeg, jdUT1)
		return alt
	}
}

// jdToCalendar converts a Julian date to calendar components.
func jdToCalendar(jd float64) (year, month, day, hour, min, sec int) {
	// Standard JD to calendar algorithm (Meeus, Astronomical Algorithms).
	jd += 0.5
	z := math.Floor(jd)
	f := jd - z

	var a float64
	if z < 2299161 {
		a = z
	} else {
		alpha := math.Floor((z - 1867216.25) / 36524.25)
		a = z + 1 + alpha - math.Floor(alpha/4)
	}

	b := a + 1524
	c := math.Floor((b - 122.1) / 365.25)
	d := math.Floor(365.25 * c)
	e := math.Floor((b - d) / 30.6001)

	dayFrac := b - d - math.Floor(30.6001*e) + f
	day = int(dayFrac)
	fracDay := dayFrac - float64(day)

	if e < 14 {
		month = int(e) - 1
	} else {
		month = int(e) - 13
	}
	if month > 2 {
		year = int(c) - 4716
	} else {
		year = int(c) - 4715
	}

	totalSec := fracDay * 86400.0
	hour = int(totalSec / 3600.0)
	totalSec -= float64(hour) * 3600.0
	min = int(totalSec / 60.0)
	sec = int(totalSec - float64(min)*60.0)

	return
}

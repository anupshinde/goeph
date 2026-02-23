// Package almanac provides astronomical event-finding functions built on the
// search package. It finds times of seasons, moon phases, sunrise/sunset,
// twilight, body risings/settings, meridian transits, and oppositions/conjunctions.
package almanac

import (
	"math"

	"github.com/anupshinde/goeph/coord"
	"github.com/anupshinde/goeph/search"
	"github.com/anupshinde/goeph/spk"
	"github.com/anupshinde/goeph/timescale"
)

// Season values returned in DiscreteEvent.NewValue by Seasons.
const (
	SpringEquinox  = 0 // Sun ecliptic longitude crosses 0°
	SummerSolstice = 1 // Sun ecliptic longitude crosses 90°
	AutumnEquinox  = 2 // Sun ecliptic longitude crosses 180°
	WinterSolstice = 3 // Sun ecliptic longitude crosses 270°
)

// Moon phase values returned in DiscreteEvent.NewValue by MoonPhases.
const (
	NewMoon      = 0 // Moon-Sun elongation crosses 0°
	FirstQuarter = 1 // Moon-Sun elongation crosses 90°
	FullMoon     = 2 // Moon-Sun elongation crosses 180°
	LastQuarter  = 3 // Moon-Sun elongation crosses 270°
)

// Twilight level values returned in DiscreteEvent.NewValue by Twilight.
const (
	Night                = 0 // Sun altitude < -18°
	AstronomicalTwilight = 1 // -18° ≤ alt < -12°
	NauticalTwilight     = 2 // -12° ≤ alt < -6°
	CivilTwilight        = 3 // -6° ≤ alt < -0.8333°
	Daylight             = 4 // alt ≥ -0.8333°
)

// sunAltitudeThreshold is the standard altitude for sunrise/sunset:
// -50 arcminutes = -0.8333° (16' solar radius + 34' refraction).
const sunAltitudeThreshold = -0.8333

// refractionThreshold is the standard altitude adjustment for atmospheric
// refraction alone (-34 arcminutes), used for non-solar body risings/settings.
const refractionThreshold = -34.0 / 60.0

// Seasons finds equinoxes and solstices in the given TDB Julian date range.
//
// Returns events with NewValue: SpringEquinox=0, SummerSolstice=1,
// AutumnEquinox=2, WinterSolstice=3 (Northern Hemisphere conventions).
func Seasons(eph *spk.SPK, startJD, endJD float64) ([]search.DiscreteEvent, error) {
	f := func(tdbJD float64) int {
		pos := eph.Apparent(spk.Sun, tdbJD)
		_, lonDeg := coord.ICRFToEcliptic(pos[0], pos[1], pos[2])
		if lonDeg < 0 {
			lonDeg += 360.0
		}
		return int(math.Floor(lonDeg/90.0)) % 4
	}
	return search.FindDiscrete(startJD, endJD, 90.0, f, 0)
}

// MoonPhases finds new moons, first quarters, full moons, and last quarters
// in the given TDB Julian date range.
//
// Returns events with NewValue: NewMoon=0, FirstQuarter=1, FullMoon=2,
// LastQuarter=3.
func MoonPhases(eph *spk.SPK, startJD, endJD float64) ([]search.DiscreteEvent, error) {
	f := func(tdbJD float64) int {
		moonPos := eph.Apparent(spk.Moon, tdbJD)
		sunPos := eph.Apparent(spk.Sun, tdbJD)
		_, moonLon := coord.ICRFToEcliptic(moonPos[0], moonPos[1], moonPos[2])
		_, sunLon := coord.ICRFToEcliptic(sunPos[0], sunPos[1], sunPos[2])
		diff := moonLon - sunLon
		if diff < 0 {
			diff += 360.0
		}
		return int(math.Floor(diff/90.0)) % 4
	}
	return search.FindDiscrete(startJD, endJD, 5.0, f, 0)
}

// sunAltitude returns the Sun's altitude in degrees as seen from a ground observer.
func sunAltitude(eph *spk.SPK, latDeg, lonDeg, tdbJD float64) float64 {
	pos := eph.Apparent(spk.Sun, tdbJD)
	jdUT1 := timescale.TTToUT1(tdbJD)
	alt, _, _ := coord.Altaz(pos, latDeg, lonDeg, jdUT1)
	return alt
}

// SunriseSunset finds sunrise and sunset times for a ground observer in the
// given TDB Julian date range.
//
// latDeg, lonDeg: observer geodetic latitude and longitude in degrees.
// Returns events with NewValue=1 (sunrise) and NewValue=0 (sunset).
func SunriseSunset(eph *spk.SPK, latDeg, lonDeg, startJD, endJD float64) ([]search.DiscreteEvent, error) {
	f := func(tdbJD float64) int {
		if sunAltitude(eph, latDeg, lonDeg, tdbJD) >= sunAltitudeThreshold {
			return 1
		}
		return 0
	}
	return search.FindDiscrete(startJD, endJD, 0.04, f, 0)
}

// Twilight finds transitions between darkness, twilight levels, and daylight
// for a ground observer in the given TDB Julian date range.
//
// Returns events with NewValue: Night=0, AstronomicalTwilight=1,
// NauticalTwilight=2, CivilTwilight=3, Daylight=4.
func Twilight(eph *spk.SPK, latDeg, lonDeg, startJD, endJD float64) ([]search.DiscreteEvent, error) {
	f := func(tdbJD float64) int {
		alt := sunAltitude(eph, latDeg, lonDeg, tdbJD)
		switch {
		case alt >= sunAltitudeThreshold:
			return Daylight
		case alt >= -6.0:
			return CivilTwilight
		case alt >= -12.0:
			return NauticalTwilight
		case alt >= -18.0:
			return AstronomicalTwilight
		default:
			return Night
		}
	}
	return search.FindDiscrete(startJD, endJD, 0.01, f, 0)
}

// bodyAltitude returns a body's altitude in degrees as seen from a ground observer.
func bodyAltitude(eph *spk.SPK, body int, latDeg, lonDeg, tdbJD float64) float64 {
	pos := eph.Apparent(body, tdbJD)
	jdUT1 := timescale.TTToUT1(tdbJD)
	alt, _, _ := coord.Altaz(pos, latDeg, lonDeg, jdUT1)
	return alt
}

// Risings finds times when a body rises above the horizon for a ground observer
// in the given TDB Julian date range.
//
// The horizon is at -34 arcminutes (atmospheric refraction). Returns events
// with NewValue=1 (body rose).
func Risings(eph *spk.SPK, body int, latDeg, lonDeg, startJD, endJD float64) ([]search.DiscreteEvent, error) {
	f := func(tdbJD float64) int {
		if bodyAltitude(eph, body, latDeg, lonDeg, tdbJD) >= refractionThreshold {
			return 1
		}
		return 0
	}
	events, err := search.FindDiscrete(startJD, endJD, 0.25, f, 0)
	if err != nil {
		return nil, err
	}
	// Filter to only rising events.
	var risings []search.DiscreteEvent
	for _, e := range events {
		if e.NewValue == 1 {
			risings = append(risings, e)
		}
	}
	return risings, nil
}

// Settings finds times when a body sets below the horizon for a ground observer
// in the given TDB Julian date range.
//
// The horizon is at -34 arcminutes (atmospheric refraction). Returns events
// with NewValue=0 (body set).
func Settings(eph *spk.SPK, body int, latDeg, lonDeg, startJD, endJD float64) ([]search.DiscreteEvent, error) {
	f := func(tdbJD float64) int {
		if bodyAltitude(eph, body, latDeg, lonDeg, tdbJD) >= refractionThreshold {
			return 1
		}
		return 0
	}
	events, err := search.FindDiscrete(startJD, endJD, 0.25, f, 0)
	if err != nil {
		return nil, err
	}
	// Filter to only setting events.
	var settings []search.DiscreteEvent
	for _, e := range events {
		if e.NewValue == 0 {
			settings = append(settings, e)
		}
	}
	return settings, nil
}

// Transits finds times when a body crosses the observer's meridian (upper
// culmination) in the given TDB Julian date range.
//
// Returns events with NewValue=1 (body crossed from east to west of meridian).
func Transits(eph *spk.SPK, body int, latDeg, lonDeg, startJD, endJD float64) ([]search.DiscreteEvent, error) {
	f := func(tdbJD float64) int {
		pos := eph.Apparent(body, tdbJD)
		jdUT1 := timescale.TTToUT1(tdbJD)
		haDeg, _ := coord.HourAngleDec(pos, lonDeg, jdUT1)
		// HA > 180° means west of meridian (past transit).
		if haDeg > 180.0 {
			return 0 // east, approaching meridian
		}
		return 1 // west, past meridian
	}
	events, err := search.FindDiscrete(startJD, endJD, 0.4, f, 0)
	if err != nil {
		return nil, err
	}
	// Filter to only east-to-west transitions (actual transits).
	var transits []search.DiscreteEvent
	for _, e := range events {
		if e.NewValue == 1 {
			transits = append(transits, e)
		}
	}
	return transits, nil
}

// OppositionsConjunctions finds times when a planet is at opposition or
// conjunction with the Sun in the given TDB Julian date range.
//
// Returns events with NewValue=0 (conjunction: planet near Sun) and
// NewValue=1 (opposition: planet opposite Sun).
func OppositionsConjunctions(eph *spk.SPK, body int, startJD, endJD float64) ([]search.DiscreteEvent, error) {
	f := func(tdbJD float64) int {
		sunPos := eph.Apparent(spk.Sun, tdbJD)
		bodyPos := eph.Apparent(body, tdbJD)
		_, sunLon := coord.ICRFToEcliptic(sunPos[0], sunPos[1], sunPos[2])
		_, bodyLon := coord.ICRFToEcliptic(bodyPos[0], bodyPos[1], bodyPos[2])
		diff := sunLon - bodyLon
		if diff < 0 {
			diff += 360.0
		}
		return int(math.Floor(diff/180.0)) % 2
	}
	return search.FindDiscrete(startJD, endJD, 40.0, f, 0)
}

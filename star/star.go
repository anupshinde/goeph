// Package star provides fixed star positions with proper motion, parallax,
// and radial velocity propagation. Matches Skyfield's Star class.
package star

import (
	"math"

	"github.com/anupshinde/goeph/coord"
)

const (
	// auKm is the IAU astronomical unit in km.
	auKm = 149597870.7

	// cKmPerS is the speed of light in km/s.
	cKmPerS = 299792.458

	// cAUPerDay is the speed of light in AU/day.
	cAUPerDay = cKmPerS * 86400.0 / auKm

	// asec2rad converts arcseconds to radians.
	asec2rad = math.Pi / (180.0 * 3600.0)

	// j2000 is the Julian date of J2000.0.
	j2000 = 2451545.0
)

// Star represents a fixed star with optional proper motion, parallax, and
// radial velocity for astrometric propagation from catalog epoch to any date.
type Star struct {
	// RAHours is the right ascension at epoch in hours (0-24).
	RAHours float64

	// DecDeg is the declination at epoch in degrees (-90 to +90).
	DecDeg float64

	// ParallaxMas is the parallax in milliarcseconds.
	// Zero or negative means effectively infinite distance.
	ParallaxMas float64

	// RAMasPerYear is the proper motion in RA (μα*) in mas/year.
	// This is the rate of change of RA multiplied by cos(Dec).
	RAMasPerYear float64

	// DecMasPerYear is the proper motion in Dec (μδ) in mas/year.
	DecMasPerYear float64

	// RadialKmPerS is the radial velocity in km/s (positive = receding).
	RadialKmPerS float64

	// Epoch is the catalog epoch as a TDB Julian date.
	// If zero, J2000.0 (2451545.0) is used.
	Epoch float64

	// precomputed ICRF position (AU) and velocity (AU/day)
	posAU [3]float64
	velAU [3]float64
	ready bool
}

// init precomputes the ICRF position and velocity vectors from the catalog
// parameters. Called lazily on first use.
func (s *Star) init() {
	if s.ready {
		return
	}
	s.ready = true

	ra := s.RAHours * 15.0 * math.Pi / 180.0 // radians
	dec := s.DecDeg * math.Pi / 180.0         // radians

	// Distance from parallax.
	parallax := s.ParallaxMas
	if parallax <= 0 {
		parallax = 1e-6 // effectively infinite distance
	}
	distance := 1.0 / math.Sin(parallax*1e-3*asec2rad) // AU

	// Position at epoch.
	cosDec := math.Cos(dec)
	sinDec := math.Sin(dec)
	cosRA := math.Cos(ra)
	sinRA := math.Sin(ra)

	s.posAU = [3]float64{
		distance * cosDec * cosRA,
		distance * cosDec * sinRA,
		distance * sinDec,
	}

	// Velocity at epoch (Doppler factor k accounts for time dilation).
	k := 1.0 / (1.0 - s.RadialKmPerS/cKmPerS)

	pmr := (s.RAMasPerYear / (parallax * 365.25)) * k  // AU/day
	pmd := (s.DecMasPerYear / (parallax * 365.25)) * k  // AU/day
	rvl := (s.RadialKmPerS * 86400.0 / auKm) * k       // AU/day

	s.velAU = [3]float64{
		-pmr*sinRA - pmd*sinDec*cosRA + rvl*cosDec*cosRA,
		pmr*cosRA - pmd*sinDec*sinRA + rvl*cosDec*sinRA,
		pmd*cosDec + rvl*sinDec,
	}
}

// PositionAU returns the ICRF position of the star in AU at the given TDB
// Julian date, propagated from the catalog epoch using proper motion and
// radial velocity. No light-time correction is applied.
func (s *Star) PositionAU(tdbJD float64) [3]float64 {
	s.init()
	epoch := s.Epoch
	if epoch == 0 {
		epoch = j2000
	}
	dt := tdbJD - epoch // days

	return [3]float64{
		s.posAU[0] + s.velAU[0]*dt,
		s.posAU[1] + s.velAU[1]*dt,
		s.posAU[2] + s.velAU[2]*dt,
	}
}

// PositionKm returns the ICRF position of the star in km at the given TDB
// Julian date, propagated from the catalog epoch.
func (s *Star) PositionKm(tdbJD float64) [3]float64 {
	pos := s.PositionAU(tdbJD)
	return [3]float64{
		pos[0] * auKm,
		pos[1] * auKm,
		pos[2] * auKm,
	}
}

// RADec returns the RA (hours) and Dec (degrees) of the star at the given
// TDB Julian date, propagated from the catalog epoch.
func (s *Star) RADec(tdbJD float64) (raHours, decDeg float64) {
	pos := s.PositionAU(tdbJD)
	r := math.Sqrt(pos[0]*pos[0] + pos[1]*pos[1] + pos[2]*pos[2])
	decDeg = math.Asin(pos[2]/r) * 180.0 / math.Pi
	ra := math.Atan2(pos[1], pos[0]) * 180.0 / math.Pi
	if ra < 0 {
		ra += 360.0
	}
	return ra / 15.0, decDeg
}

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

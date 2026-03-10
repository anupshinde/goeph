// Package eclipse provides lunar and solar eclipse detection and characterization.
//
// It finds times when the Moon enters Earth's shadow (lunar eclipses) or the
// Moon's shadow falls on Earth (solar eclipses). Classifies eclipses by type
// and computes magnitudes. Uses the Danjon enlargement correction (2%
// atmospheric enlargement of Earth's shadow) for lunar eclipses.
package eclipse

import (
	"math"

	"github.com/anupshinde/goeph/search"
	"github.com/anupshinde/goeph/spk"
)

const (
	// Lunar eclipse type constants returned in LunarEclipse.Kind.
	Penumbral = 1 // Moon enters penumbra only
	Partial   = 2 // Moon partially enters umbra
	Total     = 3 // Moon fully within umbra

	// Solar eclipse type constants returned in SolarEclipse.Kind.
	SolarPartial = 1 // Moon partially covers Sun
	SolarAnnular = 2 // Moon inside Sun's disk but smaller
	SolarTotal   = 3 // Moon fully covers Sun

	// Physical constants.
	sunRadiusKm   = 695700.0
	earthRadiusKm = 6371.0
	moonRadiusKm  = 1737.4

	// Danjon enlargement factor: atmospheric refraction enlarges
	// Earth's shadow by ~2%.
	danjonFactor = 1.02
)

// LunarEclipse describes a lunar eclipse event.
type LunarEclipse struct {
	// T is the TDB Julian date of maximum eclipse (closest approach of
	// Moon center to shadow axis).
	T float64

	// Kind is the eclipse type: Penumbral (1), Partial (2), or Total (3).
	Kind int

	// UmbralMag is the umbral magnitude: fraction of Moon's diameter
	// immersed in the umbral shadow. Negative means Moon does not reach umbra.
	UmbralMag float64

	// PenumbralMag is the penumbral magnitude: fraction of Moon's diameter
	// immersed in the penumbral shadow.
	PenumbralMag float64

	// ClosestApproachKm is the minimum distance from Moon center to the
	// shadow axis, in km.
	ClosestApproachKm float64

	// UmbralRadiusKm is the umbral shadow radius at the Moon's distance, in km.
	// Includes Danjon enlargement.
	UmbralRadiusKm float64

	// PenumbralRadiusKm is the penumbral shadow radius at the Moon's distance, in km.
	// Includes Danjon enlargement.
	PenumbralRadiusKm float64
}

// FindLunarEclipses finds all lunar eclipses in the given TDB Julian date range.
//
// The algorithm:
//  1. Find approximate full moon times (Moon-Sun elongation ≈ 180°)
//  2. Refine each to the exact time of minimum Moon-shadow separation
//  3. Compute shadow geometry and classify eclipse type
//
// Returns eclipses sorted by time. Only events where the Moon at least
// partially enters the penumbra are returned.
func FindLunarEclipses(eph *spk.SPK, startJD, endJD float64) ([]LunarEclipse, error) {
	// Step 1: Find approximate full moon times by detecting when the
	// Moon-Sun elongation phase crosses through the "full moon" quadrant.
	// We use a discrete function that returns floor(elongation/90) % 4,
	// and look for transitions to value 2 (full moon = elongation 180°-270°).
	phaseFunc := func(tdbJD float64) int {
		sunPos := eph.Apparent(spk.Sun, tdbJD)
		moonPos := eph.Apparent(spk.Moon, tdbJD)
		elong := eclipticElongation(moonPos, sunPos)
		if elong < 0 {
			elong += 360
		}
		return int(math.Floor(elong/90.0)) % 4
	}

	transitions, err := search.FindDiscrete(startJD, endJD, 5.0, phaseFunc, 0)
	if err != nil {
		return nil, err
	}

	// Collect approximate full moon times (transition to phase 2).
	var fullMoons []float64
	for _, e := range transitions {
		if e.NewValue == 2 {
			fullMoons = append(fullMoons, e.T)
		}
	}

	// Step 2: For each full moon, find minimum Moon-shadow-axis separation.
	sepFunc := func(tdbJD float64) float64 {
		return moonShadowSeparation(eph, tdbJD)
	}

	var eclipses []LunarEclipse
	for _, fm := range fullMoons {
		// Search for minimum separation in a window around full moon.
		window := 1.5 // days
		minima, err := search.FindMinima(fm-window, fm+window, 0.02, sepFunc, 0)
		if err != nil || len(minima) == 0 {
			continue
		}

		// Use the minimum closest to the full moon time.
		best := minima[0]
		for _, m := range minima[1:] {
			if math.Abs(m.T-fm) < math.Abs(best.T-fm) {
				best = m
			}
		}

		// Step 3: Compute full shadow geometry at the minimum.
		ecl := ClassifyLunarEclipse(eph, best.T)
		if ecl.Kind > 0 {
			eclipses = append(eclipses, ecl)
		}
	}

	return eclipses, nil
}

// moonShadowSeparation returns the perpendicular distance (km) from the
// Moon's center to Earth's shadow axis at the given time.
func moonShadowSeparation(eph *spk.SPK, tdbJD float64) float64 {
	sunPos := eph.GeocentricPosition(spk.Sun, tdbJD)
	moonPos := eph.GeocentricPosition(spk.Moon, tdbJD)

	// Shadow axis direction: anti-solar, from Earth away from Sun.
	sunDist := math.Sqrt(sunPos[0]*sunPos[0] + sunPos[1]*sunPos[1] + sunPos[2]*sunPos[2])
	axis := [3]float64{
		-sunPos[0] / sunDist,
		-sunPos[1] / sunDist,
		-sunPos[2] / sunDist,
	}

	// Project Moon position onto shadow axis.
	dAlong := moonPos[0]*axis[0] + moonPos[1]*axis[1] + moonPos[2]*axis[2]

	// Perpendicular vector from shadow axis to Moon.
	perpX := moonPos[0] - dAlong*axis[0]
	perpY := moonPos[1] - dAlong*axis[1]
	perpZ := moonPos[2] - dAlong*axis[2]

	return math.Sqrt(perpX*perpX + perpY*perpY + perpZ*perpZ)
}

// ClassifyLunarEclipse computes the full eclipse geometry at a given time and
// returns a LunarEclipse. If the Moon is not in the penumbra, Kind is 0.
func ClassifyLunarEclipse(eph *spk.SPK, tdbJD float64) LunarEclipse {
	sunPos := eph.GeocentricPosition(spk.Sun, tdbJD)
	moonPos := eph.GeocentricPosition(spk.Moon, tdbJD)

	sunDist := math.Sqrt(sunPos[0]*sunPos[0] + sunPos[1]*sunPos[1] + sunPos[2]*sunPos[2])
	axis := [3]float64{
		-sunPos[0] / sunDist,
		-sunPos[1] / sunDist,
		-sunPos[2] / sunDist,
	}

	// Moon distance along shadow axis (should be positive for eclipse geometry).
	dAlong := moonPos[0]*axis[0] + moonPos[1]*axis[1] + moonPos[2]*axis[2]

	// Perpendicular distance from Moon center to shadow axis.
	perpX := moonPos[0] - dAlong*axis[0]
	perpY := moonPos[1] - dAlong*axis[1]
	perpZ := moonPos[2] - dAlong*axis[2]
	sep := math.Sqrt(perpX*perpX + perpY*perpY + perpZ*perpZ)

	// Shadow cone radii at Moon's distance along the shadow axis,
	// with Danjon 2% enlargement.
	rUmbra := (earthRadiusKm - dAlong*(sunRadiusKm-earthRadiusKm)/sunDist) * danjonFactor
	rPenumbra := (earthRadiusKm + dAlong*(sunRadiusKm+earthRadiusKm)/sunDist) * danjonFactor

	// Eclipse magnitudes.
	umbralMag := (rUmbra + moonRadiusKm - sep) / (2.0 * moonRadiusKm)
	penumbralMag := (rPenumbra + moonRadiusKm - sep) / (2.0 * moonRadiusKm)

	ecl := LunarEclipse{
		T:                 tdbJD,
		UmbralMag:         umbralMag,
		PenumbralMag:      penumbralMag,
		ClosestApproachKm: sep,
		UmbralRadiusKm:    rUmbra,
		PenumbralRadiusKm: rPenumbra,
	}

	// Classify.
	switch {
	case umbralMag >= 1.0:
		ecl.Kind = Total
	case umbralMag > 0:
		ecl.Kind = Partial
	case penumbralMag > 0:
		ecl.Kind = Penumbral
	default:
		ecl.Kind = 0 // not an eclipse
	}

	return ecl
}

// SolarEclipse describes a solar eclipse event using geocentric shadow cone
// geometry. This determines whether a solar eclipse is occurring somewhere on
// Earth and classifies its type. Suitable for chart-level flags; for local
// visibility or path computation, topocentric parallax and Earth's ellipsoidal
// shape must be considered.
//
// Note: Skyfield's experimental design/solar_eclipse.py takes the topocentric
// approach — projecting the Moon's shadow onto Earth's WGS84 ellipsoid via
// line-ellipsoid intersection to compute the ground track. That is needed for
// mapping eclipse paths but is far more complex than what is required here.
// This implementation uses the simpler shadow cone approach (inverse of the
// lunar eclipse geometry): compute the Moon's umbral/penumbral cone radii at
// Earth's distance and check whether Earth's disk intersects those cones.
type SolarEclipse struct {
	// T is the TDB Julian date of the classification instant.
	T float64

	// Kind is the eclipse type: 0 (none), SolarPartial (1),
	// SolarAnnular (2), or SolarTotal (3).
	Kind int

	// Gamma is the closest approach of the Moon's shadow axis to Earth's
	// center, in Earth radii. |Gamma| < 1 means the shadow axis crosses
	// Earth's disk (central eclipse). Values up to ~1.5 can still produce
	// a partial eclipse.
	Gamma float64

	// UmbralRadiusKm is the Moon's umbral cone radius at Earth's distance,
	// in km. Positive means the umbral cone reaches Earth (total eclipse
	// geometry). Negative means the cone vertex falls short and the
	// antumbra reaches Earth instead (annular eclipse geometry).
	UmbralRadiusKm float64

	// PenumbralRadiusKm is the Moon's penumbral cone radius at Earth's
	// distance, in km.
	PenumbralRadiusKm float64
}

// ClassifySolarEclipse determines the type of solar eclipse at the given TDB
// Julian date from a geocentric perspective. The caller is responsible for
// providing a time near new moon (Sun-Moon conjunction).
//
// The algorithm is the inverse of the lunar eclipse shadow cone geometry:
// instead of Earth's shadow on the Moon, it computes the Moon's shadow cones
// (umbral and penumbral) and checks whether they intersect Earth's disk at the
// given time. This correctly handles the total vs annular distinction, which
// depends on whether the Moon's umbral cone vertex extends past Earth's surface
// (total) or falls short (annular, producing an antumbra).
func ClassifySolarEclipse(eph *spk.SPK, tdbJD float64) SolarEclipse {
	sunPos := eph.GeocentricPosition(spk.Sun, tdbJD)
	moonPos := eph.GeocentricPosition(spk.Moon, tdbJD)

	// Shadow axis direction: from Sun through Moon (toward Earth).
	shadowDir := [3]float64{
		moonPos[0] - sunPos[0],
		moonPos[1] - sunPos[1],
		moonPos[2] - sunPos[2],
	}
	dSM := vecLen(shadowDir) // Sun-Moon distance
	shadowDir[0] /= dSM
	shadowDir[1] /= dSM
	shadowDir[2] /= dSM

	// Distance from Moon to Earth (origin) along the shadow axis.
	// Vector from Moon to Earth is -moonPos; project onto shadow direction.
	dAlong := -(moonPos[0]*shadowDir[0] + moonPos[1]*shadowDir[1] + moonPos[2]*shadowDir[2])

	result := SolarEclipse{T: tdbJD}

	// If Earth is behind Moon relative to Sun (dAlong < 0), no solar eclipse.
	if dAlong < 0 {
		return result
	}

	// Perpendicular distance from Earth's center to the shadow axis.
	// Nearest point on axis to origin: moonPos + dAlong * shadowDir.
	nearX := moonPos[0] + dAlong*shadowDir[0]
	nearY := moonPos[1] + dAlong*shadowDir[1]
	nearZ := moonPos[2] + dAlong*shadowDir[2]
	perp := math.Sqrt(nearX*nearX + nearY*nearY + nearZ*nearZ)

	// Shadow cone radii at Earth's distance from Moon.
	// Same cone geometry as lunar eclipses but with Moon as the caster:
	//   umbral radius   = R_moon - d * (R_sun - R_moon) / D_sun_moon
	//   penumbral radius = R_moon + d * (R_sun + R_moon) / D_sun_moon
	rUmbra := moonRadiusKm - dAlong*(sunRadiusKm-moonRadiusKm)/dSM
	rPenumbra := moonRadiusKm + dAlong*(sunRadiusKm+moonRadiusKm)/dSM

	result.Gamma = perp / earthRadiusKm
	result.UmbralRadiusKm = rUmbra
	result.PenumbralRadiusKm = rPenumbra

	// Classification: does Earth's disk (radius earthRadiusKm) intersect
	// the shadow cones?
	if perp >= rPenumbra+earthRadiusKm {
		// Penumbra misses Earth entirely.
		return result
	}

	// The central shadow radius is |rUmbra|: positive for umbra (total
	// geometry), negative for antumbra (annular geometry).
	centralRadius := math.Abs(rUmbra)
	if perp < centralRadius+earthRadiusKm {
		if rUmbra > 0 {
			// Umbral cone reaches Earth → total eclipse somewhere on Earth.
			result.Kind = SolarTotal
		} else {
			// Antumbra reaches Earth → annular eclipse somewhere on Earth.
			result.Kind = SolarAnnular
		}
	} else {
		// Only penumbra touches Earth.
		result.Kind = SolarPartial
	}

	return result
}

// vecLen returns the Euclidean length of a 3-vector.
func vecLen(v [3]float64) float64 {
	return math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
}

// eclipticElongation returns the ecliptic longitude difference (Moon - Sun)
// in degrees [0, 360).
func eclipticElongation(moonPos, sunPos [3]float64) float64 {
	// J2000 mean obliquity.
	const obliquitySin = 0.3977771559319137062
	const obliquityCos = 0.9174820620691818140

	moonLon := eclipticLon(moonPos, obliquitySin, obliquityCos)
	sunLon := eclipticLon(sunPos, obliquitySin, obliquityCos)

	diff := moonLon - sunLon
	if diff < 0 {
		diff += 360
	}
	return diff
}

// eclipticLon returns the ecliptic longitude in degrees for an ICRF vector.
func eclipticLon(pos [3]float64, oblSin, oblCos float64) float64 {
	ey := oblCos*pos[1] + oblSin*pos[2]
	ex := pos[0]
	lon := math.Atan2(ey, ex) * 180.0 / math.Pi
	if lon < 0 {
		lon += 360
	}
	return lon
}

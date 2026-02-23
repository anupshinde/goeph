// Package eclipse provides lunar eclipse detection and characterization.
//
// It finds times when the Moon enters Earth's shadow, classifies eclipses as
// penumbral, partial, or total, and computes eclipse magnitudes. Uses the
// Danjon enlargement correction (2% atmospheric enlargement of Earth's shadow).
package eclipse

import (
	"math"

	"github.com/anupshinde/goeph/search"
	"github.com/anupshinde/goeph/spk"
)

const (
	// Eclipse type constants returned in LunarEclipse.Kind.
	Penumbral = 1 // Moon enters penumbra only
	Partial   = 2 // Moon partially enters umbra
	Total     = 3 // Moon fully within umbra

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
		ecl := classifyEclipse(eph, best.T)
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

// classifyEclipse computes the full eclipse geometry at a given time and
// returns a LunarEclipse if the Moon is at least partially in the penumbra.
func classifyEclipse(eph *spk.SPK, tdbJD float64) LunarEclipse {
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

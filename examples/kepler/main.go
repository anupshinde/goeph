// Example: Kepler Orbit Propagation
//
// Demonstrates propagating Keplerian orbits for Ceres (asteroid) and
// Halley's Comet. Shows heliocentric distances at various times.
package main

import (
	"fmt"
	"math"

	"github.com/anupshinde/goeph/kepler"
)

func main() {
	// --- Ceres (asteroid) ---
	// Orbital elements from MPC, ecliptic J2000.
	ceres := &kepler.Orbit{
		SemiMajorAxisAU: 2.7670463,
		Eccentricity:    0.0785115,
		InclinationDeg:  10.5868,
		LongAscNodeDeg:  80.3055,
		ArgPeriapsisDeg: 73.5977,
		MeanAnomalyDeg:  77.372,
		EpochJD:         2451545.0, // J2000.0
	}

	fmt.Println("Ceres heliocentric distance (AU):")
	for year := 2000; year <= 2010; year++ {
		jd := 2451545.0 + float64(year-2000)*365.25
		pos := ceres.PositionAU(jd)
		dist := math.Sqrt(pos[0]*pos[0] + pos[1]*pos[1] + pos[2]*pos[2])
		fmt.Printf("  %d: %.4f AU\n", year, dist)
	}

	// --- Halley's Comet ---
	// Orbital elements, ecliptic J2000. Perihelion on 1986-02-09.
	halley := &kepler.Orbit{
		PerihelionAU:    0.586,
		Eccentricity:    0.9671,
		InclinationDeg:  162.26,
		LongAscNodeDeg:  58.42,
		ArgPeriapsisDeg: 111.33,
		PeriapsisTimeJD: 2446467.395, // 1986-02-09
	}

	fmt.Println("\nHalley's Comet heliocentric distance (AU):")
	for _, yearOffset := range []float64{-5, -2, -1, 0, 1, 2, 5, 10, 20} {
		jd := halley.PeriapsisTimeJD + yearOffset*365.25
		pos := halley.PositionAU(jd)
		dist := math.Sqrt(pos[0]*pos[0] + pos[1]*pos[1] + pos[2]*pos[2])
		label := "periapsis"
		if yearOffset != 0 {
			label = fmt.Sprintf("%+.0fy", yearOffset)
		}
		fmt.Printf("  %s: %.3f AU\n", label, dist)
	}
}

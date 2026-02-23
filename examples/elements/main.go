// Example: Computing osculating orbital elements
//
// Demonstrates computing Keplerian orbital elements from a state vector
// (position + velocity). Uses Earth's orbit around the Sun as an example.
package main

import (
	"fmt"
	"math"
	"time"

	"github.com/anupshinde/goeph/elements"
	"github.com/anupshinde/goeph/spk"
	"github.com/anupshinde/goeph/timescale"
)

func main() {
	eph, err := spk.Open("data/de440s.bsp")
	if err != nil {
		panic(err)
	}

	t := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	jdUTC := timescale.TimeToJDUTC(t)
	jdTT := timescale.UTCToTT(jdUTC)

	fmt.Printf("Date: %s\n\n", t.Format("2006-01-02 15:04 UTC"))

	// Get Earth's position and velocity relative to the Sun
	// Earth barycenter position relative to SSB
	earthPos := eph.GeocentricPosition(spk.Sun, jdTT)
	// Negate: Sun→Earth becomes Earth→Sun, which is Earth's heliocentric position
	// Actually we need Earth relative to Sun, so use the Sun position from Earth
	// and negate it to get Earth's heliocentric position
	pos := [3]float64{-earthPos[0], -earthPos[1], -earthPos[2]}

	// Get Earth's velocity (we use the velocity function)
	// EarthVelocity gives Earth's barycentric velocity, but for elements
	// relative to the Sun we'd need heliocentric velocity.
	// For this example, we'll use a simple approximation via finite difference
	dt := 0.01 // days
	earthPos2 := eph.GeocentricPosition(spk.Sun, jdTT+dt)
	vel := [3]float64{
		(-earthPos2[0] - pos[0]) / (dt * timescale.SecPerDay),
		(-earthPos2[1] - pos[1]) / (dt * timescale.SecPerDay),
		(-earthPos2[2] - pos[2]) / (dt * timescale.SecPerDay),
	}

	// Sun's gravitational parameter (GM) in km³/s²
	muSun := 1.32712440018e11

	elem := elements.FromStateVector(pos, vel, muSun)

	fmt.Println("Earth's osculating orbital elements (heliocentric):")
	fmt.Printf("  Semi-major axis:    %.6f AU\n", elem.SemiMajorAxisKm/149597870.7)
	fmt.Printf("  Eccentricity:       %.8f\n", elem.Eccentricity)
	fmt.Printf("  Inclination:        %.4f°\n", elem.InclinationDeg)
	fmt.Printf("  Long. asc. node:    %.4f°\n", elem.LongAscNodeDeg)
	fmt.Printf("  Arg. perihelion:    %.4f°\n", elem.ArgPeriapsisDeg)
	fmt.Printf("  True anomaly:       %.4f°\n", elem.TrueAnomalyDeg)
	fmt.Printf("  Mean anomaly:       %.4f°\n", elem.MeanAnomalyDeg)
	fmt.Printf("  Orbital period:     %.2f days\n", elem.PeriodDays)
	fmt.Printf("  Perihelion dist:    %.6f AU\n", elem.PeriapsisDistanceKm/149597870.7)
	fmt.Printf("  Aphelion dist:      %.6f AU\n", elem.ApoapsisDistanceKm/149597870.7)

	// Verify: Earth's orbit should have a ≈ 1 AU, e ≈ 0.0167
	a_au := elem.SemiMajorAxisKm / 149597870.7
	if math.Abs(a_au-1.0) > 0.01 {
		fmt.Printf("\nWarning: semi-major axis %.4f AU is far from expected 1 AU\n", a_au)
	}
}

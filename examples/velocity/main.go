// Example: Computing Earth's velocity for aberration
//
// Demonstrates velocity computation from Chebyshev polynomial derivatives.
// Earth's velocity relative to the solar system barycenter is needed for
// stellar aberration correction in apparent position calculations.
package main

import (
	"fmt"
	"math"
	"time"

	"github.com/anupshinde/goeph/spk"
	"github.com/anupshinde/goeph/timescale"
)

func main() {
	eph, err := spk.Open("data/de440s.bsp")
	if err != nil {
		panic(err)
	}

	// Pick a date: 2024-03-20 (vernal equinox)
	t := time.Date(2024, 3, 20, 12, 0, 0, 0, time.UTC)
	jdUTC := timescale.TimeToJDUTC(t)
	jdTT := timescale.UTCToTT(jdUTC)

	// Get Earth's velocity relative to the solar system barycenter
	vel := eph.EarthVelocity(jdTT)

	speed := math.Sqrt(vel[0]*vel[0] + vel[1]*vel[1] + vel[2]*vel[2])
	speedKmPerSec := speed / timescale.SecPerDay

	fmt.Printf("Date: %s\n", t.Format("2006-01-02 15:04 UTC"))
	fmt.Printf("Earth velocity (km/day): [%.3f, %.3f, %.3f]\n",
		vel[0], vel[1], vel[2])
	fmt.Printf("Earth speed: %.3f km/s (%.6f c)\n",
		speedKmPerSec, speedKmPerSec/299792.458)
	fmt.Println("\nEarth orbits at ~30 km/s, causing up to ~20 arcsec")
	fmt.Println("of stellar aberration in apparent positions.")
}

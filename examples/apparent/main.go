// Example: Apparent positions with aberration and deflection
//
// Demonstrates computing apparent positions, which include light-time
// correction, stellar aberration (special relativity), and gravitational
// light deflection (general relativity) by the Sun, Jupiter, and Saturn.
package main

import (
	"fmt"
	"math"
	"time"

	"github.com/anupshinde/goeph/coord"
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

	bodies := []struct {
		name string
		id   int
	}{
		{"Mars", spk.MarsBarycenter},
		{"Jupiter", spk.JupiterBarycenter},
		{"Saturn", spk.SaturnBarycenter},
	}

	for _, b := range bodies {
		// Astrometric (light-time corrected only)
		astro := eph.Observe(b.id, jdTT)
		// Apparent (light-time + aberration + deflection)
		app := eph.Apparent(b.id, jdTT)

		// Difference shows the aberration + deflection effect
		dx := app[0] - astro[0]
		dy := app[1] - astro[1]
		dz := app[2] - astro[2]
		shift := math.Sqrt(dx*dx+dy*dy+dz*dz) / 149597870.7 * 206265 // arcsec

		// Convert apparent position to RA/Dec
		latEcl, lonEcl := coord.ICRFToEcliptic(app[0], app[1], app[2])

		fmt.Printf("%s:\n", b.name)
		fmt.Printf("  Apparent ecliptic: lat=%.4f° lon=%.4f°\n", latEcl, lonEcl)
		fmt.Printf("  Aberration+deflection shift: %.2f arcsec\n\n", shift)
	}
}

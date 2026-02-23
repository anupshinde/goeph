// Example: Computing planetary visual magnitudes
//
// Demonstrates computing apparent visual magnitudes for planets using the
// Mallama & Hilton 2018 phase curve models. Requires phase angle and
// distances from the Sun and observer to the planet.
package main

import (
	"fmt"
	"math"
	"time"

	"github.com/anupshinde/goeph/coord"
	"github.com/anupshinde/goeph/magnitude"
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
	fmt.Printf("%-8s  %6s  %8s  %8s  %6s\n",
		"Planet", "Phase°", "Sun AU", "Earth AU", "V mag")
	fmt.Println("-----------------------------------------------")

	au := 149597870.7

	planets := []struct {
		name string
		id   int
	}{
		{"Mercury", spk.Mercury},
		{"Venus", spk.Venus},
		{"Mars", spk.MarsBarycenter},
		{"Jupiter", spk.JupiterBarycenter},
		{"Saturn", spk.SaturnBarycenter},
	}

	for _, p := range planets {
		// Get positions
		sunPos := eph.GeocentricPosition(spk.Sun, jdTT) // Earth→Sun
		planetPos := eph.Observe(p.id, jdTT)             // Earth→Planet

		// Sun→Planet vector = Planet_from_Earth - Sun_from_Earth
		sunToPlanet := [3]float64{
			planetPos[0] - sunPos[0],
			planetPos[1] - sunPos[1],
			planetPos[2] - sunPos[2],
		}

		// Distances
		rSun := math.Sqrt(sunToPlanet[0]*sunToPlanet[0]+
			sunToPlanet[1]*sunToPlanet[1]+
			sunToPlanet[2]*sunToPlanet[2]) / au
		rEarth := math.Sqrt(planetPos[0]*planetPos[0]+
			planetPos[1]*planetPos[1]+
			planetPos[2]*planetPos[2]) / au

		// Phase angle
		phase := coord.PhaseAngle(planetPos, sunPos)

		mag := magnitude.PlanetaryMagnitude(p.id, phase, rSun, rEarth)

		fmt.Printf("%-8s  %6.1f  %8.4f  %8.4f  %+6.1f\n",
			p.name, phase, rSun, rEarth, mag)
	}
}

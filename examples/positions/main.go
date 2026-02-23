// Example: Loading an SPK ephemeris and computing planet positions
//
// Demonstrates the core workflow: load a JPL BSP file, convert a date to
// the TDB time scale, and query the geocentric position of Mars.
// Positions are light-time corrected and returned in ICRF (J2000) km.
package main

import (
	"fmt"
	"math"
	"time"

	"github.com/anupshinde/goeph/spk"
	"github.com/anupshinde/goeph/timescale"
)

func main() {
	// Load the JPL DE440s ephemeris (covers 1849-2150)
	eph, err := spk.Open("data/de440s.bsp")
	if err != nil {
		panic(err)
	}

	// Pick a date: 2024-06-15 12:00 UTC
	t := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	// Convert UTC -> TT (used as TDB for ephemeris lookup)
	jdUTC := timescale.TimeToJDUTC(t)
	jdTT := timescale.UTCToTT(jdUTC)

	// Get light-time corrected geocentric position of Mars
	marsPos := eph.Observe(spk.MarsBarycenter, jdTT)
	fmt.Printf("Date: %s\n", t.Format("2006-01-02 15:04 UTC"))
	fmt.Printf("Mars ICRF position (km): [%.3f, %.3f, %.3f]\n",
		marsPos[0], marsPos[1], marsPos[2])

	// Also get geometric (no light-time correction) position for comparison
	marsGeo := eph.GeocentricPosition(spk.MarsBarycenter, jdTT)
	fmt.Printf("Mars geometric pos (km): [%.3f, %.3f, %.3f]\n",
		marsGeo[0], marsGeo[1], marsGeo[2])

	// Show positions for several planets
	bodies := []struct {
		name string
		id   int
	}{
		{"Mercury", spk.Mercury},
		{"Venus", spk.Venus},
		{"Mars", spk.MarsBarycenter},
		{"Jupiter", spk.JupiterBarycenter},
		{"Saturn", spk.SaturnBarycenter},
	}
	fmt.Println("\nAll planet distances from Earth:")
	for _, b := range bodies {
		pos := eph.Observe(b.id, jdTT)
		dist := 0.0
		for i := 0; i < 3; i++ {
			dist += pos[i] * pos[i]
		}
		dist = math.Sqrt(dist)
		fmt.Printf("  %-8s %15.1f km  (%.4f AU)\n", b.name, dist, dist/149597870.7)
	}
}
// Example: Stereographic projection of planet positions
//
// Demonstrates projecting planet positions onto a 2D plane centered at
// the north ecliptic pole. Stereographic projection is conformal (preserves
// angles) and maps the center direction to the origin (0, 0).
package main

import (
	"fmt"
	"math"
	"time"

	"github.com/anupshinde/goeph/projection"
	"github.com/anupshinde/goeph/spk"
	"github.com/anupshinde/goeph/timescale"
)

func main() {
	eph, err := spk.Open("data/de440s.bsp")
	if err != nil {
		panic(err)
	}

	t := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	jdTT := timescale.UTCToTT(timescale.TimeToJDUTC(t))

	// Project centered at the Sun's position (solar-centric view of the sky).
	sunPos := eph.Observe(spk.Sun, jdTT)
	proj := projection.NewProjector(sunPos[0], sunPos[1], sunPos[2])

	fmt.Printf("Stereographic projection centered on the Sun\n")
	fmt.Printf("Date: %s\n\n", t.Format("2006-01-02"))
	fmt.Printf("%-10s %10s %10s %10s\n", "Body", "x", "y", "dist from center")
	fmt.Println("---------- ---------- ---------- ----------------")

	bodies := []struct {
		name string
		id   int
	}{
		{"Sun", spk.Sun},
		{"Moon", spk.Moon},
		{"Mercury", spk.MercuryBarycenter},
		{"Venus", spk.VenusBarycenter},
		{"Mars", spk.MarsBarycenter},
		{"Jupiter", spk.JupiterBarycenter},
		{"Saturn", spk.SaturnBarycenter},
	}

	for _, body := range bodies {
		pos := eph.Observe(body.id, jdTT)
		x, y := proj.Project(pos[0], pos[1], pos[2])
		dist := math.Sqrt(x*x + y*y)
		fmt.Printf("%-10s %10.6f %10.6f %10.6f\n", body.name, x, y, dist)
	}
}

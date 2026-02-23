// Example: Phase angle and fraction illuminated
//
// Computes the phase angle (Sun-target-observer angle) and the fraction
// of the disc that is illuminated for the Moon and visible planets.
// Phase angle 0° = fully lit (opposition), 180° = fully dark (conjunction).
package main

import (
	"fmt"
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
	jdTT := timescale.UTCToTT(timescale.TimeToJDUTC(t))

	fmt.Printf("Date: %s\n\n", t.Format("2006-01-02 15:04 UTC"))

	sunPos := eph.Observe(spk.Sun, jdTT)

	bodies := []struct {
		name string
		id   int
	}{
		{"Moon", spk.Moon},
		{"Venus", spk.Venus},
		{"Mars", spk.MarsBarycenter},
		{"Jupiter", spk.JupiterBarycenter},
		{"Saturn", spk.SaturnBarycenter},
	}

	fmt.Printf("%-10s %12s %12s\n", "Body", "Phase (°)", "Illuminated")
	fmt.Println("---------- ------------ ------------")
	for _, b := range bodies {
		bodyPos := eph.Observe(b.id, jdTT)

		// Sun-to-target vector = body position - sun position (geocentric)
		sunToTarget := [3]float64{
			bodyPos[0] - sunPos[0],
			bodyPos[1] - sunPos[1],
			bodyPos[2] - sunPos[2],
		}

		phase := coord.PhaseAngle(bodyPos, sunToTarget)
		illum := coord.FractionIlluminated(phase)

		fmt.Printf("%-10s %11.2f° %11.1f%%\n", b.name, phase, illum*100)
	}
}

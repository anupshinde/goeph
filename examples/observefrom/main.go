// Example: Observing from an arbitrary body
//
// Demonstrates computing positions as seen from bodies other than Earth.
// ObserveFrom and ApparentFrom allow any body in the ephemeris to be used
// as the observer.
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

	t := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	jdUTC := timescale.TimeToJDUTC(t)
	jdTT := timescale.UTCToTT(jdUTC)

	fmt.Printf("Date: %s\n\n", t.Format("2006-01-02 15:04 UTC"))

	// Distance from Earth to Mars
	earthToMars := eph.ObserveFrom(spk.Earth, spk.MarsBarycenter, jdTT)
	distEM := math.Sqrt(earthToMars[0]*earthToMars[0] +
		earthToMars[1]*earthToMars[1] +
		earthToMars[2]*earthToMars[2])

	// Distance from Jupiter to Mars
	jupToMars := eph.ObserveFrom(spk.JupiterBarycenter, spk.MarsBarycenter, jdTT)
	distJM := math.Sqrt(jupToMars[0]*jupToMars[0] +
		jupToMars[1]*jupToMars[1] +
		jupToMars[2]*jupToMars[2])

	fmt.Printf("Earth to Mars:   %12.1f km  (%.4f AU)\n",
		distEM, distEM/149597870.7)
	fmt.Printf("Jupiter to Mars: %12.1f km  (%.4f AU)\n\n",
		distJM, distJM/149597870.7)

	// Apparent position from different observers
	appFromEarth := eph.ApparentFrom(spk.Earth, spk.MarsBarycenter, jdTT)
	appFromJup := eph.ApparentFrom(spk.JupiterBarycenter, spk.MarsBarycenter, jdTT)

	fmt.Println("Mars apparent position (ICRF km):")
	fmt.Printf("  From Earth:   [%.1f, %.1f, %.1f]\n",
		appFromEarth[0], appFromEarth[1], appFromEarth[2])
	fmt.Printf("  From Jupiter: [%.1f, %.1f, %.1f]\n",
		appFromJup[0], appFromJup[1], appFromJup[2])
}

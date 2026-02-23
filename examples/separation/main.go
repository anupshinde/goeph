// Example: Angular separation between the Sun and Moon
//
// Computes the angular distance between two bodies as seen from Earth.
// Uses Kahan's numerically stable formula for the angle between two
// position vectors.
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

	sunPos := eph.Observe(spk.Sun, jdTT)
	moonPos := eph.Observe(spk.Moon, jdTT)

	sep := coord.SeparationAngle(sunPos, moonPos)

	fmt.Printf("Date: %s\n", t.Format("2006-01-02 15:04 UTC"))
	fmt.Printf("Sun-Moon separation: %.4f°\n", sep)

	// Show separations over a lunar month
	fmt.Println("\nSun-Moon separation over one lunar month:")
	t0 := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	for day := 0; day <= 30; day += 3 {
		ti := t0.AddDate(0, 0, day)
		jd := timescale.UTCToTT(timescale.TimeToJDUTC(ti))
		sun := eph.Observe(spk.Sun, jd)
		moon := eph.Observe(spk.Moon, jd)
		s := coord.SeparationAngle(sun, moon)
		fmt.Printf("  %s: %7.2f°\n", ti.Format("Jan 02"), s)
	}
}

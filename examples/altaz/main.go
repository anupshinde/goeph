// Example: Altitude and azimuth for a ground observer
//
// Computes the altitude (elevation above horizon) and azimuth (compass
// direction, 0=North, 90=East) of celestial bodies as seen from a specific
// location on Earth's surface.
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

	// Observer location: New York City
	lat, lon := 40.7128, -74.0060
	fmt.Printf("Observer: New York City (%.4f°N, %.4f°E)\n", lat, lon)

	t := time.Date(2024, 6, 15, 22, 0, 0, 0, time.UTC) // evening
	jdUTC := timescale.TimeToJDUTC(t)
	jdTT := timescale.UTCToTT(jdUTC)
	jdUT1 := timescale.TTToUT1(jdTT)

	fmt.Printf("Date: %s\n\n", t.Format("2006-01-02 15:04 UTC"))

	bodies := []struct {
		name string
		id   int
	}{
		{"Sun", spk.Sun},
		{"Moon", spk.Moon},
		{"Mars", spk.MarsBarycenter},
		{"Jupiter", spk.JupiterBarycenter},
	}

	for _, b := range bodies {
		pos := eph.Apparent(b.id, jdTT)
		alt, az, _ := coord.Altaz(pos, lat, lon, jdUT1)

		status := "above horizon"
		if alt < 0 {
			status = "below horizon"
		}
		fmt.Printf("%-8s  alt=%+7.2f°  az=%6.2f°  (%s)\n",
			b.name, alt, az, status)
	}

	// Also show refraction correction for a body near the horizon
	fmt.Println("\nNote: altitudes are geometric (no atmospheric refraction).")
	fmt.Println("Use coord.Refract() to add refraction correction near the horizon.")
}

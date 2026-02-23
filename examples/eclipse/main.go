// Example: Lunar Eclipses
//
// Finds all lunar eclipses in 2024 and prints their type, date, and magnitude.
package main

import (
	"fmt"
	"time"

	"github.com/anupshinde/goeph/eclipse"
	"github.com/anupshinde/goeph/spk"
	"github.com/anupshinde/goeph/timescale"
)

var kindNames = map[int]string{
	eclipse.Penumbral: "Penumbral",
	eclipse.Partial:   "Partial",
	eclipse.Total:     "Total",
}

func main() {
	eph, err := spk.Open("data/de440s.bsp")
	if err != nil {
		panic(err)
	}

	// Search for lunar eclipses in 2024.
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	startJD := timescale.UTCToTT(timescale.TimeToJDUTC(start))
	endJD := timescale.UTCToTT(timescale.TimeToJDUTC(end))

	eclipses, err := eclipse.FindLunarEclipses(eph, startJD, endJD)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Lunar eclipses in 2024 (%d found):\n\n", len(eclipses))
	for _, e := range eclipses {
		t := jdTTToTime(e.T)
		fmt.Printf("  %s  %s\n", t.Format("2006-01-02 15:04 MST"), kindNames[e.Kind])
		fmt.Printf("    Umbral magnitude:    %.4f\n", e.UmbralMag)
		fmt.Printf("    Penumbral magnitude: %.4f\n", e.PenumbralMag)
		fmt.Printf("    Closest approach:    %.0f km\n", e.ClosestApproachKm)
		fmt.Printf("    Umbral radius:       %.0f km\n", e.UmbralRadiusKm)
		fmt.Printf("    Penumbral radius:    %.0f km\n\n", e.PenumbralRadiusKm)
	}
}

func jdTTToTime(jdTT float64) time.Time {
	jdUTC := jdTT - 69.184/86400.0
	daysSinceJ2000 := jdUTC - 2451545.0
	j2000 := time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)
	return j2000.Add(time.Duration(daysSinceJ2000 * 24 * float64(time.Hour)))
}

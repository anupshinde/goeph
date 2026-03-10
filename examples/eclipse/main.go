// Example: Lunar and Solar Eclipses
//
// Finds all lunar eclipses in 2024 and classifies known solar eclipses.
package main

import (
	"fmt"
	"time"

	"github.com/anupshinde/goeph/eclipse"
	"github.com/anupshinde/goeph/spk"
	"github.com/anupshinde/goeph/timescale"
)

var lunarKindNames = map[int]string{
	eclipse.Penumbral: "Penumbral",
	eclipse.Partial:   "Partial",
	eclipse.Total:     "Total",
}

var solarKindNames = map[int]string{
	0:                    "None",
	eclipse.SolarPartial: "Partial",
	eclipse.SolarAnnular: "Annular",
	eclipse.SolarTotal:   "Total",
}

func main() {
	eph, err := spk.Open("data/de440s.bsp")
	if err != nil {
		panic(err)
	}

	// --- Lunar eclipses ---

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
		fmt.Printf("  %s  %s\n", t.Format("2006-01-02 15:04 MST"), lunarKindNames[e.Kind])
		fmt.Printf("    Umbral magnitude:    %.4f\n", e.UmbralMag)
		fmt.Printf("    Penumbral magnitude: %.4f\n", e.PenumbralMag)
		fmt.Printf("    Closest approach:    %.0f km\n", e.ClosestApproachKm)
		fmt.Printf("    Umbral radius:       %.0f km\n", e.UmbralRadiusKm)
		fmt.Printf("    Penumbral radius:    %.0f km\n\n", e.PenumbralRadiusKm)
	}

	// --- Solar eclipse classification ---

	// Classify known solar eclipses at their approximate maximum times.
	fmt.Println("Solar eclipse classification:")
	fmt.Println()

	solarTests := []struct {
		name string
		t    time.Time
	}{
		{"2017-08-21 Total", time.Date(2017, 8, 21, 18, 26, 0, 0, time.UTC)},
		{"2023-10-14 Annular", time.Date(2023, 10, 14, 18, 0, 0, 0, time.UTC)},
		{"2024-04-08 Total", time.Date(2024, 4, 8, 18, 18, 0, 0, time.UTC)},
	}

	for _, st := range solarTests {
		jd := timescale.UTCToTT(timescale.TimeToJDUTC(st.t))
		se := eclipse.ClassifySolarEclipse(eph, jd)
		fmt.Printf("  %s → %s (gamma=%.3f, umbral radius=%.1f km)\n",
			st.name, solarKindNames[se.Kind], se.Gamma, se.UmbralRadiusKm)
	}
}

func jdTTToTime(jdTT float64) time.Time {
	jdUTC := jdTT - 69.184/86400.0
	daysSinceJ2000 := jdUTC - 2451545.0
	j2000 := time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)
	return j2000.Add(time.Duration(daysSinceJ2000 * 24 * float64(time.Hour)))
}

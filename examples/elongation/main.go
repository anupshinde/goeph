// Example: Moon elongation and lunar phase
//
// Elongation is the angular distance of the Moon from the Sun measured
// along the ecliptic. It determines the Moon's phase:
//   0° = New Moon, 90° = First Quarter, 180° = Full Moon, 270° = Last Quarter
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

	// Show the Moon's elongation over one lunar month
	fmt.Println("Moon elongation and phase over one month:")
	fmt.Printf("%-12s %12s  %s\n", "Date", "Elongation", "Phase")
	fmt.Println("------------ ------------ ----------------")

	t0 := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	for day := 0; day <= 29; day++ {
		ti := t0.AddDate(0, 0, day)
		jdTT := timescale.UTCToTT(timescale.TimeToJDUTC(ti))

		sunPos := eph.Observe(spk.Sun, jdTT)
		moonPos := eph.Observe(spk.Moon, jdTT)

		// Get ecliptic longitudes
		_, sunLon := coord.ICRFToEcliptic(sunPos[0], sunPos[1], sunPos[2])
		_, moonLon := coord.ICRFToEcliptic(moonPos[0], moonPos[1], moonPos[2])

		elong := coord.Elongation(moonLon, sunLon)
		phase := phaseName(elong)

		fmt.Printf("%s %11.1f°  %s\n", ti.Format("2006-01-02"), elong, phase)
	}
}

func phaseName(elongDeg float64) string {
	switch {
	case elongDeg < 22.5 || elongDeg >= 337.5:
		return "New Moon"
	case elongDeg < 67.5:
		return "Waxing Crescent"
	case elongDeg < 112.5:
		return "First Quarter"
	case elongDeg < 157.5:
		return "Waxing Gibbous"
	case elongDeg < 202.5:
		return "Full Moon"
	case elongDeg < 247.5:
		return "Waning Gibbous"
	case elongDeg < 292.5:
		return "Last Quarter"
	default:
		return "Waning Crescent"
	}
}

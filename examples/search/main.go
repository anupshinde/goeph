// Example: Finding season changes with event search
//
// Demonstrates using the search package's FindDiscrete to locate times
// when the Sun's ecliptic longitude crosses quarter boundaries, i.e.,
// equinoxes and solstices.
package main

import (
	"fmt"
	"math"
	"time"

	"github.com/anupshinde/goeph/coord"
	"github.com/anupshinde/goeph/search"
	"github.com/anupshinde/goeph/spk"
	"github.com/anupshinde/goeph/timescale"
)

var seasonNames = [4]string{
	"Spring equinox",
	"Summer solstice",
	"Autumn equinox",
	"Winter solstice",
}

func main() {
	eph, err := spk.Open("data/de440s.bsp")
	if err != nil {
		panic(err)
	}

	// Search year 2024 for season changes.
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	startJD := timescale.UTCToTT(timescale.TimeToJDUTC(start))
	endJD := timescale.UTCToTT(timescale.TimeToJDUTC(end))

	// Discrete function: which quarter of the ecliptic is the Sun in?
	// 0 = [0°,90°), 1 = [90°,180°), 2 = [180°,270°), 3 = [270°,360°)
	seasonFunc := func(tdbJD float64) int {
		sunPos := eph.Apparent(spk.Sun, tdbJD)
		_, lonDeg := coord.ICRFToEcliptic(sunPos[0], sunPos[1], sunPos[2])
		if lonDeg < 0 {
			lonDeg += 360.0
		}
		return int(math.Floor(lonDeg / 90.0))
	}

	// Step of 90 days is sufficient — seasons last ~91 days.
	events, err := search.FindDiscrete(startJD, endJD, 90.0, seasonFunc, 0)
	if err != nil {
		panic(err)
	}

	fmt.Println("Seasons of 2024:")
	for _, e := range events {
		utc := jdTTToTime(e.T)
		fmt.Printf("  %-20s  %s\n", seasonNames[e.NewValue], utc.Format("2006-01-02 15:04:05 UTC"))
	}
}

// jdTTToTime converts a Julian date (TT) back to an approximate UTC time.Time.
func jdTTToTime(jdTT float64) time.Time {
	// TT - UTC ≈ 69.184 seconds (since 2017)
	jdUTC := jdTT - 69.184/86400.0
	// JD 2451545.0 = 2000-01-01T12:00:00 UTC
	daysSinceJ2000 := jdUTC - 2451545.0
	j2000 := time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)
	return j2000.Add(time.Duration(daysSinceJ2000 * 24 * float64(time.Hour)))
}

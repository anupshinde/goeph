// Example: Almanac — sunrise/sunset, moon phases, and seasons
//
// Demonstrates the almanac package's event-finding functions for a ground
// observer. Shows sunrise/sunset times for a week, moon phases for a month,
// and the next solstice/equinox.
package main

import (
	"fmt"
	"time"

	"github.com/anupshinde/goeph/almanac"
	"github.com/anupshinde/goeph/spk"
	"github.com/anupshinde/goeph/timescale"
)

var seasonNames = [4]string{"Spring equinox", "Summer solstice", "Autumn equinox", "Winter solstice"}
var phaseNames = [4]string{"New Moon", "First Quarter", "Full Moon", "Last Quarter"}

func main() {
	eph, err := spk.Open("data/de440s.bsp")
	if err != nil {
		panic(err)
	}

	// NYC observer.
	lat, lon := 40.7128, -74.0060

	// --- Sunrise/Sunset for one week ---
	start := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 7)
	startJD := timescale.UTCToTT(timescale.TimeToJDUTC(start))
	endJD := timescale.UTCToTT(timescale.TimeToJDUTC(end))

	events, err := almanac.SunriseSunset(eph, lat, lon, startJD, endJD)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Sunrise/Sunset for NYC (%.1f°N, %.1f°W), Jun 15-22 2024:\n", lat, -lon)
	for _, e := range events {
		t := jdTTToTime(e.T)
		kind := "Sunset "
		if e.NewValue == 1 {
			kind = "Sunrise"
		}
		fmt.Printf("  %s  %s\n", kind, t.Format("Mon Jan 02 15:04 MST"))
	}

	// --- Moon phases for a month ---
	fmt.Println()
	start2 := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	end2 := start2.AddDate(0, 1, 0)
	startJD2 := timescale.UTCToTT(timescale.TimeToJDUTC(start2))
	endJD2 := timescale.UTCToTT(timescale.TimeToJDUTC(end2))

	phases, err := almanac.MoonPhases(eph, startJD2, endJD2)
	if err != nil {
		panic(err)
	}

	fmt.Println("Moon phases, June 2024:")
	for _, e := range phases {
		t := jdTTToTime(e.T)
		fmt.Printf("  %-15s  %s\n", phaseNames[e.NewValue], t.Format("Mon Jan 02 15:04 MST"))
	}

	// --- Seasons for 2024 ---
	fmt.Println()
	start3 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end3 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	startJD3 := timescale.UTCToTT(timescale.TimeToJDUTC(start3))
	endJD3 := timescale.UTCToTT(timescale.TimeToJDUTC(end3))

	seasons, err := almanac.Seasons(eph, startJD3, endJD3)
	if err != nil {
		panic(err)
	}

	fmt.Println("Seasons of 2024:")
	for _, e := range seasons {
		t := jdTTToTime(e.T)
		fmt.Printf("  %-20s  %s\n", seasonNames[e.NewValue], t.Format("2006-01-02 15:04 MST"))
	}
}

// jdTTToTime converts a Julian date (TT) back to an approximate UTC time.Time.
func jdTTToTime(jdTT float64) time.Time {
	// TT - UTC ≈ 69.184 seconds (since 2017)
	jdUTC := jdTT - 69.184/86400.0
	daysSinceJ2000 := jdUTC - 2451545.0
	j2000 := time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)
	return j2000.Add(time.Duration(daysSinceJ2000 * 24 * float64(time.Hour)))
}

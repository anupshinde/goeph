// Example: SGP4 satellite propagation
//
// Propagates the ISS orbit using SGP4 from a Two-Line Element (TLE) set.
// Shows the sub-satellite point (latitude/longitude) at several times.
package main

import (
	"fmt"
	"time"

	"github.com/anupshinde/goeph/satellite"
)

func main() {
	// ISS TLE (epoch: 2024-01-01)
	iss := satellite.NewSat(
		"ISS (ZARYA)",
		"1 25544U 98067A   24001.00000000  .00016717  00000-0  10270-3 0  9005",
		"2 25544  51.6400 208.9163 0006703 247.1970 112.8444 15.49560830999999",
	)
	fmt.Printf("Satellite: %s\n\n", iss.Name)

	// Track the ISS over 2 hours in 15-minute steps
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	fmt.Printf("%-20s %10s %10s\n", "Time (UTC)", "Lat (°)", "Lon (°)")
	fmt.Println("-------------------- ---------- ----------")

	for minutes := 0; minutes <= 120; minutes += 15 {
		t := t0.Add(time.Duration(minutes) * time.Minute)
		lat, lon := satellite.SubPoint(iss.Sat, t)

		// Convert longitude to -180..180 for readability
		if lon > 180 {
			lon -= 360
		}

		fmt.Printf("%s %9.2f° %9.2f°\n",
			t.Format("2006-01-02 15:04:05"), lat, lon)
	}

	fmt.Println("\nNote: ISS orbits at ~51.6° inclination, completing")
	fmt.Println("one orbit every ~92 minutes.")
}

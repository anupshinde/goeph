// Example: Observer on Earth's surface
//
// Converts a geodetic position (latitude/longitude on Earth) to an ICRF
// direction vector, which can then be used to compute where on the sky
// a ground-based observer is looking. The conversion accounts for Earth
// rotation, precession, and nutation.
package main

import (
	"fmt"
	"time"

	"github.com/anupshinde/goeph/coord"
	"github.com/anupshinde/goeph/timescale"
)

func main() {
	// Observatory locations
	observatories := []struct {
		name    string
		latDeg  float64
		lonDeg  float64
	}{
		{"Greenwich, UK", 51.4769, -0.0005},
		{"Mauna Kea, Hawaii", 19.8208, -155.4681},
		{"Paranal, Chile", -24.6272, -70.4048},
		{"Sydney, Australia", -33.8688, 151.2093},
	}

	t := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	jdUTC := timescale.TimeToJDUTC(t)
	jdTT := timescale.UTCToTT(jdUTC)
	jdUT1 := timescale.TTToUT1(jdTT)

	fmt.Printf("Date: %s\n\n", t.Format("2006-01-02 15:04 UTC"))

	for _, obs := range observatories {
		// Get zenith direction in ICRF
		x, y, z := coord.GeodeticToICRF(obs.latDeg, obs.lonDeg, jdUT1)

		// Convert to ecliptic to see where zenith points
		eclLat, eclLon := coord.ICRFToEcliptic(x, y, z)

		fmt.Printf("%s (%.2f°, %.2f°):\n", obs.name, obs.latDeg, obs.lonDeg)
		fmt.Printf("  Zenith ICRF:    [%.6f, %.6f, %.6f]\n", x, y, z)
		fmt.Printf("  Zenith ecliptic: lat=%.2f° lon=%.2f°\n\n", eclLat, eclLon)
	}
}

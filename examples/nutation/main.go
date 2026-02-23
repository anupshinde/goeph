// Example: Nutation precision modes
//
// Demonstrates the two nutation precision modes available in goeph:
//   - NutationStandard (default): 30 largest luni-solar terms, ~1 arcsec, fast
//   - NutationFull: 678 luni-solar + 687 planetary terms, ~0.001 arcsec, ~70x slower
//
// Shows the practical difference in computed positions between the two modes.
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

	// Observer: Greenwich, UK
	lat, lon := 51.4769, -0.0005
	fmt.Printf("Observer: Greenwich, UK (%.4f°N, %.4f°E)\n", lat, lon)

	t := time.Date(2024, 6, 15, 22, 0, 0, 0, time.UTC)
	jdUTC := timescale.TimeToJDUTC(t)
	jdTT := timescale.UTCToTT(jdUTC)
	jdUT1 := timescale.TTToUT1(jdTT)

	fmt.Printf("Date: %s\n\n", t.Format("2006-01-02 15:04 UTC"))

	// --- Compare nutation modes for altaz ---
	fmt.Println("=== Altitude/Azimuth comparison ===")
	fmt.Println()

	bodies := []struct {
		name string
		id   int
	}{
		{"Moon", spk.Moon},
		{"Mars", spk.MarsBarycenter},
		{"Jupiter", spk.JupiterBarycenter},
	}

	for _, b := range bodies {
		pos := eph.Apparent(b.id, jdTT)

		coord.SetNutationPrecision(coord.NutationStandard)
		altStd, azStd, _ := coord.Altaz(pos, lat, lon, jdUT1)

		coord.SetNutationPrecision(coord.NutationFull)
		altFull, azFull, _ := coord.Altaz(pos, lat, lon, jdUT1)

		altDiff := altFull - altStd
		azDiff := azFull - azStd

		fmt.Printf("%-8s  Standard: alt=%+8.4f° az=%8.4f°\n", b.name, altStd, azStd)
		fmt.Printf("          Full:     alt=%+8.4f° az=%8.4f°\n", altFull, azFull)
		fmt.Printf("          Diff:     alt=%+.4f\"  az=%+.4f\" (arcseconds)\n\n",
			altDiff*3600, azDiff*3600)
	}

	// --- Compare nutation modes for geodetic→ecliptic ---
	fmt.Println("=== Geodetic→ecliptic comparison ===")
	fmt.Println()

	coord.SetNutationPrecision(coord.NutationStandard)
	xStd, yStd, zStd := coord.GeodeticToICRF(lat, lon, jdUT1)
	eclLatStd, eclLonStd := coord.ICRFToEcliptic(xStd, yStd, zStd)

	coord.SetNutationPrecision(coord.NutationFull)
	xFull, yFull, zFull := coord.GeodeticToICRF(lat, lon, jdUT1)
	eclLatFull, eclLonFull := coord.ICRFToEcliptic(xFull, yFull, zFull)

	fmt.Printf("Zenith ecliptic (Standard): lat=%+.6f° lon=%.6f°\n", eclLatStd, eclLonStd)
	fmt.Printf("Zenith ecliptic (Full):     lat=%+.6f° lon=%.6f°\n", eclLatFull, eclLonFull)
	fmt.Printf("Difference:                 lat=%+.4f\"  lon=%+.4f\" (arcseconds)\n\n",
		(eclLatFull-eclLatStd)*3600, (eclLonFull-eclLonStd)*3600)

	// Restore default
	coord.SetNutationPrecision(coord.NutationStandard)

	fmt.Println("=== Summary ===")
	fmt.Println()
	fmt.Println("NutationStandard (default): 30 terms, ~150 ns/call")
	fmt.Println("NutationFull:               1365 terms, ~10.5 µs/call (~70x slower)")
	fmt.Println()
	fmt.Println("The difference is ~1 arcsec — negligible for most applications.")
	fmt.Println("Dominant error sources (light-time ~20\", GMST formula ~0.3\"/century)")
	fmt.Println("are much larger than the nutation precision difference.")
	fmt.Println()
	fmt.Println("Use NutationFull for: exact Skyfield parity, sub-arcsecond work.")
	fmt.Println("Use NutationStandard for: batch processing, general astronomy.")
}

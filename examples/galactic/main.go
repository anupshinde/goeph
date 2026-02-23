// Example: Converting positions to galactic coordinates
//
// Demonstrates converting ICRF positions to the Galactic coordinate
// system (IAU 1958 System II). The galactic center is at l=0°, b=0°
// and the north galactic pole is at b=90°.
package main

import (
	"fmt"
	"time"

	"github.com/anupshinde/goeph/coord"
	"github.com/anupshinde/goeph/spk"
	"github.com/anupshinde/goeph/star"
	"github.com/anupshinde/goeph/timescale"
)

func main() {
	// Galactic center (Sgr A*) — should map to l~0°, b~0°
	gcX, gcY, gcZ := star.GalacticCenterICRF()
	gcLat, gcLon := coord.ICRFToGalactic(gcX, gcY, gcZ)
	fmt.Println("Galactic Center (Sgr A*):")
	fmt.Printf("  Galactic l=%.2f°, b=%.2f° (expect ~0°, ~0°)\n\n", gcLon, gcLat)

	// Planet positions in galactic coordinates
	eph, err := spk.Open("data/de440s.bsp")
	if err != nil {
		panic(err)
	}

	t := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	jdTT := timescale.UTCToTT(timescale.TimeToJDUTC(t))

	fmt.Printf("Planet galactic coordinates on %s:\n\n", t.Format("2006-01-02"))
	fmt.Printf("%-10s %10s %10s\n", "Body", "l (°)", "b (°)")
	fmt.Println("---------- ---------- ----------")

	bodies := []struct {
		name string
		id   int
	}{
		{"Sun", spk.Sun},
		{"Moon", spk.Moon},
		{"Mars", spk.MarsBarycenter},
		{"Jupiter", spk.JupiterBarycenter},
		{"Saturn", spk.SaturnBarycenter},
	}

	for _, body := range bodies {
		pos := eph.Observe(body.id, jdTT)
		lat, lon := coord.ICRFToGalactic(pos[0], pos[1], pos[2])
		fmt.Printf("%-10s %9.2f° %9.2f°\n", body.name, lon, lat)
	}
}

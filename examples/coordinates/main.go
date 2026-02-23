// Example: Converting positions to RA/Dec and ecliptic coordinates
//
// After getting a planet's ICRF position, convert it to astronomical
// coordinate systems: equatorial (RA/Dec) and ecliptic (lat/lon).
package main

import (
	"fmt"
	"math"
	"time"

	"github.com/anupshinde/goeph/coord"
	"github.com/anupshinde/goeph/spk"
	"github.com/anupshinde/goeph/timescale"
	"github.com/anupshinde/goeph/units"
)

func main() {
	eph, err := spk.Open("data/de440s.bsp")
	if err != nil {
		panic(err)
	}

	t := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	jdTT := timescale.UTCToTT(timescale.TimeToJDUTC(t))

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
		pos := eph.Observe(b.id, jdTT)

		// Ecliptic coordinates
		eclLat, eclLon := coord.ICRFToEcliptic(pos[0], pos[1], pos[2])

		// RA/Dec: convert ICRF position to RA/Dec
		r := math.Sqrt(pos[0]*pos[0] + pos[1]*pos[1] + pos[2]*pos[2])
		x, y, z := pos[0]/r, pos[1]/r, pos[2]/r
		decDeg := math.Asin(z) * 180 / math.Pi
		raHours := math.Atan2(y, x) * 12 / math.Pi
		if raHours < 0 {
			raHours += 24
		}

		// Format RA as hours:minutes:seconds
		ra := units.AngleFromHours(raHours)
		_, raH, raM, raS := ra.HMS()

		// Format Dec as degrees:arcmin:arcsec
		dec := units.AngleFromDegrees(decDeg)
		decSign, decD, decAM, decAS := dec.DMS()
		decSignStr := "+"
		if decSign < 0 {
			decSignStr = "-"
		}

		fmt.Printf("%s:\n", b.name)
		fmt.Printf("  RA:  %02dh %02dm %05.2fs\n", raH, raM, raS)
		fmt.Printf("  Dec: %s%02d° %02d' %05.2f\"\n", decSignStr, decD, decAM, decAS)
		fmt.Printf("  Ecl: lat=%.4f° lon=%.4f°\n\n", eclLat, eclLon)
	}
}

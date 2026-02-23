// Example: Identify which constellation a planet is in
//
// Demonstrates looking up the IAU constellation for planet positions.
// Uses J2000 RA/Dec coordinates, which are close enough to the B1875
// epoch used by constellation boundaries for practical purposes.
package main

import (
	"fmt"
	"math"
	"time"

	"github.com/anupshinde/goeph/constellation"
	"github.com/anupshinde/goeph/spk"
	"github.com/anupshinde/goeph/timescale"
)

func main() {
	eph, err := spk.Open("data/de440s.bsp")
	if err != nil {
		panic(err)
	}

	t := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	jdTT := timescale.UTCToTT(timescale.TimeToJDUTC(t))

	fmt.Printf("Planet constellations on %s:\n\n", t.Format("2006-01-02"))
	fmt.Printf("%-10s %8s %8s   %-4s %s\n", "Body", "RA (h)", "Dec (°)", "Abbr", "Constellation")
	fmt.Println("---------- -------- --------   ---- --------------------")

	bodies := []struct {
		name string
		id   int
	}{
		{"Sun", spk.Sun},
		{"Moon", spk.Moon},
		{"Mercury", spk.MercuryBarycenter},
		{"Venus", spk.VenusBarycenter},
		{"Mars", spk.MarsBarycenter},
		{"Jupiter", spk.JupiterBarycenter},
		{"Saturn", spk.SaturnBarycenter},
	}

	for _, body := range bodies {
		pos := eph.Observe(body.id, jdTT)
		ra, dec := icrfToRADec(pos)
		abbr := constellation.At(ra, dec)
		name := constellation.Name(abbr)
		fmt.Printf("%-10s %8.4f %+8.4f   %-4s %s\n", body.name, ra, dec, abbr, name)
	}

	// Some famous stars
	fmt.Println("\nFamous stars:")
	fmt.Printf("%-15s %8s %8s   %-4s %s\n", "Star", "RA (h)", "Dec (°)", "Abbr", "Constellation")
	fmt.Println("--------------- -------- --------   ---- --------------------")

	stars := []struct {
		name    string
		ra, dec float64
	}{
		{"Sirius", 6.75, -16.72},
		{"Betelgeuse", 5.92, 7.41},
		{"Vega", 18.62, 38.78},
		{"Polaris", 2.53, 89.26},
	}

	for _, s := range stars {
		abbr := constellation.At(s.ra, s.dec)
		name := constellation.Name(abbr)
		fmt.Printf("%-15s %8.4f %+8.4f   %-4s %s\n", s.name, s.ra, s.dec, abbr, name)
	}
}

// icrfToRADec converts an ICRF position vector to RA (hours) and Dec (degrees).
func icrfToRADec(pos [3]float64) (raHours, decDeg float64) {
	r := math.Sqrt(pos[0]*pos[0] + pos[1]*pos[1] + pos[2]*pos[2])
	decDeg = math.Asin(pos[2]/r) * 180.0 / math.Pi
	raHours = math.Atan2(pos[1], pos[0]) * 180.0 / math.Pi / 15.0
	if raHours < 0 {
		raHours += 24.0
	}
	return
}

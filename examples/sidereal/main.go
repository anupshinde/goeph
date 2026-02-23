// Example: Sidereal time and Earth Rotation Angle
//
// Shows GMST (Greenwich Mean Sidereal Time), GAST (Greenwich Apparent
// Sidereal Time), and the modern ERA (Earth Rotation Angle) at a given
// UT1 date. GAST includes nutation; ERA is the IAU 2000 replacement for GMST.
package main

import (
	"fmt"
	"time"

	"github.com/anupshinde/goeph/coord"
	"github.com/anupshinde/goeph/timescale"
)

func main() {
	t := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	jdUTC := timescale.TimeToJDUTC(t)
	jdTT := timescale.UTCToTT(jdUTC)
	jdUT1 := timescale.TTToUT1(jdTT)

	fmt.Printf("Date: %s\n", t.Format("2006-01-02 15:04 UTC"))
	fmt.Printf("JD (UT1): %.6f\n\n", jdUT1)

	gmst := coord.GMST(jdUT1)
	gast := coord.GAST(jdUT1)
	era := coord.EarthRotationAngle(jdUT1)

	fmt.Printf("GMST: %10.6f°  (%s)\n", gmst, degToHMS(gmst))
	fmt.Printf("GAST: %10.6f°  (%s)\n", gast, degToHMS(gast))
	fmt.Printf("ERA:  %10.6f°  (%s)\n\n", era, degToHMS(era))

	// The equation of the equinoxes (nutation correction)
	eqEq := gast - gmst
	fmt.Printf("Equation of equinoxes (GAST-GMST): %.4f\" (arcseconds)\n", eqEq*3600)

	// Show how sidereal time advances over 24 hours
	fmt.Println("\nSidereal time at 6-hour intervals:")
	for h := 0; h <= 24; h += 6 {
		t2 := t.Add(time.Duration(h) * time.Hour)
		jdUT12 := timescale.TTToUT1(timescale.UTCToTT(timescale.TimeToJDUTC(t2)))
		gmst2 := coord.GMST(jdUT12)
		fmt.Printf("  +%02dh: GMST = %s\n", h, degToHMS(gmst2))
	}
}

func degToHMS(deg float64) string {
	for deg < 0 {
		deg += 360
	}
	for deg >= 360 {
		deg -= 360
	}
	hours := deg / 15.0
	h := int(hours)
	m := int((hours - float64(h)) * 60)
	s := (hours - float64(h) - float64(m)/60.0) * 3600
	return fmt.Sprintf("%02dh %02dm %05.2fs", h, m, s)
}

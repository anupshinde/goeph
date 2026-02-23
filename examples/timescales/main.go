// Example: Time scale conversions (UTC -> TT -> UT1, TDB-TT)
//
// Demonstrates the time conversion chain used by the library:
// UTC (civil time) -> TT (uniform atomic-based time for ephemeris)
// -> UT1 (Earth rotation time for sidereal computations).
// Also shows the TDB-TT difference (< 2ms periodic term).
package main

import (
	"fmt"
	"time"

	"github.com/anupshinde/goeph/timescale"
)

func main() {
	// A specific date
	t := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	fmt.Printf("Input: %s\n\n", t.Format("2006-01-02 15:04:05 UTC"))

	// UTC Julian Date
	jdUTC := timescale.TimeToJDUTC(t)
	fmt.Printf("JD (UTC): %.8f\n", jdUTC)

	// Leap second offset: TAI-UTC in seconds
	leapSec := timescale.LeapSecondOffset(jdUTC)
	fmt.Printf("Leap seconds (TAI-UTC): %.0f s\n", leapSec)

	// TT = UTC + leap_seconds + 32.184s
	jdTT := timescale.UTCToTT(jdUTC)
	fmt.Printf("JD (TT):  %.8f\n", jdTT)
	fmt.Printf("  TT - UTC = %.3f s\n", (jdTT-jdUTC)*timescale.SecPerDay)

	// UT1 (Earth rotation time, differs from TT by delta-T)
	jdUT1 := timescale.TTToUT1(jdTT)
	fmt.Printf("JD (UT1): %.8f\n", jdUT1)
	fmt.Printf("  TT - UT1 (delta-T) = %.3f s\n", (jdTT-jdUT1)*timescale.SecPerDay)

	// TDB-TT periodic difference (< 2 ms)
	tdbMinusTT := timescale.TDBMinusTT(jdTT)
	fmt.Printf("  TDB - TT = %.6f s (%.3f ms)\n\n", tdbMinusTT, tdbMinusTT*1000)

	// Delta-T varies over centuries
	fmt.Println("Delta-T through history:")
	years := []float64{1900, 1950, 2000, 2020, 2050, 2100}
	for _, y := range years {
		dt := timescale.DeltaT(y)
		fmt.Printf("  Year %.0f: delta-T = %7.2f s\n", y, dt)
	}
}

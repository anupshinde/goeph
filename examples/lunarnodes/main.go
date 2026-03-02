// Example: True and mean lunar node positions
//
// The lunar nodes are where the Moon's orbit crosses the ecliptic plane.
// The ascending (north) node regresses westward with an 18.6-year period.
// Eclipses can only occur when the Sun is near a lunar node.
//
// The true node is computed from the Moon's actual position and velocity
// in the JPL ephemeris, capturing all perturbations. The mean node is a
// smooth analytical approximation (Meeus polynomial) that ignores
// short-period oscillations of up to ~2 degrees.
package main

import (
	"fmt"
	"time"

	"github.com/anupshinde/goeph/lunarnodes"
	"github.com/anupshinde/goeph/spk"
	"github.com/anupshinde/goeph/timescale"
)

func main() {
	eph, err := spk.Open("data/de440s.bsp")
	if err != nil {
		panic(err)
	}

	// Compare true vs mean node positions over a few years
	fmt.Println("Lunar node positions (ecliptic longitude, north node):")
	fmt.Printf("%-12s %12s %12s %10s\n", "Date", "True Node", "Mean Node", "Diff")
	fmt.Println("------------ ------------ ------------ ----------")

	for year := 2024; year <= 2028; year++ {
		t := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		jdTT := timescale.UTCToTT(timescale.TimeToJDUTC(t))

		trueN, _ := lunarnodes.TrueNode(eph, jdTT)
		meanN, _ := lunarnodes.MeanLunarNodes(jdTT)

		diff := trueN - meanN
		if diff > 180 {
			diff -= 360
		}
		if diff < -180 {
			diff += 360
		}

		fmt.Printf("%d-01-01   %9.2f°   %9.2f°   %+6.2f°\n", year, trueN, meanN, diff)
	}

	fmt.Println("\nThe north node moves ~19.3° per year (retrograde),")
	fmt.Println("completing one full cycle in ~18.6 years.")
	fmt.Println("The true node oscillates around the mean by up to ~2°.")

	// Show monthly true node positions for 2024
	fmt.Println("\nMonthly true north node position for 2024:")
	for month := time.January; month <= time.December; month++ {
		t := time.Date(2024, month, 1, 0, 0, 0, 0, time.UTC)
		jdTT := timescale.UTCToTT(timescale.TimeToJDUTC(t))
		trueN, _ := lunarnodes.TrueNode(eph, jdTT)
		fmt.Printf("  %s: %.2f°\n", t.Format("Jan"), trueN)
	}
}
